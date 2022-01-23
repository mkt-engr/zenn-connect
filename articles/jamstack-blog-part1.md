---
title: "Next.js × TypeScript × microCMS × Tailwind CSSでJamStackなブログを作ってみた"
emoji: "🕌"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: ["Next", "TypeScript", "microCMS", "TailwindCSS"]
published: false
---

# 概要

こんにちは。まっきんとっしゅです([@mkt_phys](https://twitter.com/mkt_phys))。今回は Next.js と microCMS を使って Jamstack なブログを作ってみました。リポジトリは[こちら](https://github.com/mkt-engr/jamstack_blog)です(README.md ちゃんと書きます..)。

## 作ったもの

作成したブログの URL は[こちら](https://mkt-blue.com/)です。ほんとにシンプルな構成でページとしては記事一覧ページと記事詳細ページだけです笑。具体的には以下の通りです。検索機能とか、タグとかは作ってないです(作ります)。

- 記事一覧ページ
- 記事詳細ページ
- シンタックスハイライト
- SSG+ISR+CSR
- Google Analytics 連携
- 独自ドメイン

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

<!-- TODO:表にする -->

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

[こちら](https://maku.blog/p/ny9fmty/)のサイトに既存のプロジェクトに TypeScript を導入する方法が記載してありました。

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

[公式サイト](https://v2.tailwindcss.com/docs/guides/nextjs)に手順があるのでそれを引用します。

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

## ブログサイト全体のレイアウトを決める (`Layout.tsx`)

## 記事一覧ページ(`pages/index.tsx`)

NextPage 型は

> `Page` type, use it as a guide to create `pages`.

だそうです。pages ディレクトリ下で使うファイルだということを明示するために使う型のようです。VFC を使ってもエラーは出なかったのでどちらでも問題はないのかなと思います。。ちゃんとした意味があった場合は教えてください笑

## 記事詳細ページ(`pages/index.tsx`)

## 作った機能その３

# まとめ
