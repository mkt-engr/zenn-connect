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

## 記事一覧ページ

NextPage 型は

> `Page` type, use it as a guide to create `pages`.

だそうです。pages ディレクトリ下で使うファイルだということを明示するために使う型のようです。VFC を使ってもエラーは出なかったのでどちらでも問題はないのかなと思います。。ちゃんとした意味があった場合は教えてください笑

## 作った機能その３

# まとめ
