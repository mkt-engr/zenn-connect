---
title: "Error Boundary/Suspenseはどこに置く？ テスト・Storybook・UXのためのReactコンポーネント設計"
emoji: "😊"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [React, Vitest, Storybook]
published: false
---

## はじめに

Error Boundary と Suspense をどこに置くか悩みませんか？かつての自分はこの設計をうまくできず、テストが書きづらい、1 つのエラーで全体が落ちる といった問題に苦しみました。
本記事では、その設計が

- テストの書きやすさ
- Storybook の書きやすさ
- ユーザー体験（UX）

にどう影響するかを、実装例を通じて紹介します。

「全体を 1 つの Error Boundary でラップ」という私がよくやってしまっていたよくないパターンと、「セクションごとに個別にラップ」するパターンを比較しながら、適切な配置について説明していきます。

### 対象読者

Error Boundary/Suspense を使ったことがあるが、それらを用いた設計に悩んでいる方に特に読んでいただきたいです！
特に下記のような事象についてこの記事が解決の糸口になればと思います。

- Error Boundary と Suspense を使っているが、どこに配置すべきか迷う
- 1 つのコンポーネントがエラーになると画面全体がエラーになる
- コンポーネント単体でローディング状態やエラー状態をテストしたい
- Storybook で個別コンポーネントのローディングやエラー表示を確認したい

### 主に利用したライブラリとそのバージョン

主に利用したライブラリとそのバージョンはこちらです。
Next.js や RSC は使っていないのでご注意ください。

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

実装のコードは[こちら](https://github.com/mkt-engr/react-vite-error-boundary-suspense-msw-storybook-test/tree/article/error-boundary-suspense-demo)のリポジトリで確認できます。

## 題材

下記の 3 つで構成された EC サイトを題材とします。

- 商品カート
- 商品一覧
- 今日の名言

それぞれのセクションでは異なる API をコールして、カート・商品・今日の名言を取得しています。

:::message
今日の名言のセクションは社長がどうしてもと要望されて入れたもので、EC サイトの構成要素としては全く不要なものとします。
:::

![ECサイトの完成系の画像](/images/20251012_article-error-boundary-suspense-msw-storybook-test/total-view.jpeg =450x)

## よくない Error Boundary と Suspense の使い方と問題点（自分の実例）

### よくない例

Error Boundary や Suspense を知った当初は便利だなーと思いつつ、使い方がよくわかっていませんでした。
具体的には下記のように、コンポーネントのトップの要素にのみ`<ErrorBoundary>`と`<Suspense>` をラップする形をよく採用していました。
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

このコンポーネント設計には下記の 3 つのデメリットがあります。
これらが発生する原因はいずれも、エラーをキャッチする Error Boundary、Promise を受け取る Suspense がルート要素にしかないためです。

#### どれか 1 つの API でも通信中なら全画面ローディングになる

すべての API の通信が完了するまで、画面全体が「読み込み中...」と表示されます。
例えば、商品一覧とカートのデータ取得は完了していても、
今日の名言の API だけが遅い場合、ユーザーは何も操作できません。

全画面ローディングの Storybook は[こちら](https://68e2285ed2a65b4c23a54763-idqbuptadi.chromatic.com/?path=/story/features-cheapshop--loading)です。

#### 1 つのエラーで全機能が使用不可になる

優先度の低い「今日の名言」の API がエラーになっただけで、
商品一覧もカートも含めた画面全体が「エラーが発生しました」になります。

全画面エラーの Storybook は[こちら](https://68e2285ed2a65b4c23a54763-idqbuptadi.chromatic.com/?path=/story/features-cheapshop--error)です。

ユーザーの立場では、下記のようなユーザー体験が損なわれる可能性があります。

- 本日の名言でエラーが発生すると、本来使えるはずの機能（商品閲覧・カート）が使えない
- どこでエラーが起きたのかわからない

#### コンポーネント単体のテスト・Storybook が書きにくい

`<Cart />`コンポーネント単体でテストを書く場合、
`<Cart />`コンポーネントの親コンポーネントにあった`<ErrorBoundary>`と`<Suspense>`がないため、ローディングやエラーのテストや Storybook が書きづらくなります。

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
it("カートに商品が表示されること", () => {
  server.use(/** レスポンスをモック */);
  render(<Cart />);
  expect(screen.getByText("商品1")).toBeInTheDocument();
});
```

## ErrorBoundary と Suspense の配置場所の改善案

### コンポーネント設計の方針

カート、商品一覧、今日の名言で共通する実装パターンは以下の通りです。

API をコールするコンポーネントの親コンポーネントに`<ErrorBoundary>`と`<Suspense>`をラップさせます。
こうすることでエラーとローディングの範囲を限定することができます。

なお Error Boundary の実装は[こちら](https://ja.react.dev/reference/react/Component#catching-rendering-errors-with-an-error-boundary)を参考にしています。

```tsx
export const Content: FC<Props> = (props) => (
  <Error Boundary fallback={<Error />}>
    <Suspense fallback={<Loading />}>
      <Inner {...props} /> //下のコンポーネントで定義
    </Suspense>
  </Error Boundary>
);

const Inner: FC<Props> = (props) => {
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
- `<Loading/>`コンポーネントにはスケルトン

を使うといった工夫をしても良いかもしれません。

### メリット

改善案の設計にすることで、以下の 3 つの大きなメリットが得られます。

#### 1. ユーザー体験の向上

先ほどの「よくない例」では、1 つの API が遅い、または失敗すると画面全体が使えなくなりました。
改善案では、各セクションが独立しているため、以下のような利点があります。

- EC サイトのユーザーにとっては優先度の低い今日の名言の API が失敗しても、EC サイトの主要機能（商品閲覧・カート）は正常に動作する
- 商品一覧が表示されていれば、カートの読み込みが遅くてもユーザーは商品を閲覧できる
- エラーが発生した箇所が明確にわかる

また、セクションごとに fallback を最適化できます。

- 重要な商品一覧では、詳細なスケルトンと丁寧なエラーメッセージ＋リトライボタンを表示
- 優先度の低い今日の名言では、シンプルなローディング表示と、エラー時は非表示で OK

#### 2. テストと Storybook の書きやすさ

各コンポーネントが`<ErrorBoundary>`と`<Suspense>`を内包しているため、
単体テストで以下の状態を簡単にテストできます。

```tsx
describe("Cart", () => {
  it("正常時、カートの商品が表示される", () => {
    server.use(/* 成功レスポンス */);
    render(<Cart />);
    expect(screen.getByText("商品1")).toBeInTheDocument();
  });

  it("ローディング時、ローディング表示が出る", () => {
    server.use(/* 遅延レスポンス */);
    render(<Cart />);
    expect(screen.getByText("データを取得中です。")).toBeInTheDocument();
  });

  it("エラー時、エラーメッセージが表示される", () => {
    server.use(/* エラーレスポンス */);
    render(<Cart />);
    expect(
      screen.getByText("データの取得に失敗しました。")
    ).toBeInTheDocument();
  });
});
```

また、各コンポーネントのストーリーで、成功・ローディング・エラーの 3 状態を簡単に再現できます。
MSW のハンドラーを切り替えるだけで、デザイナーや PM も各状態の見た目を確認できます。

#### 3. 並列開発が可能

エラーやローディングの範囲が各コンポーネントに限定されているため、チームメンバーが独立して作業をすることができます。

以下のように並列開発が可能です。

- A さんはカート機能を実装
- B さんは商品一覧を実装
- C さんは今日の名言を実装

## 商品の一覧表示機能の実装

ここからは具体的な実装に移ります。
一番複雑な商品一覧のコンポーネント、テスト、Storybook の実装を行います。
カート、今日の名言のコンポーネントも商品一覧と同様に作成すれば良いため詳細な説明は省略します。
最初に記載したリポジトリも参考にしてください。

### 商品一覧を取得する API をコールする関数

組み込み API の fetch を使います。
ここで作成した関数はカスタムフックに組み込みます。

```ts
export const fetchProducts = async ({
  query,
}: Args): Promise<ProductsSearchResponse> => {
  const response = await fetch(generateApiUrl(`/products/search?q=${query}`));

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  const data = await response.json();
  const result = productsSearchResponseSchema.safeParse(data);

  if (!result.success) {
    throw new Error(`Invalid products data: ${result.error.message}`);
  }

  return result.data;
};
```

正常系 1 つ、異常系 2 つのテストコードを実装しています。

- 正常系
- 異常系
  - HTTP エラーが起きた場合
  - 正常にレスポンスが返ってきたが、スキーマに違反する場合

なお、テストで使用している`generateProductsSearchMock`と`buildGetProductsSearchHandler`については、[備考 - モックの生成関数と MSW ハンドラのビルダー](#モックの生成関数と-msw-ハンドラのビルダー)セクションを参照してください。

```ts
describe("fetchProducts", () => {
  it("正常にProductsSearchResponseを返し、クエリパラメタが正しく含まれる", async () => {
    const onRequestSearchParams = vi.fn();
    const mockData = generateProductsSearchMock({
      products: [
        generateProductInSearchMock({
          id: 10,
          title: "テスト商品",
          description: "テスト用の商品説明",
          price: 1000,
        }),
      ],
      total: 1,
    });

    server.use(
      buildGetProductsSearchHandler.success({
        response: mockData,
        onRequestSearchParams,
      })
    );

    const { products, total } = await fetchProducts({ query: "test" });

    expect(products).toHaveLength(1);
    expect(products[0]).toMatchObject({
      title: "テスト商品",
      description: "テスト用の商品説明",
      price: 1000,
    });
    expect(total).toBe(1);

    // クエリパラメタがq="test"でリクエストされていることを確かめる
    expect(onRequestSearchParams).toBeCalledWith({ q: "test" });
  });

  it("HTTPエラーの場合はエラーを投げる", async () => {
    server.use(buildGetProductsSearchHandler.error({ status: 500 }));

    await expect(fetchProducts({ query: "test" })).rejects.toThrow(
      "HTTP error! status: 500"
    );
  });

  it("不正なデータの場合はエラーを投げる", async () => {
    const invalidData = generateProductsSearchMock({
      products: [
        generateProductInSearchMock({
          id: "invalid" as unknown as number, // numberではなくstring
          title: "テスト商品",
        }),
      ],
    });

    server.use(
      buildGetProductsSearchHandler.success({
        response: invalidData,
      })
    );

    await expect(fetchProducts({ query: "test" })).rejects.toThrow(
      "Invalid products data:"
    );
  });
});
```

### コンポーネント

上部は検索ボックスで、そこに入力された内容に基づいて検索を行い、下部に商品一覧を表示します。

![商品一覧のセクションの画像](/images/20251012_article-error-boundary-suspense-msw-storybook-test/product-list-default.png)

#### 検索ボックスと商品一覧を表示する`<ProductList>`コンポーネント

入力フォームと検索結果の表示を担当します。
実際の API 呼び出しとデータ表示は、次に説明する`<Result>`コンポーネントに委譲しています。

本筋とは関係ないですが、入力するたびに一瞬ローディング画面が表示されるチラつきを防ぐために、React v19 から登場した[`useDeferredValue`](https://ja.react.dev/reference/react/useDeferredValue)を利用しています。
`useDeferredValue`は、UI の更新を遅延させることで、ユーザーの入力がスムーズに見えるようにします。
`query !== deferredQuery`の間は「検索中」と表示し、バックグラウンドで新しい検索結果を取得しています。

```tsx
export const ProductList = () => {
  const [query, setQuery] = useState("");
  const deferredQuery = useDeferredValue(query);

  return (
    <main>
      <h2>商品一覧</h2>
      <label>
        検索
        <input
          type="text"
          name="検索"
          onChange={(e) => setQuery(e.target.value)}
        />
      </label>
      {query !== deferredQuery ? <span>検索中</span> : null}
      <Result query={deferredQuery} />
    </main>
  );
};
```

#### 商品一覧を表示する`<Result>`コンポーネント

`<Result>`コンポーネントは、実際の API 呼び出しとデータ表示を担当します。
`<ProductList>`から受け取った`query`を使って商品を検索します。

この設計のポイントは下記の 3 つです。

1. `<Result>`コンポーネントが`<ErrorBoundary>`と`<Suspense>`を持つことで、このコンポーネント単体でテストや Storybook が作成できます
2. 実際のデータ取得と表示ロジックは`<Inner>`コンポーネントに分離し、責任を明確化しています
3. エラーやローディングの処理を`<Inner>`から分離することで、`<Inner>`は「データをどう表示するか」だけに集中できます

```tsx
type Props = {
  query: string;
};

export const Result: FC<Props> = ({ query }: Props) => {
  return (
    <Error Boundary fallback={<div>商品一覧でエラーが発生しました</div>}>
      <Suspense fallback={<div>商品一覧を読み込み中...</div>}>
        <Inner query={query} />
      </Suspense>
    </Error Boundary>
  );
};

const Inner: FC<Props> = ({ query }) => {
  const { data } = useProducts({ query });

  if (data.products.length === 0) {
    return (
      <div>
        <h3>商品がありませんでした。</h3>
      </div>
    );
  }

  return (
    <div>
      <h3>商品件数:{data.total}件</h3>
      <div>
        {data.products.map((product) => (
          <div key={product.id}>
            <img src={product.thumbnail} alt={product.title} width={100} />
            <h3>{product.title}</h3>
            <p>{product.description}</p>
            <div>{product.price}円</div>
            <hr />
          </div>
        ))}
      </div>
    </div>
  );
};
```

### カスタムフック

商品データを取得する`useProducts`カスタムフックを実装します。

TanStack Query の`useSuspenseQuery`を使っています。
`useSuspenseQuery`は、ローディング中は自動的に親の`<Suspense>`の fallback を表示し、
エラー時は自動的に親の`<ErrorBoundary>`の fallback を表示します。
そのため、コンポーネント側でローディングやエラーの状態を管理する必要がなく、`data`のみを返しています。

```ts
type Args = {
  query: string;
};

export const useProducts = ({ query }: Args) => {
  const { data } = useSuspenseQuery({
    queryKey: ["products", query],
    queryFn: async () => {
      const response = await fetchProducts({ query });
      return response;
    },
  });

  return { data };
};
```

### テスト

`<Result>`コンポーネントの 4 つの状態をテストします。
`<ErrorBoundary>`と`<Suspense>`が内包されているため、このコンポーネント単体でローディングやエラーのテストが可能です。

下記の 4 つのテストを実装します。

- 商品がある場合
- 商品がない場合
- ローディング中
- エラー

#### テストのセットアップ

テストでは`customRender`という独自のレンダー関数を使用しています。
これは、`<QueryClientProvider>`などの必要な Provider でコンポーネントをラップするためのユーティリティです。
詳細は[備考 - customRender と TestProvider](#customrender-と-testprovider)セクションを参照してください。

#### 商品がある場合

商品が 1 件ある場合のテストです。
このテストでは以下を確認します。

- 商品情報が正しく表示される
- 商品件数が正しく表示される
- クエリパラメタ`q`が正しく API リクエストに含まれている

```tsx
describe("Result", () => {
  it("商品が1つある場合、商品一覧と件数が表示されること", async () => {
    const onRequestSearchParams = vi.fn();

    server.use(
      buildGetProductsSearchHandler.success({
        response: generateProductsSearchMock({
          products: [
            generateProductInSearchMock({
              id: 1,
              title: "iPhone 15 Pro",
              description: "最新のApple製スマートフォン",
              category: "スマートフォン",
              price: 159800,
            }),
          ],
          total: 1,
        }),
        onRequestSearchParams,
      })
    );

    customRender(<Result query="iPhone" />);

    expect(await screen.findByText("商品件数:1件")).toBeInTheDocument();
    expect(await screen.findByText("iPhone 15 Pro")).toBeInTheDocument();

    // products/search?q=iPhoneとなっているかを確かめる
    expect(onRequestSearchParams).toBeCalledWith({ q: "iPhone" });
  });
});
```

#### 商品がない場合

商品が 0 件の場合、「商品がありませんでした。」というメッセージが表示されることを確認します。

```tsx
describe("Result", () => {
   it("商品が0の場合、商品がないメッセージが表示されること", async () => {
    server.use(
      buildGetProductsSearchHandler.success({
        response: generateProductsSearchMock({
          products: [],
          total: 0,
        }),
      })
    );

    customRender(<Result query="" />);

    expect(
      await screen.findByText("商品がありませんでした。")
    ).toBeInTheDocument();
});
```

#### ローディング中

API 通信中は`<Suspense>`の fallback が表示されることを確認します。
`buildGetProductsSearchHandler.loading()`を使うことで、MSW でレスポンスを遅延させています。

```tsx
describe("Result", () => {
  it("ローディング中は読み込み中のメッセージが表示されること", async () => {
    server.use(buildGetProductsSearchHandler.loading());

    customRender(<Result query="test" />);

    expect(screen.getByText("商品一覧を読み込み中...")).toBeInTheDocument();
  });
});
```

#### エラー

API でエラーが発生した場合は`<ErrorBoundary>`の fallback が表示されることを確認します。
ここでは 500 エラーを想定していますが、404 やネットワークエラーなど他のエラーでも同様に動作します。

```tsx
describe("Result", () => {
  it("エラー発生時はエラーメッセージが表示される", async () => {
    server.use(buildGetProductsSearchHandler.error({ status: 500 }));

    customRender(<Result query="test" />);

    expect(
      await screen.findByText("商品一覧でエラーが発生しました")
    ).toBeInTheDocument();
  });
});
```

### Storybook

`<Result>`コンポーネントの 4 つの状態を Storybook で確認できるようにします。
MSW を使って API のレスポンスをモックすることで、各状態を簡単に再現できます。

下記の 4 つのストーリーを実装します。

- 商品がある場合（Default）
- 商品がない場合（NoProduct）
- ローディング中（Loading）
- エラー（Error）

#### Storybook の設定

`inline: false`を設定することで、各ストーリーを個別の iframe で実行します。
これにより、TanStack Query のキャッシュの競合を回避し、各ストーリーが独立して動作するようになります。

```tsx
const meta: Meta<typeof component> = {
  tags: ["autodocs"],
  component,
  parameters: {
    docs: {
      story: {
        // inline:falseにより、各Storyを個別のiframeで実行するように設定した
        // DocsページでもStoryが独立したiframeで動作し、TanStack Queryのキャッシュ競合を回避できる
        inline: false,
        iframeHeight: 200,
      },
    },
  },
};

export default meta;

type Story = StoryObj<typeof meta>;
```

#### 商品がある場合（Default）

商品が 3 件表示される状態です。
実際の UI でユーザーが商品を検索して結果が表示された状態を再現します。

```tsx
export const Default: Story = {
  parameters: {
    msw: {
      handlers: [
        buildGetProductsSearchHandler.success({
          response: generateProductsSearchMock({
            products: [
              generateProductInSearchMock({
                id: 1,
                title: "iPhone 9",
                description: "An apple mobile which is nothing like apple",
                category: "smartphones",
                price: 549,
                brand: "Apple",
              }),
              /** 同様のものを2件追加する　*/
            ],
            total: 3,
          }),
        }),
      ],
    },
  },
};
```

![商品がある場合の画面](/images/20251012_article-error-boundary-suspense-msw-storybook-test/features-Shop-ProductList-Result-Default.png)

商品がある場合の Storybook は[こちら](https://68e2285ed2a65b4c23a54763-idqbuptadi.chromatic.com/?path=/story/features-shop-productlist-result--default)です。

#### 商品がない場合（NoProduct）

検索結果が 0 件の場合、「商品がありませんでした。」というメッセージが表示されます。

```tsx
export const NoProduct: Story = {
  parameters: {
    msw: {
      handlers: [
        buildGetProductsSearchHandler.success({
          response: generateProductsSearchMock({
            products: [],
            total: 0,
          }),
        }),
      ],
    },
  },
};
```

商品がない場合の Storybook は[こちら](https://68e2285ed2a65b4c23a54763-idqbuptadi.chromatic.com/?path=/story/features-shop-productlist-result--no-product)です。

#### ローディング中（Loading）

API 通信中の状態です。
`<Suspense>`の fallback に設定している「商品一覧を読み込み中...」が表示されます。

```tsx
export const Loading: Story = {
  parameters: {
    msw: {
      handlers: [buildGetProductsSearchHandler.loading()],
    },
  },
};
```

ローディング中の Storybook は[こちら](https://68e2285ed2a65b4c23a54763-idqbuptadi.chromatic.com/?path=/story/features-shop-productlist-result--loading)です。

#### エラー（Error）

API でエラーが発生した場合の状態です。
`<ErrorBoundary>`の fallback に設定している「商品一覧でエラーが発生しました」が表示されます。

```tsx
export const Error: Story = {
  parameters: {
    msw: {
      handlers: [buildGetProductsSearchHandler.error({ status: 500 })],
    },
  },
};
```

エラーの Storybook は[こちら](https://68e2285ed2a65b4c23a54763-idqbuptadi.chromatic.com/?path=/story/features-shop-productlist-result--error)です。

## カートと今日の名言のコンポーネントの実装

商品一覧と同様のパターンなので簡単に説明します。
各コンポーネントが`<ErrorBoundary>`と`<Suspense>`を内包することで、独立してテストと Storybook が作成できます。

### カート

カートコンポーネントは、ユーザー ID を受け取ってカート内の商品を表示します。

```tsx
export const Cart = () => {
  return (
    <div>
      <h2>カートの中身</h2>
      <Content />
    </div>
  );
};
```

`<Content>`コンポーネントが`<ErrorBoundary>`と`<Suspense>`を内包し、エラーとローディングの範囲を局所化しています。

```tsx
export const Content = () => {
  return (
    <Error Boundary fallback={<div>カートの取得に失敗しました。</div>}>
      <Suspense fallback={<div>カートの読み込み中...</div>}>
        <Inner />
      </Suspense>
    </Error Boundary>
  );
};

const Inner = () => {
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

[こちら](https://68e2285ed2a65b4c23a54763-idqbuptadi.chromatic.com/?path=/story/features-shop-cart--default)ではカートの 4 つの状態（商品あり/なし/ローディング/エラー）を確認できます。

### 今日の名言

今日の名言コンポーネントは、外部 API から名言を取得して表示します。
このコンポーネントがエラーになっても、カートや商品一覧は正常に動作します。

```tsx
export const Quote = () => {
  return (
    <div>
      <h2>今日の名言</h2>
      <Content />
    </div>
  );
};
```

`<Content>`コンポーネントが`<ErrorBoundary>`と`<Suspense>`を内包し、エラーとローディングの範囲を局所化しています。

```tsx
export const Content = () => {
  return (
    <Error Boundary fallback={<Fallback />}>
      <Suspense fallback={<Loading />}>
        <Inner />
      </Suspense>
    </Error Boundary>
  );
};

const Inner = () => {
  const { data } = useQuote();

  return (
    <blockquote>
      <p>"{data.quote}"</p>
      <cite>— {data.author}</cite>
    </blockquote>
  );
};

const Fallback = () => {
  return (
    <blockquote>
      <p>今日の名言の取得に失敗しました。</p>
    </blockquote>
  );
};

const Loading = () => {
  return (
    <blockquote>
      <p>今日の名言を読み込み中...</p>
    </blockquote>
  );
};
```

[こちら](https://68e2285ed2a65b4c23a54763-idqbuptadi.chromatic.com/?path=/story/features-shop-quote--default)では今日の名言の 4 つの状態（成功/ローディング/エラー）を確認できます。

### カートと今日の名言の実装のポイント

カートと今日の名言のコンポーネントは、商品一覧と同じパターンで実装されています。

- `<Content>`コンポーネントが`<ErrorBoundary>`と`<Suspense>`を内包
- カスタムフックで`useSuspenseQuery`を使用してデータ取得
- テストと Storybook で 4 つの状態（成功/データなし/ローディング/エラー）を確認可能

## カート、商品一覧、今日の名言のコンポーネントを組み込む

これまで個別に実装してきた 3 つのコンポーネントを組み込みます。
ここでは、「よくない例」の`<CheapShop>`と「改善例」の`<Shop>`を比較します。

### CheapShop のよくない実装例

`<ErrorBoundary>`と`<Suspense>`がトップレベルにのみ配置されています。
`<Cart>`、`<ProductList>`、`<Quote>`の各コンポーネント内部には`<ErrorBoundary>`と`<Suspense>`がありません。

```tsx
export const CheapShop = () => (
  <Error Boundary fallback={<div>全画面エラーが発生しました</div>}>
    <Suspense fallback={<div>全画面読み込み中...</div>}>
      <div>
        <h1>Super coolなECサイト</h1>
        <Cart />
        <ProductList />
        <Quote />
      </div>
    </Suspense>
  </Error Boundary>
);
```

### Shop コンポーネントの実装の改善

各コンポーネント（`<Cart>`、`<ProductList>`、`<Quote>`）が独立して`<ErrorBoundary>`と`<Suspense>`を持っているため、1 つのコンポーネントがエラーやローディング状態でも、他のコンポーネントは正常に動作します。

```tsx
export const Shop = () => (
  <div>
    <h1>Super coolなECサイト</h1>
    <Cart />
    <ProductList />
    <Quote />
  </div>
);
```

### 今日の名言のみエラーになっている場合でも、その他の機能は使えることを確認

「よくない例」の`<CheapShop>`と「改善例」の`<Shop>`で、今日の名言がエラーになった場合の動作を Storybook で比較します。

`<Shop>`では、今日の名言の API がエラーになっても、その部分だけがエラー表示になります。
一方、カートと商品一覧は正常に表示され、ユーザーは EC サイトの主要機能を問題なく利用できます。

優先度の低い今日の名言のセクションで発生したエラーが、重要な機能に影響を与えないため、ユーザー体験が大きく向上します。

![今日の名言だけエラーになっている画像](/images/20251012_article-error-boundary-suspense-msw-storybook-test/total-view-with-quota-error.png)

今日の名言のみエラーになっているが、それ以外のカートと商品一覧は確認できる Storybook はこちらです。

```tsx
export const ErrorOnQuote: Story = {
  parameters: {
    msw: {
      handlers: [buildGetQuoteHandler.error({ status: 500 }), ...handlers],
    },
  },
};
```

ErrorOnQuote の Storybook は[こちら](https://68e2285ed2a65b4c23a54763-idqbuptadi.chromatic.com/?path=/story/features-shop--error-on-quote)です。

このように、Error Boundary と Suspense を適切に配置することで、エラーの影響範囲を最小限に抑え、ユーザー体験を大きく向上させることができます。

## まとめ

この記事では、Error Boundary と Suspense を適切に配置することの重要性を、実装例を通じて紹介しました。

### 基本的な考え方

API をコールするコンポーネントの親コンポーネントに`<ErrorBoundary>`と`<Suspense>`を配置することで、エラーとローディングの範囲を限定できます。

### このコンポーネント設計で得られること

今回紹介した Error Boundary と Suspense の適切な配置を意識したコンポーネント設計により下記のような 3 つのことが得られます。

1. ユーザー体験の向上

   - エラーやローディングを局所化することで、一部の機能が失敗しても他の機能は正常に動作
   - 優先度の低い機能のエラーが、主要機能に影響を与えない

2. テストと Storybook の書きやすさ

   - コンポーネント単体で 4 つの状態（成功/データなし/ローディング/エラー）をテスト可能
   - MSW を使って各状態を簡単に再現でき、デザイナーや PM も確認できる

3. 並列開発のしやすさ

   - 各コンポーネントが独立しているため、チームメンバーが同時並行で開発できる
   - Storybook で個別に開発・確認しながら、最終的に統合できる

## 備考

記事の本筋からは外れますが、テストと Storybook で利用したユーティリティについて説明します。

### モックの生成関数と MSW ハンドラのビルダー

テストと Storybook で MSW を使う際、モックデータの生成とハンドラの作成を効率化するユーティリティを用意しています。

参考: [MSW の成功・失敗・ローディング・カスタムレスポンス・引数のテストをラクにするハンドラービルダー関数](https://tech.jxpress.net/entry/2025/01/14/103618)

#### モックの生成関数

API レスポンスのモックデータを生成する関数です。
テストや Storybook で一貫したデータ構造を簡単に作成できます。

```ts
export const generateProductsSearchMock = (
  override?: Partial<ProductsSearchResponse>
): ProductsSearchResponse => ({
  products: [],
  total: 0,
  skip: 0,
  limit: 30,
  ...override,
});
```

#### MSW ハンドラのビルダー

モック生成関数を使って、MSW のハンドラを簡単に作成できます。

```ts
// ハンドラービルダーを作成
const buildGetProductsSearchHandler = buildHttpHandlerBuilder({
  path: "/products/search",
  method: "get",
  defaultResponse: generateProductsSearchMock(),
});

// 正常系
const successHandler = buildGetProductsSearchHandler.success({
  response: generateProductsSearchMock({ products: [...], total: 3 }),
});

// ローディング
const loadingHandler = buildGetProductsSearchHandler.loading();

// エラー
const errorHandler = buildGetProductsSearchHandler.error({ status: 500 });

// リクエストパラメータのキャプチャ
const onRequestSearchParams = vi.fn();
const captureHandler = buildGetProductsSearchHandler.success({
  response: generateProductsSearchMock(),
  onRequestSearchParams,
});
```

### customRender と TestProvider

テストでコンポーネントをレンダリングする際、`QueryClientProvider`などの必要な Provider でラップする必要があります。
`TestProvider`で Provider をまとめ、`customRender`でそれを使ってレンダリングすることで、テストコードをシンプルに保ちます。

参考: [Setup | Testing Library](https://testing-library.com/docs/react-testing-library/setup/#custom-render)

#### TestProvider

テストで必要な Provider をまとめたコンポーネントです。

```tsx
export const TestProvider = ({ children }: Props) => {
  // NOTE: Storybookやテストごとに新たにクライアントを作らないと最初の表示したものが別のケースに依存してしまう。
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        staleTime: 0,
      },
    },
  });

  return (
    <QueryClientProvider client={queryClient}>
      <Error Boundary fallback={<div>全画面エラー</div>}>
        <Suspense fallback={<div>全画面読み込み中...</div>}>
          {children}
        </Suspense>
      </Error Boundary>
    </QueryClientProvider>
  );
};
```

#### customRender

`TestProvider`を使ってコンポーネントをラップする独自のレンダー関数です。

```tsx
export const customRender = (
  ui: ReactElement,
  options?: Omit<RenderOptions, "wrapper">
) => render(ui, { wrapper: TestProvider, ...options });
```

通常の`render`を使う場合、毎回 Provider を手動でラップする必要があります。

```tsx
// render を使う場合（冗長）
it("商品が表示される", async () => {
  render(
    <QueryClientProvider client={queryClient}>
      <Error Boundary fallback={<div>エラー</div>}>
        <Suspense fallback={<div>読み込み中</div>}>
          <Result query="iPhone" />
        </Suspense>
      </Error Boundary>
    </QueryClientProvider>
  );
  expect(await screen.findByText("iPhone 15 Pro")).toBeInTheDocument();
});
```

`customRender`を使えば、Provider のラップが自動化されシンプルになります。

```tsx
// customRender を使う場合（シンプル）
it("商品が表示される", async () => {
  customRender(<Result query="iPhone" />);
  expect(await screen.findByText("iPhone 15 Pro")).toBeInTheDocument();
});
```

## 参考

- [Component - Catching rendering errors with an error boundary | React](https://ja.react.dev/reference/react/Component#catching-rendering-errors-with-an-error-boundary)
- [Suspense - React](https://ja.react.dev/reference/react/Suspense)
- [useDeferredValue - React](https://ja.react.dev/reference/react/useDeferredValue)
- [TanStack Query - Installation](https://tanstack.com/query/latest/docs/framework/react/installation)
- [DummyJSON - Fake REST API](https://dummyjson.com/docs)
- [MSW の成功・失敗・ローディング・カスタムレスポンス・引数のテストをラクにするハンドラービルダー関数](https://tech.jxpress.net/entry/2025/01/14/103618)
- [Setup | Testing Library](https://testing-library.com/docs/react-testing-library/setup)
- [Auto Cleanup in Vitest | Testing Library](https://testing-library.com/docs/react-testing-library/setup#auto-cleanup-in-vitest)
