---
title: "Error Boundary と Suspense の配置で決まる、
  コンポーネントの開発体験とユーザー体験"
emoji: "😊"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [React]
published: false
---

## TODO：消せ

タイトルの案

- "Error Boundary と Suspense の配置で決まる、
  コンポーネントの開発体験とユーザー体験"

- "エラー境界の局所化が生む 3 つの価値 - テスト・並列開発・UX 向上を実現する React 設計パターン ―"

- "エラーとローディングの局所化、テスト、コンポーネントカタログを意識した React コンポーネント設計
  ― Dummy API を使って Vitest と Storybook を実践しながら、ErrorBoundary / Suspense の配置を学ぶ ―"

## はじめに

この記事では、**Error Boundary と Suspense の配置**という、
一見地味だが重要な設計判断が、

- テストの書きやすさ
- 並列開発の可能性
- ユーザー体験

これら 3 つにどう影響するかを、実装例を通じて紹介します。

「全体を 1 つの Error Boundary でラップする」という
よくある実装（CheapShop）と、
「コンテンツ層で個別にラップする」実装（Shop）を
比較しながら、適切な配置の自分なりの原則について説明します。

### 対象読者

Error Boundary/Suspense を使ったことがあるが、それらを用いた設計に悩んでいる方に特に読んでいただきたいです！
特に下記のような事象についてこの記事が解決の糸口になればと思います。

- Error Boundary と Suspense を使っているが、どこに配置すべきか迷う
- 1 つのコンポーネントがエラーになると画面全体がエラーになる
- コンポーネント単体でローディング状態やエラー状態をテストしたい
- Storybook で個別コンポーネントのローディングやエラー表示を確認したい

### 主に利用したライブラリとそのバージョン

主に利用したライブラリとそのバージョンはこちらです。

| ライブラリ                  | バージョン |
| --------------------------- | ---------- |
| React                       | 19.1.1     |
| Vite                        | 7.1.6      |
| Vitest                      | 3.2.4      |
| Storybook                   | 9.1.7      |
| MSW                         | 2.11.2     |
| @tanstack/react-query       | 5.90.2     |
| @testing-library/react      | 16.3.0     |
| @testing-library/jest-dom   | 6.8.0      |
| @testing-library/user-event | 14.6.1     |
| zod                         | 4.1.11     |

なお、ダミーの API として[DummyJSON](https://dummyjson.com/docs)を利用しています。

TODO:ブランチも書く

## 題材

下記の画像のような下記の 3 つで構成された EC サイトを題材とします。

- 商品カート
- 商品一覧
- 今日の名言

それぞれのセクションでは異なる API をコールして、カート・商品・今日の名言を取得しています。

:::message
今日の名言のセクションは社長がどうしてもと要望されて入れたもので、EC サイトの構成要素としては全く不要なものとします。
:::

TODO:画像の余白を小さくする
TODO:コンポーネントの名前を画像に含める

![完成形のコード構成](/images/202510XX_article-error-boundary-suspense-msw-storybook-test/image.png)

## よくない ErrorBoundary と Suspense の使い方と問題点（自分の実例）

### よくない例

Error Boundary や Suspense を知った当初は便利だなーと思いつつ、使い方がよくわかっていませんでした。
具体的には下記のように、コンポーネントのトップの要素に**のみ**`<ErrorBoundary>`と`<Suspense>` をラップする形をよく採用していました。
（`<Cart>`などのコンポーネントには含まれていない）

```tsx
export const CheapShop = () => {
  return (
    <ErrorBoundary fallback={<div>全画面エラーが発生しました</div>}>
      <Suspense fallback={<div>全画面読み込み中...</div>}>
        <div>
          <h1>Super coolなECサイト</h1>
          <Cart />
          <ProductList />
          <Quote />
        </div>
      </Suspense>
    </ErrorBoundary>
  );
};
```

### デメリット

さきほどのよくない例であげたようなコンポーネント設計にすると下記のような 3 つのデメリットがあります。
下記のあげるデメリットが発生する原因はいずれも、エラーをキャッチする Error Boundary、Promise を受け取る Suspense がルート要素にしかないために発生するものです。

#### 1. どれか 1 つの API でも通信中なら全画面ローディングになる

すべての API の通信が完了するまで、画面全体が「読み込み中...」と表示されます。
例えば、商品一覧とカートのデータ取得は完了していても、
今日の名言の API だけが遅い場合、ユーザーは何も操作できません。

TODO:最新の Chromatic の URL を貼る
[全画面ローディングの Story]()

#### 2. 1 つのエラーで全機能が使用不可になる

優先度の低い「今日の名言」の API がエラーになっただけで、
商品一覧もカートも含めた画面全体が「エラーが発生しました」になります。

TODO:最新の Chromatic の URL を貼る
[全画面エラーの Story]()

ユーザーから見ると下記のような不便さがあります。

- 本日の名言でエラーが発生すると、本来使えるはずの機能（商品閲覧・カート）が使えない
- どこでエラーが起きたのか分からない

#### 3. コンポーネント単体のテスト・Storybook が書きにくい

`<Cart />`コンポーネント単体でテストを書く場合、
`<Cart />`コンポーネント親コンポーネントにあった`<ErrorBoundary>`と`<Suspense>`がないため、ローディングやエラーのテストや Storybook が書きづらくなります。

例えば下記のようなカートを表示するコンポーネントの単体テストを考えます。
なお、useItem の内部では`TanStack Query`の`useSuspenseQuery`を使ってカートの情報を取得しています。

```tsx
export const Content = () => {
  const { cart } = useItem({ userId: "1" });

  if (cart.products.length === 0) {
    return <div>カートには何もありません。</div>;
  }

  return (
    <div>
      <div>カートの商品の金額:{cart.total}円</div>
      <ul>
        {cart.products.map((product) => (
          <li key={product.id}>{product.title}</li>
        ))}
      </ul>
    </div>
  );
};
```

このコンポーネントに商品が表示されることのテストを書こうとすると、Cart 内部で Promise をキャッチできる Suspense がないためエラーになります。

```tsx
import { render, screen } from "@testing-library/react";

it("カートに商品が表示されること", () => {
  server.use(/** レスポンスをモック */);

  render(<Cart />);

  expect(screen.getByText("商品1")).toBeInTheDocument();
});
```

## 改善案：ErrorBoundary / Suspense の局所化設計

### コンポーネント設計の方針

カート、商品一覧、今日の名言でそれぞれ共通する書き方はこちらの通りです。

API をコールするコンポーネントの親コンポーネントに`<ErrorBoundary>`と`<Suspense>`をラップさせます。
こうすることでエラーとローディングの範囲を限定することができます。

TODO:ErrorBoundary の具体的なコードは最後に載せる

```tsx
export const Content: FC<Props> = (props) => (
  <ErrorBoundary fallback={<Error />}>
    <Suspense fallback={<Loading />}>
      <Inner {...props} /> //下のコンポーネントで定義
    </Suspense>
  </ErrorBoundary>
);

const Content: FC<Props> = (props) => {
  //APIをコールしてデータを取得
  const { data } = useContent();

  return (
    <div>
      <div>{data.id}</div>
      <div>{data.name}</div>
    </div>
  );
};

const Error = () => <div>データの取得に失敗しました。</div>;

const Loading = () => <div>データを取得中です。</div>;
```

今回は簡単のためにエラーとローディングの時はテキスト表示のみにしています。

- `<Error/>`コンポーネントには再読み込みボタン
- `<Loading/> `コンポーネントにはスケルトン

を使うといった工夫をしても良いかもしれません。

### メリット

改善案の設計にすることで、以下の 3 つの大きなメリットが得られます。

#### 1. ユーザー体験の向上

**部分的なローディング・エラー表示が可能**

先ほどの「よくない例」では、1 つの API が遅い、または失敗すると画面全体が使えなくなりました。
改善案では、各セクションが独立しているため：

- EC サイトのユーザーにとっては優先度の低い今日の名言の API が失敗しても、EC サイトの主要機能（商品閲覧・カート）は正常に動作する
- 商品一覧が表示されていれば、カートの読み込みが遅くてもユーザーは商品を閲覧できる
- エラーが発生した箇所が明確にわかる

**優先度に応じた体験設計**

セクションごとに fallback を最適化できます：

- 重要な商品一覧：詳細なスケルトンと丁寧なエラーメッセージ＋リトライボタン
- 優先度の低い今日の名言：シンプルなローディング表示と、エラー時は非表示で OK

#### 2. テストと Storybook の書きやすさ

**コンポーネント単体でのテストが可能**

各コンポーネントが`<ErrorBoundary>`と`<Suspense>`を内包しているため、
単体テストで以下の状態を簡単にテストできます：

```tsx
describe("Cart", () => {
  it("正常時：カートの商品が表示される", () => {
    server.use(/* 成功レスポンス */);
    render(<Cart />);
    expect(screen.getByText("商品1")).toBeInTheDocument();
  });

  it("ローディング時：ローディング表示が出る", () => {
    server.use(/* 遅延レスポンス */);
    render(<Cart />);
    expect(screen.getByText("データを取得中です。")).toBeInTheDocument();
  });

  it("エラー時：エラーメッセージが表示される", () => {
    server.use(/* エラーレスポンス */);
    render(<Cart />);
    expect(
      screen.getByText("データの取得に失敗しました。")
    ).toBeInTheDocument();
  });
});
```

Storybook での状態確認が容易

各コンポーネントのストーリーで、成功・ローディング・エラーの 3 状態を簡単に再現できます。
MSW のハンドラーを切り替えるだけで、デザイナーや PM も各状態の見た目を確認できます。

#### 3. 並列開発が可能

エラーやローディングの範囲が各コンポーネントに限定されているため、チームメンバーが独立して作業をすることができます。

各セクション（カート、商品一覧、今日の名言）が独立しているため：

- A さんはカート機能を実装
- B さんは商品一覧を実装
- C さんは今日の名言を実装

このように、同時並行で開発を進められます。
Storybook を使えば、エラーやローディングを含めを個別にコンポーネント開発・確認しながら、最終的に統合できます。

## 商品の一覧コンポーネントの実装

### MSW のハンドラを生成するユーティリティ

### コンポーネント

### テスト

- API モック方法（MSW または fetch のスタブ）
- 成功 / エラー / ローディングケースを局所化してテスト
- コンポーネント単位のテスト設計
- Storybook とテストの役割分担

### Storybook

## カートと今日の名言のコンポーネントの実装

商品の一覧コンポーネントと同様の流れで行うので簡単に書く。

### カート

### 今日の名言

## まとめ

- エラーとローディングを局所化する設計の利点
- テストとコンポーネントカタログを意識することで得られる開発体験
- 本番 API へ置き換える際にも活きる設計指針

```

```
