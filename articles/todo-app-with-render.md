---
title: "Herokuの代替となるPaaS「Render」で簡単なToDoアプリを作って遊んでみた"
emoji: "🐈"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: ["Render", "TypeScript", "Nextjs", "Nestjs", "Postgres"]
published: false
---

# はじめに

:::message
この記事は[Sun\* Advent Calendar 2022](https://adventar.org/calendars/8211)の 9 日目の記事です。
:::

株式会社 Sun Asterisk に所属する森真輝人です。社内では Web エンジニアとして働いていて Heroku を用いたシステム開発を行っています。

# 本記事の概要

Heroku はアプリケーションをホスティングする PaaS として有名ですが残念なことに 2022 年 11 月 28 日で無料プランが終了してしまいました。現在の仕事には影響はないものの今後どんな PaaS が来るのかなといろいろ探しているところ Render を見つけました。

この記事では簡単な ToDo アプリを作成して Render にデプロイする中でメリット・デメリットがお話しできればと思います。

## 作成した ToDo アプリ

アプリは[こちら](https://example-service-next.onrender.com/)にデプロイしました。アプリの見た目はこんな感じです（ほんとに簡素ですね）。その代わりフロントエンド、バックエンド、データベースはすべて無料でホスティングできています。無料プランで作成したので 15 分アクセスがないとサーバが停止します。停止した状態でアクセスするとサーバの再起動に時間がかかってなかなかアプリが表示されないので気を付けてください。
![ToDoアプリの画面](/images/todo-app-with-render/todo-app.png)

ToDo アプリとしての機能は

- タスク追加（10 件まで）
- タスク削除（物理削除）
- タスク完了と未完了の切り替え

と少なめですがぜひ遊んでみてください（エラーハンドリングはしてないので不具合が起こっても我慢してください）。

::: details ToDo アプリを実際に動かしてみた様子（GIF）

![実際に動かしてみた](/images/todo-app-with-render/todo-app-gif.gif)

:::

## 利用した技術スタック

ToDo アプリを作成するにあたって主に利用したライブラリとそのバージョンは以下の通りです。

- Node.js : v16.13.1
- バックエンド（リポジトリは[こちら](https://github.com/mkt-engr/render-app-nest)）
  | ライブラリ | バージョン |
  | ---------------- | ---------- |
  | @nestjs/cli | ^8.0.0 |
  | prisma | ^4.7.1 |
  | typescript | ^4.3.5 |

- フロントエンド（リポジトリは[こちら](https://github.com/mkt-engr/render-app-next)）
  | ライブラリ | バージョン |
  | ---------------- | ---------- |
  | next | 13.0.6 |
  | react | 18.2.0 |
  |typescript|4.9.4|
  |@mui/icons-material|^5.10.16|

- データベース
  | 種類 | バージョン |
  | ---------------- | ---------- |
  | PostgreSQL | 14 |

PostgreSQL のバージョンを 14 にしたのは Heroku で対応しているバージョンが 14 だったので Render でもなんとなく 14 にしました。Render では 15 まで使うことができます。

余談ですが MUI のバージョンが v5 になってからはじめて触りましたが v4 より使いやすくなっているような気がしました。

# Render とは

:::message
今回紹介している PaaS は「Render」なのだろうか。それとも「render」なのだろうか。「Render」の方がかっこいいので本記事では「Render」と書いてます。
:::

Render は Web アプリケーション、ウェブサイト、データベース、クーロンなどを作成することができるプラットフォームです。Render は Ruby、Java、Node.js、PHP などさまざまな言語に対応しています。

公式サイトで[Heroku VS Render](https://render.com/render-vs-heroku-comparison)なるページがありそこでは

> We’ve built Render to help developers and businesses avoid the cost and inflexibility traps of legacy Platform-as-a-Service solutions like Heroku. Our customers often tell us Render is what Heroku could have been. This page explains why so many former Heroku customers consider Render to be the best Heroku alternative.

日本語訳としては

> 私たちは、開発者や企業が Heroku のようなレガシーな Platform-as-a-Service ソリューションのコストや柔軟性の問題を回避できるように、Render を開発しました。私たちの顧客はしばしば、Render は Heroku がなり得たかもしれないものだと言います。このページでは、なぜ多くの元 Heroku ユーザーが Render を最高の Heroku の代替品と考えるのかを説明します。

となかなか挑戦的なことを言っています笑。個人的な感想としては正直 Heroku の方が使いやすい感はありますが Render にしかない機能もあります。Render にしかない機能も含め Render を使うメリットをいくつか紹介します。

## Render のメリット

//TODO:いくつ？
Render を使うメリットをいくつか紹介します。

### GitHub との連携機能

GitHub（や GitLab）と連携して特定のブランチ（main ブランチなど）へのプッシュをトリガーに自動デプロイできます。これに関しては Render に限らず Heroku や Netlify などでも存在する機能ですね。Render ではダッシュボードから自動デプロイをオフにできます。

おそらく Render 特有と思われますがおもしろい連携機能が 2 つありました。

1. コミットメッセージに特定の文言を入れることで自動デプロイをスキップできる

2. ダッシュボードから特定のコミットの状態にロールバックできる

まず 1 つ目の「コミットメッセージに特定の文言を入れることで自動デプロイをスキップできる」の方ですが Git のコミットメッセージに` [<KEYWORD> skip]` か `[skip <KEYWORD>] `が含まれていると自動デプロイがスキップされます。`KEYWORD`は`render`、`deploy`、`cd`の 3 つのうちの 1 つを使ってください。たとえばコミットメッセージを`[render skip] Update README`として GitHub にプッシュすれば自動デプロイはスキップされます。
自動デプロイがスキップされたかどうかはダッシュボードからも確認できます。`Deploy skipped for ...`とありますね。
![自動デプロイのスキップ](/images/todo-app-with-render/skip-auto-deploy.png)
README.md だけを更新してデプロイの必要がない時に使えそうですね。

次に 2 つ目の「ダッシュボードから特定のコミットの状態にロールバックできる」の方ですが`Rollback to this deploy`をクリックすればロールバックできます。間違えてデプロイしてしまった時に特定の時点の状態へ戻したい時に使えますね。
![特定のコミットの状態にロールバック](/images/todo-app-with-render/rollback-deploy.png)

### 無料プランでも PostgreSQL に IP 制限ができる

Heroku で DB に IP 制限をかけようと思うと割と高額なエンタープライズプランを契約しないといけません。Render では無料プランでも IP 制限ができるのは嬉しいですね。
![DBへのアクセス制限](/images/todo-app-with-render/db-access-control.png)
IP 制限もダッシュボードから簡単に行えます。自分の環境からのみアクセスできるようにしたい時は`Use my IP address`をクリックすれば自動で入力されるので地味に嬉しい機能です。

### HTTP2 で通信しているとか書く？

TODO:HTTP2 で通信しているとか書く？

### a

- render.yaml で IaC として管理できる

## Render のデメリット

デメリットもいくつかあるのでこちらで紹介します。

### デプロイが割と遅い

無料プランの場合だけかもしれませんがデプロイが割と遅いです。ダッシュボードに「デプロイは遅いですか？アップグレードしますか？」的な文章が目に入ってきます笑。今回の作成した Next、Nest ともにデプロイ完了まで 3 分〜5 分くらいかかります。同じくらいのサイズのアプリを Heroku にデプロイする際は無料プランでも 1 分もかかってなかったのでここは Heroku に軍配が上がります。

### 無料プランのインスタンスはアクセスがないと停止する

無料プランの Web Service（アプリをホスティングするサーバ）の説明で

> Web Services on the free instance type are automatically spun down after 15 minutes of inactivity. When a new request for a free service comes in, Render spins it up again so it can process the request.
> This can cause a response delay of up to 30 seconds for the first request that comes in after a period of inactivity.

とありました。日本語訳は

> フリーインスタンスタイプの Web サービスは、15 分間使用されないと自動的にスピンダウンされます。無料サービスに対する新しいリクエストが来ると、Render はリクエストを処理できるように再びスピンアップします。
> このため、一定期間使用されていない状態で最初に来たリクエストに対して、最大で 30 秒の応答遅延が発生する可能性があります。

です。15 分アクセスがないとそのアプリは止まっちゃうってことですね。リクエストが来たら再起動して最大 30 秒の応答遅延があると書いてますがそんなに短くないです。少なくとも数分は待つ必要がありました。

### npm が使えない(かもしれない)

デメリットでもないですが Render のダッシュボードで指定するビルドコマンドとスタートコマンドに npm が使えません。
ビルドコマンドに関するドキュメントには例として yarn が使われていましたが npm が使えないとは書いてなかったのでもしかしたら npm も使えるかもしれません。

# データベース編(Postgres)

# バックエンド編(Nestjs)

# フロントエンド編(Nextjs)

# まとめ

# 参考

自動デプロイスキップの話
https://render.com/docs/deploys

無料の WebService は 15 分使用されないとインスタンスが停止する件
https://render.com/docs/free#free-web-services
