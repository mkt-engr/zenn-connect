---
title: "【TypeScript】Axiosのリクエストとレスポンスに型をつける"
emoji: "🗂"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: ["Axios", "TypeScript", "React"]
published: false
---

# 概要

React×TypeScript のプロジェクトで Axios を使った場合のリクエストとレスポンスの型についてまとめます。最終的にやることは フェイク API の [JSONPlaceholder](https://jsonplaceholder.typicode.com/) からユーザー情報を受けとりそれらの情報を表示することです。最終的なコードの全容は「まとめ」に書いてあるので急ぎの方はそちらを見てください！

よくある記事としてはプレーンな JavaScript で書いた

```jsx
axios.get(url);
```

に TypeScript で型をつける方法です。しかし私は

```jsx
const options = {
  url,
  method: "GET",
};
axios(options);
```

と書いたものに型をつけたいと思ってました。

どこか間違いがあったりコードや文章の書き方のアドバイスがありましたらお気軽にコメントしていただければうれしいです！

# 事前準備

まずは Create React App でプロジェクトを作成します。

```sh
npx create-react-app . --template typescript
```

メインである Axios をインストールします。

```sh
npm i axios
```

# リクエストの型

## `axios(options)`の`options`の型

`options`には以下のように`AxiosRequestConfig`という型を当てます。`AxiosRequestConfig`は`axios`から import します。

```tsx
import axios, { AxiosRequestConfig } from "axios";

const url = "https://jsonplaceholder.typicode.com";

const options: AxiosRequestConfig = {
  url: `${url}/users`,
  method: "GET",
};
```

## `AxiosRequestConfig`の型をつける必要わかった経緯

ここは型をつける理由がわかった経緯を書いているので読まなくても型をつける方法はわかります。

最初は何も考えずに以下のように書いてました。

```tsx
const options: AxiosRequestConfig = {
  url: `${url}/users`,
  method: "GET",
};
axios(options);
```

この時 VSCode 上で`options`のところに以下のアラートが出てきました。

```sh
No overload matches this call.
  Overload 1 of 2, '(config: AxiosRequestConfig<any>): AxiosPromise<any>', gave the following error.
    Argument of type '{ url: string; method: string; }' is not assignable to parameter of type 'AxiosRequestConfig<any>'.
  Overload 2 of 2, '(url: string, config?: AxiosRequestConfig<any> | undefined): AxiosPromise<any>', gave the following error.
    Argument of type '{ url: string; method: string; }' is not assignable to parameter of type 'string'.ts(2769)
```

`No overload matches this call.`の意味はわかりませんでしたが（誰か教えてください）、ざっくり意訳してみると「型`{ url: string; method: string; }`が型`AxiosRequestConfig<any>`に割り当てられない」という感じです。`options`型をあてていないので型推論により`{ url: string; method: string; }`が割り当てられたようです。なので`options`に型`any`をつけても問題は解消されます。

# レスポンスの型

# まとめ

# 参考文献
