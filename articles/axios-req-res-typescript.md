---
title: "【TypeScript】Axiosのリクエストとレスポンスに型をつける"
emoji: "🗂"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: ["Axios", "TypeScript", "React"]
published: true
---

# 概要

React×TypeScript のプロジェクトで Axios を使った場合のリクエストとレスポンスの型についてまとめます。最終的にやることはフェイク API の [JSONPlaceholder](https://jsonplaceholder.typicode.com/) からユーザー情報を 10 件取得してそれらの情報を表示することです。最終的なコードの全容は「コードの全容」に書いてあるので急ぎの方はそちらを見てください！

この記事の中身について説明します。よくある記事としてはプレーンな JavaScript で書いた

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

# リクエスト`axios(options)`の`options`の型

`options`には以下のように`AxiosRequestConfig`という型を当てます。`AxiosRequestConfig`は`axios`から import します。

```tsx
import axios, { AxiosRequestConfig } from "axios";

const url = "https://jsonplaceholder.typicode.com";

const options: AxiosRequestConfig = {
  url: `${url}/users`,
  method: "GET",
};
```

# レスポンスの型

レスポンスには以下のように`AxiosResponse`という型を当てます。`AxiosResponse`は`axios`から import します。本題とはちょっとそれるかもしれませんがエラーの型は`AxiosError`という型を当て、こちらも`axios`から import します。

```tsx
import axios, { AxiosRequestConfig, AxiosResponse, AxiosError } from "axios";
import user from "./user.json"; //レスポンスのJSON(詳しくは補足で)

type USER = typeof user; //画面に表示するユーザー情報の型

const options: AxiosRequestConfig = {
  url: `${url}/users`,
  method: "GET",
};

axios(options)
  .then((res: AxiosResponse<USER[]>) => {
    const { data, status } = res;
    //やりたいことをやる
  })
  .catch((e: AxiosError<{ error: string }>) => {
    // エラー処理
    console.log(e.message);
  });
```

# コードの全容

API 通信は`useEffect`の中で行いました。ユーザー情報の表示の部分もまだ記載していなかったのでここでコードの全容を書きます。

ユーザー情報の配列とレスポンスのステータスは以下のように定義しています。

```tsx
const [users, setUsers] = useState<USER[]>([]);
const [status, setStatus] = useState<number | null>(null);
```

```tsx:AxiosResReqType.tsx
import React, { useState, useEffect } from "react";
import axios, { AxiosRequestConfig, AxiosResponse, AxiosError } from "axios";
import user from "./user.json"; // レスポンスのJSON(詳しくは補足で)
const url = "https://jsonplaceholder.typicode.com";

type USER = typeof user; // 画面に表示するユーザー情報の型

const options: AxiosRequestConfig = {
  url: `${url}/users`,
  method: "GET",
};

const AxiosResReqType: React.FC = () => {
  const [users, setUsers] = useState<USER[]>([]);
  const [status, setStatus] = useState<number | null>(null);

  //API通信を行う箇所
  useEffect(() => {
    axios(options)
      .then((res: AxiosResponse<USER[]>) => {
        const { data, status } = res;
        setUsers(data);
        setStatus(status);
      })
      .catch((e: AxiosError<{ error: string }>) => {
        // エラー処理
        console.log(e.message);
      });
  }, []);

  //ユーザー情報を表示する箇所
  return (
    <div>
      <h1>Axiosのリクエストとレスポンスに型をつける</h1>
      <h2>ステータス:{status}</h2>
      <ul>
        {users.map(({ id, name }) => {
          return (
            <li key={id}>
              {id} : {name}
            </li>
          );
        })}
      </ul>
    </div>
  );
};

export default AxiosResReqType;
```

# まとめ

- リクエスト（`axios(options)`の`options`）に使う型：`AxiosRequestConfig`
- レスポンスに使う型：`AxiosResponse<T>`
  - T には自分の必要な型を書く
  - この記事の場合は`User`型

まとめるとたった 2 つだけでした笑。

リクエストとレスポンスの型には直接関係ないですが以下のように JSON を使って簡単に型を定義できるのは驚きでした。

```tsx
import user from "./user.json";
type USER typeof user
```

冒頭にも書きましたがスペルミス、間違っている箇所、もっとこうしたほうがいいなどのアドバイスがありましたらお気軽にコメントしていただけるとうれしいです！

# 補足

本筋とは外れますが残しておきたかったことをここに書いておきます。読まなくてもこの記事で言いたいことはわかります！

## 補足その 1`AxiosRequestConfig`の型をつける必要わかった経緯

ここは型をつける理由がわかった経緯を書いているので読まなくても型をつける方法はわかります。

最初は何も考えずに以下のように書いてました。

```tsx
const options: AxiosRequestConfig = {
  url: `${url}/users`,
  method: "GET",
};
axios(options);
```

この時 VS Code 上で`options`のところに以下のアラートが出てきました。

```sh
No overload matches this call.
  Overload 1 of 2, '(config: AxiosRequestConfig<any>): AxiosPromise<any>', gave the following error.
    Argument of type '{ url: string; method: string; }' is not assignable to parameter of type 'AxiosRequestConfig<any>'.
  Overload 2 of 2, '(url: string, config?: AxiosRequestConfig<any> | undefined): AxiosPromise<any>', gave the following error.
    Argument of type '{ url: string; method: string; }' is not assignable to parameter of type 'string'.ts(2769)
```

`No overload matches this call.`の意味はわかりませんでしたが（誰か教えてください）、ざっくり意訳してみると「型`{ url: string; method: string; }`が型`AxiosRequestConfig<any>`に割り当てられない」という感じです。`options`型をあてていないので型推論により`{ url: string; method: string; }`が割り当てられたようです。なので`options`に型`any`をつけても問題は解消されます（されてない）。

## 補足その 2`AxiosResponse`の型をつける必要わかった経緯

`options`に型`AxiosRequestConfig`を適用することがわかった後の話です。VS Code 上で`AxiosRequestConfig`を ctrl ＋クリックすると`node_modules/axios/index.d.ts`のファイルが表示されます。ここに型の情報がいろいろ書かれています。

この中にレスポンスに関する型もあると予想し`response`で検索をかけると以下の記述がありました。

```ts
export interface AxiosResponse<T = any, D = any> {
  data: T;
  status: number;
  statusText: string;
  headers: AxiosResponseHeaders;
  config: AxiosRequestConfig<D>;
  request?: any;
}
```

これを見つけた時レスポンスに使う型はこれかなと思いました。その理由は以下の 2 つです。

- レスポンスに含まれる`data`という変数に欲しい情報（今回の場合はユーザー情報）が入ることを把握していた
- `data`はどのようなリクエストをするかによって変数の型は変わるのでジェネリクスで`USER`型をどこかで指定しないといけないのはなんとなく予想していた

よって以下ように`res:AxiosResponse<USER[]>`を書きました。今回はユーザー情報を 10 件取得するので配列にしています。`data`と`status`にもきっちり型が効いてます。

```tsx
axios(options).then((res: AxiosResponse<USER[]>) => {
  const { data, status } = res;
  //やりたいことをやる
});
```

## 補足その 3(user.json と`type USER = typeof user;`に関して)

1 つのユーザー情報には id,name,email などたくさんあるのでそれを 1 つずつ`{"id":number,name:string,"email":string,...}`と 1 つずつ書いていくとかなり手間になります。楽して`USER`の型を定義するタメに JSON と`typeof`を使います。

user.json は[こちらの JSONPlaceholder](https://jsonplaceholder.typicode.com/users)にアクセスして表示される配列の 1 つ目の要素をコピペしたものです。
:::details user.json（ちょっと長め）

```JSON:user.json
{
  "id": 1,
  "name": "Leanne Graham",
  "username": "Bret",
  "email": "Sincere@april.biz",
  "address": {
    "street": "Kulas Light",
    "suite": "Apt. 556",
    "city": "Gwenborough",
    "zipcode": "92998-3874",
    "geo": {
      "lat": "-37.3159",
      "lng": "81.1496"
    }
  },
  "phone": "1-770-736-8031 x56442",
  "website": "hildegard.org",
  "company": {
    "name": "Romaguera-Crona",
    "catchPhrase": "Multi-layered client-server neural-net",
    "bs": "harness real-time e-markets"
  }
}

```

:::

この JSON に対して

```tsx
type USER typeof user
```

とすると簡単に USER 型が得られます。ちなみに VS Code 上で`USER`をホバーすると以下の画像のようなものが表示されます。
![](/images/USER_type.png)
