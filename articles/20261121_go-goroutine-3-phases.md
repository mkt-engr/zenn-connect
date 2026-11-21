---
title: "Goのgoroutineでワーカーを安全に終了する方法"
emoji: "🐹"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Go, goroutine, 並行処理]
published: false
---

## はじめに

Goのgoroutineは、少ない記述で並行処理を始められる便利な仕組みです。一方で、`go` を付けて関数を起動するだけでは、終了管理や同時実行数、処理失敗時の扱いまでは考えられていません。

この記事では、goroutineを使った処理を次の3段階で整理します。

1. goroutineの起動と終了を管理する
2. Worker Poolで同時実行数を制御し、Graceful Shutdownを実現する
3. 非同期メッセージ処理で、再配信や二重処理、データ整合性を考える

段階を追って、単に「並行して動かす」状態から「運用を想定して制御する」状態へ改善していきます。

## goroutineとは

goroutineは、Goのランタイムが管理する軽量な実行単位です。関数を呼び出すときに `go` を付けると、その関数を別の流れで実行できます。

通常の関数呼び出しでは、呼び出した関数が終わるまで次の行に進みません。`go` を付けると、呼び出し元はすぐ次の行に進み、関数は裏で並行して動きます。

```go
func task(name string) {
    // 時間のかかる処理
}

func main() {
    task("A")    // Aが終わるまでBは始まらない
    task("B")

    go task("A") // Aは裏で動き、mainはすぐ次へ進む
    go task("B") // AとBが並行して動く
}
```

goroutineは軽量なので、数百〜数千個を同時に起動することも一般的です。一方で、`main` 関数が終わるとプログラム全体が終了し、動作中のgoroutineも止まります。そのため、goroutineを起動したあとに「終わるまで待つ」「止める」という管理が必要になります。この記事の後半では、その管理方法を段階的に扱います。

### goroutineが役立つ場面

- 複数のHTTPリクエストを同時に送り、すべての結果をまとめて受け取る
- 画像変換やファイル処理など、時間のかかる処理を並行して進める
- キューから届いた仕事を、裏のWorkerで順番に処理する（Phase 2・Phase 3で扱う内容）
- 処理にタイムアウトやキャンセルを設け、途中で止められるようにする（Phase 1のcontextで扱う内容）

### 対象読者

- `go` でgoroutineを起動したことはあるが、終了管理や停止処理に自信がない方
- Worker PoolやGraceful Shutdownの考え方を整理したい方
- Pub/Subなどの非同期メッセージ処理で、ACKや冪等性をどう考えるか知りたい方

### 主に利用したライブラリとそのバージョン

この記事のコードは、Goの標準ライブラリのみを使っています。

| ライブラリ | バージョン |
| --- | --- |
| Go | 1.27.4 |
| sync（標準ライブラリ） | Goに同梱 |
| context（標準ライブラリ） | Goに同梱 |
| time（標準ライブラリ） | Goに同梱 |

Pub/Subなどのメッセージングサービスの具体的なSDKは使わず、ACK/NACKは概念的な例として示しています。

## Phase 1：goroutineを起動するだけでは不十分

### まずはgoroutineを起動する

goroutineは、関数呼び出しの前に `go` を付けることで起動できます。

```go
func worker() {
    fmt.Println("working")
}

func main() {
    go worker()
}
```

ただし、このコードには問題があります。`main`
関数が終了するとプログラム自体が終了するため、起動したgoroutineの処理が完了する保証はありません。

goroutineを起動することと、その終了を管理することは別です。

### WaitGroupで終了を待つ

複数のgoroutineの終了を待つには、`sync.WaitGroup` を使えます。

```go
func worker(wg *sync.WaitGroup) {
    defer wg.Done()

    fmt.Println("working")
}

func main() {
    var wg sync.WaitGroup

    wg.Add(1)
    go worker(&wg)

    wg.Wait()
}
```

`Add` で待つgoroutineの数を登録し、goroutineの処理が終わったら `Done`
を呼びます。`Wait` は登録された処理がすべて終わるまで待機します。

### channelで仕事を受け渡す

goroutine間でデータを受け渡すにはchannelを使います。

```go
func worker(jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()

    for job := range jobs {
        fmt.Println("processing:", job)
    }
}

func main() {
    jobs := make(chan int)

    var wg sync.WaitGroup
    wg.Add(1)
    go worker(jobs, &wg)

    jobs <- 1
    jobs <- 2

    close(jobs)
    wg.Wait()
}
```

`close(jobs)`
によって、これ以上仕事が送られないことをWorkerに伝えます。`for range`
はchannelが閉じられ、バッファ内の値もなくなると終了します。

### contextでキャンセルを伝える

`WaitGroup`
は終了を待つ仕組みですが、処理を止める指示を出すものではありません。キャンセルを伝えるには
`context` を使います。

```go
func worker(ctx context.Context, jobs <-chan int) {
    for {
        select {
        case <-ctx.Done():
            fmt.Println("worker stopped:", ctx.Err())
            return
        case job := <-jobs:
            fmt.Println("processing:", job)
        }
    }
}
```

`select` で、キャンセル通知（`ctx.Done()`）と仕事の受信（`jobs`）を同時に待ちます。`default` を置いて何もせずループすると、CPUを消費し続けてしまうため注意が必要です。

呼び出し側は `context.WithCancel`
でキャンセル可能なContextを作り、`cancel()` で終了を通知できます。

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

jobs := make(chan int)

var wg sync.WaitGroup
wg.Add(1)

go func() {
    defer wg.Done()
    worker(ctx, jobs)
}()

jobs <- 1
jobs <- 2

// 必要なタイミングで終了を要求
cancel()
wg.Wait()
```

実行結果の例です（`cancel()` の直前に少し待つと、2件とも処理されます）。

```
processing: 1
processing: 2
worker stopped: context canceled
```

### Phase 1のポイント

- goroutineは起動しただけでは終了を管理できない
- `WaitGroup` はgoroutineの終了を待つ
- channelはgoroutine間のデータ受け渡しに使う
- `close` は送信側が「これ以上送らない」と伝える
- `context` はキャンセルや期限を伝える

**改善の考え方：**
goroutineを起動するだけでなく、仕事の受け渡し方と終了条件を設計する。


## Phase 2：Worker PoolとGraceful Shutdown

Phase
1ではgoroutineの起動・終了を学びました。次は、仕事が大量に届く場合を考えます。

### goroutineを無制限に起動する問題

例えば、100件の仕事を次のように処理するとします。

```go
for _, job := range jobs {
    go process(job)
}
```

この方法では仕事の数だけgoroutineが起動します。仕事が大量に届くと、同時実行数が増えすぎて、CPUやメモリ、DB接続などのリソースを圧迫する可能性があります。

そこでWorker Poolを使います。

### Worker Poolで同時実行数を制御する

Worker
Poolでは、決められた数のWorkerを起動し、channel経由で仕事を渡します。

```go
func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()

    for job := range jobs {
        fmt.Printf("worker %d processing job %d\n", id, job)
    }
}

func main() {
    jobs := make(chan int)
    var wg sync.WaitGroup

    workerCount := 3

    for i := 0; i < workerCount; i++ {
        wg.Add(1)
        go worker(i, jobs, &wg)
    }

    for job := 1; job <= 10; job++ {
        jobs <- job
    }

    close(jobs)
    wg.Wait()
}
```

この例ではWorkerを3つだけ起動しています。仕事が10件あっても、同時に仕事を処理するWorkerは最大3つです。

### Graceful Shutdownとは

サーバーやWorkerを停止するとき、いきなりプロセスを終了すると処理中の仕事が中断される可能性があります。

Graceful
Shutdownは、新しい仕事の受付を止め、処理中の仕事を可能な範囲で完了させてから終了する考え方です。

基本的な流れは次のとおりです。

1. シャットダウン要求を受け取る
2. 新しい仕事の受付を止める
3. Workerに終了を伝える
4. 処理中の仕事が終わるのを待つ
5. 制限時間を超えた場合はタイムアウトとして扱う

### context.WithTimeoutで終了期限を設ける

終了を無期限に待つのではなく、期限を設けることも重要です。

```go
shutdownCtx, cancel := context.WithTimeout(
    context.Background(),
    10*time.Second,
)
defer cancel()
```

例えば、Workerの終了を待つ処理を別goroutineで実行し、終了通知とタイムアウトを
`select` で待つことができます。

```go
done := make(chan struct{})

go func() {
    wg.Wait()
    close(done)
}()

select {
case <-done:
    fmt.Println("all workers stopped")
case <-shutdownCtx.Done():
    fmt.Println("shutdown timeout:", shutdownCtx.Err())
}
```

ここで注意したいのは、タイムアウトしたからといって、処理中のgoroutineが自動的に強制終了されるわけではないことです。各Workerや処理関数がContextを確認し、キャンセルに協力する設計が必要です。

### 処理中の仕事をどう扱うか

Graceful Shutdownでは、単に `cancel()` を呼べばよいわけではありません。

例えば、次のような方針を決める必要があります。

- すでに開始した仕事は完了まで待つ
- まだ開始していない仕事は破棄する
- 未処理の仕事をキューや永続ストレージに戻す
- 外部API呼び出しなど、途中でキャンセルできる処理はContextで止める

どの方針が適切かは、仕事の性質や再実行可能性によって変わります。

### Phase 2のポイント

- Worker Poolで同時実行数を制御する
- `WaitGroup` でWorkerの終了を待つ
- channelを閉じて新しい仕事がないことを伝える
- `context.WithCancel` でキャンセルを伝える
- `context.WithTimeout` で終了を待つ期限を設ける
- タイムアウトはgoroutineを強制終了する機能ではない
- 処理中・未処理の仕事をどう扱うか決めておく

**改善の考え方：**
goroutineの数を制御し、停止時にも仕事を安全に扱えるようにする。


## Phase 3：非同期メッセージ処理とデータ整合性

Phase
2までは、主にアプリケーション内のgoroutineやWorkerの制御を扱いました。実務では、Pub/Subなどのメッセージングサービスから仕事を受け取って非同期処理するケースもあります。

ここでは、メッセージを受信して処理するWorkerを考えます。

### メッセージ処理の基本的な流れ

例えば、注文イベントを受信して後続処理を行う場合、概念的には次の流れになります。

1. メッセージを受信する
2. メッセージの内容を検証する
3. 注文に関する処理を実行する
4. 成功したらACKする
5. 失敗したらNACKやリトライなど、システムの方針に従って扱う

ACKはメッセージ処理が完了したことをブローカーに伝える操作です。NACKは処理できなかったことを伝え、再配信などにつなげるために使われます。実際の挙動は利用するメッセージングサービスや設定によって異なります。

### ACKを早く返しすぎない

例えば、処理をgoroutineに渡した直後にACKしてしまうと、実際の処理が終わる前にメッセージが完了扱いになる可能性があります。

```go
func handleMessage(msg Message) {
    go process(msg)
    msg.Ack() // 処理完了前にACKしてしまう可能性がある
}
```

この設計では、処理中にプロセスが停止すると、メッセージが再配信されず、仕事が失われる可能性があります。

基本的には、処理が成功したことを確認してからACKする設計にします。

```go
func handleMessage(msg Message) {
    if err := process(msg); err != nil {
        msg.Nack()
        return
    }

    msg.Ack()
}
```

これは概念例です。実際のSDKではACK/NACKのAPIやエラー処理が異なります。

### 再配信による二重処理を考える

メッセージングシステムでは、処理が失敗したりACKが届かなかったりすると、同じメッセージが再配信されることがあります。

そのため、同じ注文イベントを複数回受け取っても、結果が壊れないようにする必要があります。

この考え方を冪等性（idempotency）といいます。

例えば、注文IDに一意制約を設定し、同じ注文を重複登録できないようにする方法があります。

```sql
CREATE TABLE orders (
    id BIGINT PRIMARY KEY,
    amount INTEGER NOT NULL
);
```

ただし、単に一意制約を設けるだけで、あらゆる副作用が安全になるわけではありません。重複時の扱い、更新処理、外部API呼び出しなども含めて設計する必要があります。

### DBトランザクションと整合性

例えば、注文を登録した後に決済サービスを呼び出す場合、DBへの保存と外部サービスへのリクエストは、通常ひとつのDBトランザクションとしてまとめられません。

そのため、次のような問題が起き得ます。

- DB登録は成功したが、外部API呼び出しに失敗した
- 外部APIは成功したが、応答を受け取る前にタイムアウトした
- リトライによって同じ処理が複数回実行された

こうしたケースでは、冪等性キー、処理状態の管理、トランザクション、Outboxパターンなどを組み合わせて検討します。どの方式を使うかは、求める整合性やシステム構成によって変わります。

### Phase 3のポイント

- 非同期メッセージ処理では、処理完了とACKの順序が重要
- 失敗時のNACKやリトライの挙動を理解する
- メッセージの再配信による二重処理を想定する
- 冪等性を確保し、重複実行でも安全にする
- DBと外部サービスをまたぐ処理では、部分的な失敗を考慮する

**改善の考え方：**
goroutineやWorkerの制御だけでなく、メッセージの再配信、二重実行、障害時のデータ整合性まで設計する。


## 3段階のまとめ

| 段階 | 主な課題 | 学ぶこと |
| --- | --- | --- |
| Phase 1 | goroutineを起動した後の終了管理 | channel、select、WaitGroup、context |
| Phase 2 | 同時実行数と安全な停止 | Worker Pool、Graceful Shutdown、Timeout |
| Phase 3 | 非同期処理の失敗や重複実行 | ACK/NACK、リトライ、冪等性、DB整合性 |

goroutineを使ううえで大切なのは、単に並行して処理を実行することではありません。

- いくつ同時に動かすのか
- いつ、どのように止めるのか
- 処理中の仕事をどう扱うのか
- 失敗した仕事をどう再実行するのか
- 同じ仕事が複数回実行されても安全か

こうした点を順番に考えることで、実務で扱える並行処理に近づけます。

## 補足：この記事で扱った範囲

この記事は、goroutine学習で扱ったテーマを3段階に整理したものです。コード例は考え方を示すための簡略化したものを含み、実際の本番実装ではエラー処理、Contextの伝播、ログ、メトリクス、リソース制限、リトライ方針などを追加で設計する必要があります。

また、race condition、deadlock、goroutine
leakの検出・防止や、テスト・ベンチマークについては、別途理解を深めていく必要があります。
