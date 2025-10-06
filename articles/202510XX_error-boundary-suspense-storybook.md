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

## 商品の一覧表示機能の実装

ここからは具体的な実装に移ります。
一番複雑な商品一覧のコンポーネント、テスト、Storybook の実装を行います。
カート、今日の名言のコンポーネントも商品一覧と同様に作成すれば良いため詳細な説明は省略します。
最初に記載したブランチも参考にしてください。

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

```ts
describe("fetchProducts", () => {
  it("正常にProductsSearchResponseを返し、クエリパラメータが正しく含まれる", async () => {
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

#### 検索ボックスと商品一覧を表示する`<ProductList>`コンポーネント

入力フォームと検索結果の表示を担当します。
実際の API 呼び出しとデータ表示は、次に説明する`<Result>`コンポーネントに委譲しています。

TODO:商品一覧の画像を貼る。検索ボックスと一覧の表示がわかるように赤枠で囲う。

上部はテキスト入力欄で、そこに入力された内容に基づいて検索を行い、下部に結果を表示します。

```tsx
export const Shop = () => {
  return (
    <div>
      <h1>Super coolなECサイト</h1>
      <Cart />
      <ProductList />
      <Quote />
    </div>
  );
};
```

TODO:useDeferredValue の公式サイトのリンクを貼る

本筋とは関係ないですが、入力するたびに一瞬ローディング画面が表示されるチラつきを防ぐために、React v19 から登場した`useDeferredValue`を利用しています。
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

**この設計のポイント：**

- `<Result>`コンポーネントが`<ErrorBoundary>`と`<Suspense>`を持つことで、このコンポーネント単体でテストや Storybook が作成できます
- 実際のデータ取得と表示ロジックは`<Inner>`コンポーネントに分離し、責任を明確化しています
- エラーやローディングの処理を`<Inner>`から分離することで、`<Inner>`は「データをどう表示するか」だけに集中できます

```tsx
type Props = {
  query: string;
};

export const Result: FC<Props> = ({ query }: Props) => {
  return (
    <ErrorBoundary fallback={<div>商品一覧でエラーが発生しました</div>}>
      <Suspense fallback={<div>商品一覧を読み込み中...</div>}>
        <Inner query={query} />
      </Suspense>
    </ErrorBoundary>
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

下記の 4 つのテストを実装します：

- 商品がある場合
- 商品がない場合
- ローディング中
- エラー

#### テストのセットアップ

テストでは`customRender`という独自のレンダー関数を使用しています。
これは、`<QueryClientProvider>`などの必要な Provider でコンポーネントをラップするためのユーティリティです。
詳細は「備考」セクションを参照してください。

TODO:できれば備考セクションへのリンクをつける

#### 商品がある場合（正常系）

商品が 1 件ある場合のテストです。
このテストでは以下を確認します：

- 商品情報が正しく表示される
- 商品件数が正しく表示される
- クエリパラメータ`q`が正しく API リクエストに含まれている

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

#### 商品がない場合（正常系）

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

## カートと今日の名言のコンポーネントの実装

商品の一覧コンポーネントと同様の流れで行うので簡単に書く。

### カート

## カート、商品一覧、今日の名言のコンポーネントを組み込む

TODO:今日の名言がエラーになっても、カートや商品一覧などの重要な機能は利用できることを示す。

### 今日の名言

## まとめ

- エラーとローディングを局所化する設計の利点
- テストとコンポーネントカタログを意識することで得られる開発体験
- 本番 API へ置き換える際にも活きる設計指針

```

```

## 備考

### MSW ハンドラのビルダー

同じ記述の繰り返しを避けるために、テストと Storybook で利用する MSW のハンドラを作成するユーティリティを作成します。
TODO: 参考にしているページを明記する

### モックの生成関数

### customRender

### TestProvider

## 参考
