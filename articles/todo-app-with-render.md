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

## 技術スタック

アプリは[こちら](https://example-service-next.onrender.com/)にデプロイしました。ToDo アプリとしての機能は

- タスク追加
- タスク削除
- タスク完了・未完了の切り替え

と少なめですがぜひ遊んでみてください（エラーハンドリングはしてないので不具合が起こっても我慢してください）。

ToDo アプリを作成するにあたって主に利用したライブラリとそのバージョンは以下の通りです。

- Node.js : v16.13.1
- バックエンド
  | ライブラリ | バージョン |
  | ---------------- | ---------- |
  | @nestjs/cli | ^8.0.0 |
  | prisma | ^4.7.1 |
  | typescript | ^4.3.5 |

- フロントエンド
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

バックエンド、フロントエンドのソースをアップした GitHub のリポジトリはこちらです。

フロントエンド：https://github.com/mkt-engr/render-app-next
バックエンド：https://github.com/mkt-engr/render-app-nest

# Render とは

- Render の説明
- 「Render」なのか「render」なのか
- 「Render は Heroku の上位互換だ」的な発言を見たけど原点を載せる
- メリット
  - GitHub と連携して main ブランチへのプッシュをトリガーに自動デプロイ
  - 無料で
  - 無料で PostgreSQL に IP 制限ができる
    - Heroku では割と高額なエンタープライズプランを契約しないとできない
  - GUI である時点のデプロイにロールバックできる
- デメリット
  - デプロイがめちゃ遅い
  - npm が使えない

[Heroku VS Render](https://render.com/render-vs-heroku-comparison)

# データベース編(Postgres)

# バックエンド編(Nestjs)

# フロントエンド編(Nextjs)

# まとめ
