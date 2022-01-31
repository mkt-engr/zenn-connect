---
title: "Next.js × TypeScript × microCMS × Tailwind CSSでJamStackなブログを作ってみた"
emoji: "🕌"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: ["Next", "TypeScript", "microCMS", "TailwindCSS"]
published: false
---

# 概要

こんにちは。まっきんとっしゅです([@mkt_phys](https://twitter.com/mkt_phys))。今回は Next.js と microCMS を使って Jamstack なブログを作ってみました。リポジトリは[こちら](https://github.com/mkt-engr/jamstack_blog)です(公開日時点のブランチは release/2022_0131 です)。実際のブログは[こちら](https://mkt-blue.com) です（今は完全にブログを作ることがゴールになってしまってます泣）。

## 作ったもの

作成したブログの URL は[こちら](https://mkt-blue.com/)です。ほんとにシンプルな構成でページとしては記事一覧ページと記事詳細ページだけです笑。具体的には以下の通りです。検索機能とか、タグとかは作ってないです(作ります)。

- 記事一覧ページ
- 記事詳細ページ
- シンタックスハイライト
- SG(Static Generation)
- ISR(Incremental Static Regeneration)
- CSR(Client-Side Rendering 、クライアントフェッチ)
- Google Analytics 連携

## やってないこと

ここでやってないことをつらつらと書き並べます。

- Prettier
- Linter
- ダークモード
- 検索機能
- ページネーション
- コメント機能
- 僕の SNS へのリンク
- PWA

特に上 2 つは追加しないとダメですね。今後記事書きます！

# 主なライブラリのバージョン

作成したブログに用いた主なライブラリのバージョンはこちらです。

- Next : 12.0.4
- React : 17.0.2
- TypeScript : 4.5.2
- SWR : 1.1.2
- Tailwind CSS : 2.2.19

## プロジェクトのセットアップ

### Next のプロジェクトの作成

なにはともあれ Create Next App ですね。

```sh
npx create-next-app jamstack-blog
```

いつの間にかデフォルトで yarn ではなく npm になってました。このコマンドでオプションで`--ts`をつけていれば TypeScript が導入できたのですがうっかりオプションをつけるのを忘れていました。

[こちら](https://maku.blog/p/ny9fmty/)のサイトに従って既存のプロジェクトに TypeScript を導入します。

```sh
touch tsconfig.json # プロジェクトのルートでtsconfig.jsonを作成
npm install typescript @types/node @types/react # TypeScriptと型の情報をインストール
npm run dev # tsconfig.jsonに書き込まれる。
```

デフォルトだと tsconfig.json の strict が false になっているのでこちらを true に変更しました。true に変更することで暗黙的な`any`型が使われている場合にエラーを吐き出すようになります。TypeScript は初心者ですが TypeScript に少しでも慣れるために制限を厳しくしてます笑。このオプションのせいでビルドの時にエラーがでまくりましたが自分のためと思って頑張りました。

```json:tsconfig.json
{
 "compilerOptions": {
    "target": "es5",
    "lib": [
      "dom",
      "dom.iterable",
      "esnext"
    ],
    ~~省略~~
    "strict":true // こちらを変更
}
```

### Tailwind CSS の導入

[公式サイト](https://v2.tailwindcss.com/docs/guides/nextjs)の手順に従って Tailwind CSS を導入します。

1. Tailwind CSS のインストール

```sh
npm install -D tailwindcss@latest postcss@latest autoprefixer@latest
npx tailwindcss init -p
```

2 つ目のコマンドでルート直下に tailwind.config.js と postcss.config.js がつくられます。

2. tailwind.config.js の編集
   `purge`オプションで
   - どのディレクトリに置いた
   - どの拡張子のものに

Tailwind CSS を適用させるかを記述します。また実装で使われることがなかった Tailwind CSS のクラスは実装の時にビルドされなくなります。

```js
module.exports = {
  purge: ["./pages/**/*.{js,ts,jsx,tsx}", "./components/**/*.{js,ts,jsx,tsx}"],
  darkMode: false, // or 'media' or 'class'
  theme: {
    extend: {
      fontFamily: {
        rock: ["Rock Salt"],
      },
    },
  },
  variants: {},
  plugins: [],
};
```

これでセットアップは完了です。

なお`fontFamily`の部分はかっちょいい GoogleFont を使うために定義しています。`font-rock`のクラスを適用することでかっちょいいフォントになります。

## Google Analytics と Favicon の設定(`_document.tsx`)

Google Analytics と Favicon は Head タグに書く必要があるのでここに書きます。

### `_document.tsx`

Google Analytics とファビコンに関する記述はコンポーネント化しているので`_document.tsx`ではおれを読み込んでいるだけです。

```ts:_document.tsx
import NextDocument, { Html, Head, Main, NextScript } from "next/document";
import Favicon from "../components/Favicons";
import GA from "../components/GA";

type Props = {};
class Document extends NextDocument<Props> {
  render() {
    return (
      <Html lang="ja">
        <Head>
          {/*Google Analyticsのコンポーネント*/}
          <GA />
          {/*ファビコン関連のコンポーネント*/}
          <Favicon />
        </Head>
        <body className="leading-relaxed box-content">
          <Main />
          <NextScript />
        </body>
      </Html>
    );
  }
}

export default Document;
```

### Google Analytics のイベントを発火させる関数(`gtag.ts`)

`_document.tsx`にあった`GA`コンポーネントで使う関数を作ります。型を使いたいのでまず最初に型情報を import します。

```sh
npm i @types/gtag.js
```

Vercel のリポジトリに[サンプル](https://github.com/vercel/next.js/blob/canary/examples/with-google-analytics/lib/gtag.js)があったのでそのまま使います。

```ts:gtag.ts
export const GA_TRACKING_ID = process.env.NEXT_PUBLIC_GOOGLE_ANALYTICS_ID || "";

//参考： https://developers.google.com/analytics/devguides/collection/gtagjs/pages
export const pageview = (url: string): void => {
  if (!GA_TRACKING_ID) return;
  window.gtag("config", GA_TRACKING_ID, {
    page_path: url,
  });
};

//参考： https://developers.google.com/analytics/devguides/collection/gtagjs/events
type GaEventProps = {
  action: string;
  category: string;
  label: string;
  value?: number;
};

export const event = ({
  action,
  category,
  label,
  value,
}: GaEventProps): void => {
  if (!GA_TRACKING_ID) return;
  window.gtag("event", action, {
    event_category: category,
    event_label: label,
    value,
  });
};
```

型を import したのは`window.gtag`を使うためです。

### Google Analytics

[こちら](https://github.com/vercel/next.js/blob/canary/examples/with-google-analytics/pages/_app.js)に Vercel が作成したサンプルコードの中に GA を組み込んだ物があったのでそれを採用します。

```ts:GA.tsx
import { VFC } from "react";
import { GA_TRACKING_ID } from "../lib/gtag";
import Script from "next/script";
const GA: VFC = () => {
  return (
    <>
      {GA_TRACKING_ID && (
        <>
          <Script
            strategy="afterInteractive"
            src={`https://www.googletagmanager.com/gtag/js?id=${GA_TRACKING_ID}`}
          />
          <Script
            strategy="afterInteractive"
            dangerouslySetInnerHTML={{
              __html: `
                window.dataLayer = window.dataLayer || [];
                function gtag(){dataLayer.push(arguments);}
                gtag('js', new Date());
                gtag('config', '${GA_TRACKING_ID}', {
                  page_path: window.location.pathname,
                });`,
            }}
          />
        </>
      )}
    </>
  );
};

export default GA;
```

`Script`タグ(デフォルトで用意してある頭文字が小文字の script ではない)には`strategy`と呼ばれる属性を設定することができます。この属性には 3 つのオプションがあります。それぞれ

- `beforeInteractive`：ページが操作可能になる前に src 属性に書かれたスクリプトをロードする
- `afterInteractive`：ページが操作可能になった直後 src 属性に書かれたスクリプトをロードする
- `lazyOnload`：ページが操作可能になってアイドル状態になった後に src 属性に書かれたスクリプトをロードする

詳しくは[Next の公式ページ](https://nextjs.org/docs/basic-features/script) や[こちらの記事](https://zenn.dev/aiji42/articles/9a6ab12ab5f6e6)のページをご覧ください。

### Favicon

[こちら](https://qiita.com/purpleeeee/items/cd9aca1ae735ad678355) の記事で紹介されている Favicon Generator というサイトを使ってファビコンの生成を行いました。こちらのサイトに画像をアップロードすると、デスクトップ用、モバイル用などのファビコンが生成されます。

```ts:Favicons.tsx
import { VFC } from "react";

const Favicon: VFC = () => {
  return (
    <>
      <link
        rel="apple-touch-icon"
        sizes="180x180"
        href="/favicons/apple-touch-icon.png"
      />
      <link
        rel="icon"
        type="image/png"
        sizes="32x32"
        href="/favicons/favicon-32x32.png"
      />
      <link
        rel="icon"
        type="image/png"
        sizes="16x16"
        href="/favicons/favicon-16x16.png"
      />
      <link rel="manifest" href="/favicons/site.webmanifest" />
      <link
        rel="mask-icon"
        href="/favicons/safari-pinned-tab.svg"
        color="#5bbad5"
      />
      <meta name="msapplication-TileColor" content="##3B1EEc" />
      <meta name="theme-color" content="#ffffff" />
    </>
  );
};

export default Favicon;
```

## Google Analytics をページ遷移時にも対応させる(`_app.tsx`)

これも Vercel の(サンプル)[https://github.com/vercel/next.js/blob/canary/examples/with-google-analytics/pages/_app.js]にあります。

```ts:_app.tsx
import { useEffect } from "react";
import { useRouter } from "next/router";
import { AppProps } from "next/app";
import * as gtag from "../lib/gtag";
import "../styles/globals.scss";

const MyApp = ({ Component, pageProps }: AppProps) => {
  // Google Analyticsをページ遷移時にも対応させる
  const router = useRouter();
  useEffect(() => {
    const handleRouteChange = (url: string) => {
      gtag.pageview(url);
    };
    router.events.on("routeChangeComplete", handleRouteChange);
    return () => {
      router.events.off("routeChangeComplete", handleRouteChange);
    };
  }, [router.events]);
  return <Component {...pageProps} />;
};

export default MyApp;

```

`router.events.on`、の第一引数はイベントを発火させるタイミングを表ています。`routeChangeComplete`はルーティングが完了した時にイベントを発火させます。もう 1 つの`router.events.off`でイベントを unsubscribe します。

## ブログサイト全体のレイアウトを決める (`Layout.tsx`)

単純なレイアウトです。上からヘッダー、コンテンツ、フッターがあるレイアウトです。

```ts:Layout.tsx
import { ReactNode, VFC } from "react";
import Head from "next/head";
import Header from "../common/Header";
import Footer from "../common/Footer";

interface Props {
  children: ReactNode;
  title?: string;
}
const Layout: VFC<Props> = ({ children, title }) => {
  return (
    <>
      <Head>
        <title>{title}</title>
      </Head>
      <div className={`flex bg-gray-100 flex-col h-screen`}>
        <Header />
        <div className="flex-1 px-4 md:px-18 xl:px-36 bg-gray-100 blogContent">
          <main>{children}</main>
        </div>
        <Footer />
      </div>
    </>
  );
};

export default Layout;
```

h1,h2,h3 タグにグローバルなスタイルを適用させたかったので`blogContent`というクラスを作成しました。`blogContent`の下の h1,h2,h3 タグにスタイルが適用されるようにしています。
:::details global.scss

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

// ブログ詳細のスタイル
@layer base {
  .blogContent {
    h1 {
      @apply text-4xl sm:text-5xl font-bold mt-8 mb-4;
    }
    h2 {
      @apply mb-0  text-xl sm:text-2xl font-bold;
    }
    h3 {
      @apply text-xl sm:text-lg;
    }
  }
}
```

:::

## 記事一覧ページ(`pages/index.tsx`)

このファイルで SG、ISR、CSR を実現しています。

まず SG と ISR には`getStaticProps`を使って実現しています。この後すぐ出てきますが`getStaticProps`の型は`InferGetStaticPropsType`を使って定義しています。

```ts:getStaticProps
export const getStaticProps = async () => {
  const data = await getAllArticles();

  return {
    props: { staticArticles: data.contents },
    revalidate: 3,
  };
};
```

`getAllArticles()`は micro CMS から全記事記事を取得する処理です。こちらは外部ファイル化しました。
:::details getAllArticles()

```ts
export const getAllArticles = async (): Promise<CONTENTS> => {
  const options: AxiosRequestConfig = {
    url: `${process.env.API_URL}/blog`,
    method: "GET",
    headers: { "X-MICROCMS-API-KEY": process.env.API_KEY! },
  };

  const res = await axios(options);

  const { data }: { data: Promise<CONTENTS> } = res;
  return data;
};
```

:::

ISR ってものすごくありがたい機能なのですが再ビルドが走るのが１度アクセスがあってからなんですよね。ということは micro CMS で記事をアップしてから最初に見てくれる人には古いコンテンツを見せることになってしまいます。むしろ最初にきてくれた人に最新の記事を見せたいので CSR も実装します（正直なところ小さい自分のブログなので更新後自分でアクセスしちゃえば問題ないといえば問題ないですね..。このせいで Lighthouse の点数も下がっているような気がします）。CSR には SWR を使うことにしました。

記事を表示している部分も含めると以下の通りです。ちなみに一行目にあるのが`getStaticProps`で取得した値の型です。staticArticles は`getStaticProps`で取得した記事です。それを`useSWR`の`fallbackData`に設定することで CSR するまえのデータを画面に表示しています。その後 CSR した最新のデータを表示するという流れです。なので記事の公開後に初めてブログにアクセスした人は一瞬古い状態のサイトが見えます。

CSR を実装しているのは以下の部分です。

```ts:index.tsxの一部
const Home: NextPage<Props> = ({ staticArticles }) => {
  const { data: articles, mutate } = useSWR<ARTICLE[]>("/api/blog/", fetcher, {
    fallbackData: staticArticles,
  });

  useEffect(() => {
    //SWRで取得するデータを最新化する
    mutate();
  }, [mutate]);

  return <>...</>
}
```

CSR で記事を取得する際も micro CMS のシークレットキーが必要です。単純にシークレットキーを設定した環境変数に`NEXT_PUBLIC`をつければ CSR を実装できます。しかし`NEXT_PUBLIC`のプレフィックスをつけると JavaScript にインラインで公開されてしまいます（ブラウザで容易にシークレットキーが見つけられてしまう）。そこで今回は CSR は API ルートを経由することにしました。

`page/`ディレクトリの下に`api/`ディレクトリを作るとそこが API のエンドポイントとして扱われます。

やっていることは単純で`api/blog`にアクセスしたら全記事を返却しているだけです。(`getStaticProps`で実装したものとほぼ同じ内容です)

```ts:api/blog/index.ts
import type { NextApiRequest, NextApiResponse } from "next";
import { getAllArticles } from "../../../lib/articles";

const updateTopPage = async (req: NextApiRequest, res: NextApiResponse) => {
  // micro CMSから全データを取得する
  const data = await getAllArticles();
  res.status(200).json(data.contents);
};

export default updateTopPage;
```

以上を全てまとめたものが以下のコードです。
::: details index.tsx

```ts:index.tsx
import { GetStaticPaths, InferGetStaticPropsType, NextPage } from "next";
import { useRouter } from "next/router";
import { getAllArticleIds, getArticleById } from "../lib/articles";
import { formatYYYYMMDD } from "../lib/dayjs";
import { highlightByHighlightJs } from "../lib/highlightCode";
import "highlight.js/styles/hybrid.css";
import Layout from "../components/top/Layout";
type Props = InferGetStaticPropsType<typeof getStaticProps>;

const Home: NextPage<Props> = ({ staticArticles }) => {
  const { data: articles, mutate } = useSWR<ARTICLE[]>("/api/blog/", fetcher, {
    fallbackData: staticArticles,
  });

  useEffect(() => {
    //SWRで取得するデータを最新化する
    mutate();
  }, [mutate]);

  return (
    <Layout title="Mkt Memo">
      <div className="py-2 space-y-4">
        {articles?.map((article) => {
          return (
            <Link href={`/${article.id}`} key={article.id}>
              <a className="block">
                <Card {...article} />
              </a>
            </Link>
          );
        })}
      </div>
    </Layout>
  );
};

export default Home;

export const getStaticProps = async () => {
  const data = await getAllArticles();

  return {
    props: { staticArticles: data.contents },
    revalidate: 3,
  };
};
```

なお、NextPage 型は

> `Page` type, use it as a guide to create `pages`.

だそうです。pages ディレクトリ下で使うファイルだということを明示するために使う型のようです。VFC を使ってもエラーは出なかったのでどちらでも問題はないのかなと思います。
:::

## 記事詳細ページ(`pages/[id].tsx`)

このファイルでシンタックスハイライトや存在しないページへアクセスした時の対応をしています。

### ビルド時に生成するパスの取得

今回作成したブログの URL は記事の ID を採用しています（`pages/[id].tsx`のようにファイルを作成して Dynamic Routes を利用しています）。ビルドする前（HTML を生成する前）に Next 側は記事の ID つまり`[id]`に入る部分が何がわからないので`getStaticPaths`を用いて以下のように全記事の ID を取得します。

```ts:getStaticPaths
export const getStaticPaths: GetStaticPaths = async () => {
  const paths = await getAllArticleIds();//全ての記事のIDを取得
  return { paths, fallback: true };
};
```

なお`getAllArticleIds`は以下の通りです。
::: details getAllArticleIds

```ts:getAllArticleIds
export const getAllArticleIds = async (): Promise<ArticleId[]> => {
  const options: AxiosRequestConfig = {
    url: `${process.env.API_URL}/blog`,
    method: "GET",
    headers: { "X-MICROCMS-API-KEY": process.env.API_KEY! },//!をつけてnullでないことを明示する
  };

  const res = await axios(options);
  const articles: ARTICLE[] = res.data.contents;
  //IDだけを抽出して返却
  return articles.map((article) => {
    return {
      params: {
        id: String(article.id),
      },
    };
  });
};
```

:::

`getStaticPaths`で取得した`paths`の配列の要素の数`getStaticProps`がビルド時に実行されます。

```ts:getStaticProps
export const getStaticProps = async ({ params }: ParamType) => {
  //記事のIDを元に記事を取得(params.idのidは[id].tsxのidと対応している)
  const article = await getArticleById(params.id);//getArticleByIdのメソッドは後述します。

  //記事が取得できなかった場合はトップページへリダイレクトする
  if (!article) {
    return {
      redirect: {
        destination: "/",
        permanent: false,
      },
    };
  }

  //シンタックスハイライトをつける
  const body = highlightByHighlightJs(article.body);

  return {
    props: {
      article: { ...article, body },
    },
    revalidate: 1,
  };
};

```

### ビルド時に生成されなかったページにアクセスしようとした時の対応

`getStaticProps`はビルド時に実行されますがビルド時以外にも実行される場合があります。それは以下の 2 つの条件を満たす場合です。

- `getStaticPaths`の返り値を`fallback:true`にしている
- `getStaticPaths`の`paths`の中に含まれて**いない**URL にアクセスした

ビルド時以外に実行されて何が嬉しいかを説明します。例えば

- ビルド前：ID が１、２、３(３記事の HTML が生成された)

の場合を考えます。ビルド後に ID が４の記事が公開されたあと`/4`の URL にアクセスした時に`getStaticProps`が実行されて記事が取得され、ブラウザに表示されます。

`getStaticProps`が実行されている間にローディング画面などの待機画面を表示したい時は`next/router`の`router.isFallback`を用います。記事が取得できるまでは「Loading」が表示されます。

```ts:待機画面の実装([id].tsx)
const Blog: NextPage<Props> = ({ article }) => {
  const router = useRouter();
  if (router.isFallback || !article) {
    return <div>Loading...</div>;
  }
}
```

### まだ問題が残っていた

存在しない URL にアクセスした時の対処がまだでした。今のままだと`/fefefe`などにアクセスした場合は ID が`fefefe`の記事を探しに行きますがそのような記事はないのでエラーで落ちてしまいます。つまり記事を取得しにいく実装でエラーハンドリングが必要です。それを踏まえて作成した ID を元に記事を取得するメソッドが以下の通りです。

```ts:getArticleById
export async function getArticleById(id: string): Promise<ARTICLE> {
  const options: AxiosRequestConfig = {
    url: `${process.env.API_URL}/blog/${id}`,
    method: "GET",
    headers: { "X-MICROCMS-API-KEY": process.env.API_KEY! },
  };

  let res: AxiosResponse<ARTICLE>;

  try {
    res = await axios(options);
  } catch (e) {
    //記事が取得できなかった場合の処理
    if (axios.isAxiosError(e) && e.response?.status === 404) {
      return e.response?.data;
    }
  }
  return res!.data;
}
```

上記の実装では記事を取得できなかった場合は`null`を返しています。この`null`を利用してトップページへ返す実装をしています。(404 ページを作る必要がユーザーフレンドリーですよね..)

```ts:getStaticProps(一部)
export const getStaticProps = async ({ params }: ParamType) => {
  const article = await getArticleById(params.id);

  //記事が取得できなかった場合はトップページへリダイレクトする
  if (!article) {
    return {
      redirect: {
        destination: "/",
        permanent: false,
      },
    };
  }
  ///省略
}
```

### `[id].tsx`の全容

最後に全容をお見せします。
::: details [id].tsx

```ts:[id].tsx
import { GetStaticPaths, InferGetStaticPropsType, NextPage } from "next";
import { useRouter } from "next/router";
import { getAllArticleIds, getArticleById } from "../lib/articles";
import { formatYYYYMMDD } from "../lib/dayjs";
import { highlightByHighlightJs } from "../lib/highlightCode";
import "highlight.js/styles/hybrid.css";
import Layout from "../components/top/Layout";

type Props = InferGetStaticPropsType<typeof getStaticProps>;

const Blog: NextPage<Props> = ({ article }) => {
  const router = useRouter();
  if (router.isFallback || !article) {
    return <div>Loading...</div>;
  }
  const { title, body, createdAt, updatedAt } = article;

  return (
    <Layout title={title}>
      <div className="p-4 md:p-12 bg-white rounded">
        <div className="text-center text-4xl font-bold mb-2">{title}</div>
        <div className="space-x-2 text-right">
          <div className="">作成日 : {formatYYYYMMDD(createdAt)}</div>
          <div className="">更新日 : {formatYYYYMMDD(updatedAt)}</div>
        </div>
        <div
          dangerouslySetInnerHTML={{
            __html: `${body}`,
          }}
        ></div>
      </div>
    </Layout>
  );
};

export default Blog;

// 静的生成のためのパスを指定する(ビルド時に実行)
export const getStaticPaths: GetStaticPaths = async () => {
  const paths = await getAllArticleIds();
  return { paths, fallback: true };
};

interface ParamType {
  params: {
    id: string;
  };
}

//params.idでダイナミックルートの値が取得できる([id].tsxの[id]の部分)
export const getStaticProps = async ({ params }: ParamType) => {
  //記事のIDを元に記事を取得(params.idのidは[id].tsxのidと対応している)
  const article = await getArticleById(params.id);

  //記事が取得できなかった場合はトップページへリダイレクトする
  if (!article) {
    return {
      redirect: {
        destination: "/",
        permanent: false,
      },
    };
  }

  //シンタックスハイライトをつける
  const body = highlightByHighlightJs(article.body);

  return {
    props: {
      article: { ...article, body },
    },
    revalidate: 1,
  };
};

```

:::

作成日、更新日をフォーマットするメソッド`formatYYYYMMDD`とシンタックスハイライトをつけるメソッド`highlightByHighlightJs`も載せておきます。
::: details formatYYYYMMDD

```ts:days.ts
import dayjs from "dayjs";

import ja from "dayjs/locale/ja";
//日本に言語設定
dayjs.locale(ja);

/**
 * YYYY年MM月DD日にフォーマットする
 * @param date 日付
 * @returns YYYY年MM月DD日にフォーマットされた日付
 */
export const formatYYYYMMDD = (date: Date | string) => {
  const dateDayjs = dayjs(date);
  return dateDayjs.format("YYYY年MM月DD日");
};

```

:::

::: details highlightByHighlightJs

```ts:highlightByHighlightJs
export const highlightByHighlightJs = (content: string) => {
  const $ = cheerio.load(content);

  $("pre code").each((_, elm) => {
    const result = hljs.highlightAuto($(elm).text());
    $(elm).html(result.value);
    $(elm).addClass("hljs");
  });
  return $.html();
};
```

:::

# まとめ

# 参考文献

- https://maku.blog/p/ny9fmty/
- https://v2.tailwindcss.com/docs/guides/nextjs
- https://github.com/vercel/next.js/blob/canary/examples/with-google-analytics/lib/gtag.js
- https://github.com/vercel/next.js/blob/canary/examples/with-google-analytics/pages/_app.js
- https://nextjs.org/docs/basic-features/script
- https://zenn.dev/aiji42/articles/9a6ab12ab5f6e6
- https://qiita.com/purpleeeee/items/cd9aca1ae735ad678355
- https://github.com/vercel/next.js/blob/canary/examples/with-google-analytics/pages/_app.js
- https://blog.microcms.io/syntax-highlighting-on-server-side/
- https://nextjs.org/docs/basic-features/environment-variables
- https://nextjs-ja-translation-docs.vercel.app/docs/api-routes/introduction
- https://qiita.com/matamatanot/items/1735984f40540b8bdf91
- https://qiita.com/thesugar/items/47ec3d243d00ddd0b4ed
- https://zenn.dev/ria/articles/b709ae94e919c76f814a
