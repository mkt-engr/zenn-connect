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

#### どれか 1 つの API でも通信中なら全画面ローディングになる

すべての API の通信が完了するまで、画面全体が「読み込み中...」と表示されます。
例えば、商品一覧とカートのデータ取得は完了していても、
今日の名言の API だけが遅い場合、ユーザーは何も操作できません。

TODO:最新の Chromatic の URL を貼る
[全画面ローディングの Story]()

#### 1 つのエラーで全機能が使用不可になる

優先度の低い「今日の名言」の API がエラーになっただけで、
商品一覧もカートも含めた画面全体が「エラーが発生しました」になります。

TODO:最新の Chromatic の URL を貼る
[全画面エラーの Story]()

ユーザーから見ると下記のような不便さがあります。

- 本日の名言でエラーが発生すると、本来使えるはずの機能（商品閲覧・カート）が使えない
- どこでエラーが起きたのか分からない

#### コンポーネント単体のテスト・Storybook が書きにくい

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
  //カート取得のAPIのレスポンスをモック

  render(<Cart />);
  expect(screen.getByText("商品1")).toBeInTheDocument();
});
```

## 改善案：ErrorBoundary / Suspense の局所化設計

### どんな設計か

- 各セクションを独立させる設計
- フォールバックをセクションごとに最適化
- 再試行の粒度を細かく制御できる
- ユーザー体験の向上とデバッグ容易性

### メリット

- テストでローディングやエラー時の挙動も確認できる
- Storybook についても同様
- 細かく分割することで並列で作業をすることができる

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
