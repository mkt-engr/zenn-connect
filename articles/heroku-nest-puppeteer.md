---
title: "Heroku上でNest.jsを使ってPuppeteerを動かす"
emoji: "📌"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Heroku, Nestjs, TypeScript, Puppeteer]
published: false
---

sudo rm -rf /Library/Developer/CommandLineTools
sudo xcode-select --install

# 概要

2022 年 11 月 28 日に無料プランが終了します。11 月 28 日以降に以下の内容を再現しようとすると料金が発生しますのでご注意ください。

Heroku 上で NestJS を用いて Puppeteer を動かせるとこまでをこちらの記事で説明します。

最終的な成果物は Puppeteer で「Puppeteer」とグーグル検索した

# Heroku 環境準備

![Herokuアプリ一覧画面](/images/heroku-nest-puppeteer.md/heroku-app-list.png)
右上の「New」をクリックして
![Herokuアプリ作成画面](/images/heroku-nest-puppeteer.md/heroku-create-app.png)
適当な App name を設定して「Create app」をクリックしてください

Heroku CLI を用いてデプロイを行います。

Heroku CLI のインストールがまだの方は[こちら](https://devcenter.heroku.com/ja/articles/heroku-cli)を参考にインストールを行ってください。

```
heroku --version
```

# NestCLI でテンプレート作成&デプロイ

# Puppeteer を動かす&デプロイ

#

# まとめ

# 参考
