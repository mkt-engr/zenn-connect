---
title: "Nextjsでnuqsを用いて検索条件をURLのクエリパラメタを型安全に保存する検索機能の実装（サーバーコンポーネントでAPIコールするよ）" # 記事のタイトル
emoji: "😊" # アイキャッチとして使われる絵文字（1文字だけ）
type: "tech" # tech: 技術記事 / idea: アイデア記事
topics: ["Nextjs", "nuqs", "Server Component"] # タグ。["markdown", "rust", "aws"]のように指定する
published: true # 公開設定（falseにすると下書き）
---

# 概要

URL のクエリパラメタを型安全に扱うことができるライブラリ[nuqs](https://nuqs.47ng.com/)を Nextjs で利用し、検索条件を URL のクエリパラメタを型安全に保存する検索機能を作成してみました。

Nextjs で検索機能を実装しようとしたとき、まず最初に浮かぶのが`useState`で検索条件を保持して API をコールする手法かと思います。ただ、その場合だと

- リロードすると検索結果がリセットされてしまう
- 検索結果を他人にシェアすることができない

といったデメリットがあります。そこで検索条件を URL のクエリパラメタに保存することでこれらのデメリットを解消することができました。

GitHub のリポジトリは[こちら](https://github.com/mkt-engr/playground-nuqs/tree/main)です。
実装内容を反映したものは[こちら](https://playground-nuqs-kxjukqgg0-mktengrs-projects.vercel.app/search)の URL から確認できます。

## 主に利用したライブラリとそのバージョン

主に利用したライブラリとそのバージョンはこちらです。

| ライブラリ | バージョン |
| ---------- | ---------- |
| next       | 15.1.2     |
| nuqs       | ^2.2.3     |
| typescript | ^5         |

ダミーの API には[DummyJson](https://dummyjson.com/docs/products#products-search)を用いました。

# 実装手順

大きく以下の3つを実装します。

- クエリパラメタのスキーマ
- 検索フォームのコンポーネント
- 検索結果を表示するコンポーネント
## クエリパラメタのスキーマ

クエリパラメタを型安全にするためにスキーマを作成します。`withDefault`を利用してクエリパラメタのデフォルト値を空文字にしておきます。

```tsx
import { createSearchParamsCache, parseAsString } from "nuqs/server";

export const searchParamsCache = createSearchParamsCache({
  /** 検索クエリ */
  q: parseAsString.withDefault(""),
});
```

## 全体感

以下に 2 種類のコンポーネントを作成します。API のコールはサーバーコンポーネントで行います。

- 検索フォームのコンポーネント
- 検索結果を表示するコンポーネント

```tsx
type Props = {
  /** 検索クエリを含む検索パラメータ */
  searchParams: Promise<SearchParams>;
};

/** 検索フォームと検索結果を表示するコンポーネント */
export default async function Page({ searchParams }: Props) {
  // 検索クエリのパース
  const { q } = await searchParamsCache.parse(searchParams);

  // 検索クエリをもとに商品を取得
  const { products } = await fetchProducts({ q });

  return (
    <div>
      <Search />
      <hr />
      <Result products={products} />
    </div>
  );
}
```

## 検索フォームのコンポーネント(`<Search />`)

検索フォームは入力欄を含むのでクライアントコンポーネントで実装します。

`nuqs`の`useQueryState`を用いて React のステートとクエリパラメタを繋げます。
実際に入力してみるとわかりますが、入力した内容がクエリパラメタ(例.`search/?q=fefefe`)に反映されています。

```tsx
"use client";

export const Search = () => {
  const [query, setQuery] = useQueryState("q", { defaultValue: "" });

  const handleSubmit = useHandleSubmit();

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        name="q"
      />
      <button type="submit">検索する</button>
    </form>
  );
};
```

なお`useHandleSubmit`の中身は下記のようになっており、以下の役割を果たしています。

- フォームのデフォルトの挙動（フォームの送信）を塞ぐ
- （サーバコンポーネントで検索を行うために）`router.refresh()`を実行する（★）

とくに 2 つ目の（★）が重要です。これによりクライアントコンポーネントで更新したクエリパラメタを用いてサーバコンポーネントで検索を行っています。

```tsx
export const useHandleSubmit = () => {
  const router = useRouter();

  const handleSubmit = useCallback(
    (e: FormEvent<HTMLFormElement>) => {
      e.preventDefault();

      // URLのクエリパラメタに検索条件が保持されているので、refreshすることでサーバーコンポーネントで再度検索を行うことができる。
      router.refresh();
    },
    [router]
  );

  return handleSubmit;
};
```

## 検索結果を表示するコンポーネント(`<Result />`)

API から取得した商品の情報を表示しているだけで特段難しいことはしていません。
クエリパラメタに検索条件を保存しているおかげで、リロードしても検索結果が保持されていることがわかります。

```tsx
/** 商品の検索結果ーを表示するコンポーネント */
export const Result: FC<Props> = ({ products }) =>
  products.map((product) => {
    return (
      <div key={product.id}>
        <h2>title:{product.title}</h2>
        <p>price:{product.price}</p>
      </div>
    );
```

# 結論

- `nuqs`を使うとURLのクエリパラメタとReactのステートを簡単に型安全に繋ぐことができる
- 検索条件をクエリパラメタに保存するとリロードしても検索結果が保持され、URLを用いて検索結果を他人に共有することもできる

# 参考

- https://nuqs.47ng.com/
- https://nextjs.org/docs/app/api-reference/functions/use-router
- https://nextjs.org/docs/app/api-reference/functions/fetch
- https://dummyjson.com/docs
