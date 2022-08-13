---
title: "画面遷移時に不要なAPI通信を中止してパフォーマンスを向上させる"
emoji: "👺"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: ["React", "TypeScript", "JavaScript", "Axios"]
published: true
---

# 概要

:::message
この記事は PC で見た方がよいかもしれません。2 種類のソースの動作の違いをわかりやすくするためにブラウザの機能を使って通信速度を遅くするためです。
:::
こんにちは。まっきんとっしゅです([@mkt_phys](https://twitter.com/mkt_phys))。今回は画面遷移した際に API 通信を行い、特定のユーザーの情報を取得しブラウザに表示するシュチュエーションを考えます。4 つのボタンがありそれぞれのボタンをクリックするとそれぞれ異なるユーザーが表示されます。
![](/images/abort-fetch-with-axios/example.png)

この時、最初は 1 をクリックのユーザー情報の表示を待っていた時に「やっぱりユーザー 2 の情報が欲しいから 2 をクリックしよう」と思った場合に表示する必要のないユーザー 1 のレスポンスは不要ですよね。このような不要なレスポンスを返す通信を中止する方法について記事にしました。

記事のタイトルの「パフォーマンスが上がる」というのは「不要な通信を中止し、不要なレンダリングを避けることができる」という意味で使っています。

後ほどコードはお見せしますが私の好きな React で書いています。ですが考え方自体はプレーンな JavaScript や Vue でも生かせると思います。

なお、主に利用したライブラリのバージョン情報は以下の通りです。

| ライブラリ       | バージョン |
| ---------------- | ---------- |
| React            | 17.0.0     |
| React-Router-DOM | 6.0.0      |
| Axios            | 0.27.2     |

# 事前準備

実際にコードを動かす前に不要なリクエストを打ち切った場合とそうでない場合の違いをわかりやすくするためにブラウザの開発者ツールから通信速度の調整をします。

ネットワークタブを開いて以下の画像のように No throttling をクリックし Slow 3G を選択してください。こうすることで通信速度が遅くなり、不要なレンダリングが発生しないことがわかりやすくなります。

![](/images/abort-fetch-with-axios/network1.png)
![](/images/abort-fetch-with-axios/network2.png)

# リクエストを中止しない場合とする場合のレンダリングの違い

実際に実装したものはこちらです。

- リクエストを**中止しない**場合の挙動を見たい時：Not Cancel Fetch をクリック
- リクエストを**中止する**場合の挙動を見たい時：Cancel Fetch Fetch をクリック

@[codesandbox](https://codesandbox.io/embed/apitong-xin-wozhong-duan-suru-fz68vf?fontsize=14&hidenavigation=1&theme=dark)

`/1`の URL にアクセスした場合、ユーザー ID が 1 のものを取得しブラウザに表示しています。4 つのボタンがありますが、それぞれのボタンの数字が URL 対応しています。ユーザー ID が 3 のユーザーの情報が欲しい場合は 3 をクリックすると言った具合です。

## 通信を中止しない場合の実装

`useEffect`内で Axios を用いて通信を行い、ユーザー情報を取得します。取得したユーザーのうち ID、ユーザー名、メールアドレスを表示します。`useEffect`の第 2 引数に id を用いているのは画面遷移をした場合にのみ通信を行いたいからです。

```ts:NotCancelFetch.tsx
const NotCancelFetch: FC = () => {
  //URLのパラメタを取得
  const { id } = useParams();

  const [user, setUser] = useState<User | null>();
  useEffect(() => {
    //APIのURLとリクエストメソッド
    const options: AxiosRequestConfig = {
      url: `${url}/users/${id}`, //url:JSON Placeholder
      method: "GET",
    };

    axios(options).then((res: AxiosResponse<User>) => {
      setUser(res.data);
    })
    .catch((e) => {
      console.log(e.message);
    });
    //画面遷移する度(URLのパラメタのidが変わる度)に通信を行う
  }, [id]);

  return (
    <div>
      // リンクなどは省略
      <ul>
        <li>ID : {user?.id}</li>
        <li>ユーザー：{user?.name}</li>
        <li>メールアドレス：{user?.email}</li>
      </ul>
    </div>
  );
};
```

特段変わったところもないよくある実装だと思います（Axios の型付けに関しては私が書いた[こちら](https://zenn.dev/mkt_engr/articles/axios-req-res-typescript)の記事を参照してください笑）。事前準備でお話しした通り通信速度を遅くしてみると次のような挙動になります。
![リクエストを中止しない場合の挙動](/images/abort-fetch-with-axios/notCancelFetch.gif)

初めに ID が 1 のユーザーの情報が表示されているとします。まず 2 をクリックして ID が 2 のユーザーの情報が表示される前に 3,4 とクリックすると最終的には必要のないユーザーの情報（ID が 2,3）が一瞬レンダリングレンダリングされて画面がちらついているのがわかると思います。

今回はこのムダな通信を中止しレンダリングさせない実装を行います。

## 通信を中止する場合の実装

リクエストを注視する場合にキモとなってくるのは`AbortController`です（ちなみに英単語の abort は「中止する」や「中断する」といった意味があります）。この`AbortController`と`Axios`を紐づけることで画面遷移する際に通信を中断できます。具体的な実装は以下の通りで、通信を中止しない時と比べた場合のソースからの変更点は 3 箇所です。

```diff ts:CancelFetch.tsx
const CancelFetch: FC = () => {
  const { id } = useParams();
  const [user, setUser] = useState<User | null>();
  useEffect(() => {
+   const controller = new AbortController();
    const options: AxiosRequestConfig = {
      url: `${url}/users/${id}`,
      method: "GET",
+     signal: controller.signal, //AbortControllerとAxiosの紐付け
    };

    axios(options)
      .then((res: AxiosResponse<User>) => {
        setUser(res.data);
      })
      .catch((e) => {
        console.log(e.message);
      });
+    return () => {
+      //画面遷移する時(アンマウントする時)に通信を中止する
+      controller.abort();
+    };
  }, [id]);

  return (
    <div>
      // リンクなどは省略
      <ul>
        <li>ID : {user?.id}</li>
        <li>ユーザー：{user?.name}</li>
        <li>メールアドレス：{user?.email}</li>
      </ul>
    </div>
  );
};
```

事前準備でお話しした通り通信速度を遅くしてみると次のような挙動になります。
![通信を中止した場合の挙動](/images/abort-fetch-with-axios/cancelFetch.gif)

初めに ID が 1 のユーザーが表示されていてその後 2,3,4 とクリックするユーザーの動きは通信を中止しない場合と変えていません。しかし 2,3 の通信は行われておらずレンダリングもされていません。下記の画像のように 2,3 の通信が中止されていることがわかります。(UX を考えるとローディング画面が必要ですが付けてません。すみません。)

![通信を中止した場合のネットワークの様子](/images/abort-fetch-with-axios/cancelFetchNetwork.png)

最後に 3 つの変更点について説明します。

1. AbortController を new する
2. AbortController と Axios の紐付けを行う
3. クリーンアップ時の処理に通信を中止する処理を行う

1 と 2 はセットなのでまとめて書くと

```ts
const controller = new AbortController();
const options: AxiosRequestConfig = {
  url: `${url}/users/${id}`,
  method: "GET",
  signal: controller.signal,
};
```

となります。AbortController を new して Axios の設定に追加するだけです。AbortController 自体は React の特有のものではなくプレーンな JavaScript のインターフェースです。

最後に 3 の変更ですが画面遷移時（アンマウント時）に通信を中断したいのでクリーンアップ時の処理を追加します。

```ts
useEffect(() => {
  //通信などの処理は省略

  return () => {
    controller.abort();
  };
}, [id]);
```

# まとめ

- 通信を中止したいときは AbortController を使う
- 画面遷移（アンマウント）する時に不要な通信を中止することでパフォーマンスの向上にもつながる

# 参考

[Axios の公式ドキュメント（Axios Cancellation）](https://axios-http.com/docs/cancellation)
[MDN（AbortController）](https://developer.mozilla.org/ja/docs/Web/API/AbortController)
[All useEffect Mistakes Every Junior React Developer Makes](https://www.youtube.com/watch?v=QQYeipc_cik)
ちなみにこの YouTube の動画は v0.22.0 以降の Axios では非推奨の CancelToken を使った方法も説明しています。
