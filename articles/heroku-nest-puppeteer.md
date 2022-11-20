---
title: "Heroku上でNest.jsを使ってPuppeteerを動かす"
emoji: "📌"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Heroku, Nestjs, TypeScript, Puppeteer]
published: false
---

# 概要

2022 年 11 月 28 日に無料プランが終了します。11 月 28 日以降に以下の内容を再現しようとすると料金が発生しますのでご注意ください。

Heroku 上で NestJS を用いて Puppeteer を動かせるとこまでをこちらの記事で説明します。

最終的な成果物は Puppeteer で「Puppeteer」とグーグル検索した

# Heroku 環境準備

![Herokuアプリ一覧画面](/images/heroku-nest-puppeteer/heroku-app-list.png)
右上の「New」をクリックして
![Herokuアプリ作成画面](/images/heroku-nest-puppeteer/heroku-create-app.png)
適当な App name を設定して「Create app」をクリックしてください

Heroku CLI を用いてデプロイを行います。

Heroku CLI のインストールがまだの方は[こちら](https://devcenter.heroku.com/ja/articles/heroku-cli)を参考にインストールを行ってください。

```
heroku --version
# heroku/7.66.4 darwin-x64 node-v14.19.0
```

のように version が出ればインストールが完了しています。

# Heroku にデプロイ

## NestCLI でテンプレート作成

まずは NestJS のプロジェクトを作成します。お好きなフォルダへ移動して以下のコマンドを入力します。

```
nest new .
```

なお上記コマンドの最後を　.　とすることで現在のフォルダに作成したファイルを展開できます。

```
npm run start:dev
```

で起動し、localhost:3000 で「Hello World!」と表示されたら正しくテンプレートが作成できています。

## Procfile の作成

Heroku 上で起動時にアプリが実行するコマンドを指定するために Procfile を作成します。拡張子は不要なので気を付けてください。

```Procfile:Procfile
web: npm run start:prod
```

このコマンドは package.json に定義されている

```json:package.json
{
  "script": {
    "start:prod": "node dist/main"
  }
}
```

から記述しています。

## Heroku CLI でデプロイ

次に main.ts でアプリが使うポートを環境変数を用いて指定します。

```diff ts:main.ts
async function bootstrap() {
  const app = await NestFactory.create(AppModule);
-  await app.listen(3000);
+  await app.listen(process.env.PORT || 3000);
}
bootstrap();
```

次に作成したテンプレートを Heroku にデプロイするために Heroku にログインします。

```
heroku login
```

パスワードの入力を求められるのでパスワードを入力してください。入力するパスワードはブラウザで Heroku の管理画面へログインする時に使ったものと同じものです。

次に Git リポジトリを作るために

```
git init
heroku git:remote -a heroku-nest-puppeteer
# heroku git:remote -a <作成したプロジェクト名>
```

とします。これで自分のプロジェクトと Heroku が連携されました。このプロジェクトをデプロイするために以下のコマンドを入力します。デプロイ中は少々長いログが出ます。

```
git add .
git commit -m "最初のコミット"
git push heroku main
```

正しくデプロイされているか確認するために以下のコマンドを入力してデプロイ先の URL をブラウザで開いてみます。

```
heroku open
```

これでブラウザに「Hello World!」と表示されていれば正しくデプロイされています！

# Puppeteer を Heroku 上で動かす

## Heroku で Puppeteer の buildpack を追加

Buildpack とは[公式サイト](https://jp.heroku.com/elements/buildpacks)によると

> Heroku Buildpack は、Heroku でアプリケーションをコンパイルするために使用するオープンソーススクリプトを集めたもの

ということだそうです。

つまり Heroku で Puppeteer を使うには NPM で Puppeteer をインストールするだけではダメで buildpack を追加する必要があります。buildpack を追加するには

```
heroku buildpacks:add jontewks/puppeteer
```

としてください。追加で日本語対応をしている Heroku の buildpack を入れる場合は

```
heroku buildpacks:add https://github.com/gnuletik/puppeteer-heroku-buildpack-fonts
```

としてください。管理画面から buildpack を追加する場合は
![Heroku設定画面](/images/heroku-nest-puppeteer/heroku-settings-1.png)
の Setting から
![Heroku設定画面](/images/heroku-nest-puppeteer/heroku-add-buildpack.png)
に buildpack の URL（https://github.com/gnuletik/puppeteer-heroku-buildpack-fonts)を入力して保存してください。
こちらも日本語対応の buildpack は廃止されたので使わないように気を付けてください。
https://github.com/CoffeeAndCode/puppeteer-heroku-buildpack

## Puppeteer のエンドポイントを作成

まず Puppeteer のためのエンドポイントを作ります。

```
npx nest generate resource
```

いろいろ聞かれるので

```
? What name would you like to use for this resource (plural, e.g., "users")?
# puppeteer
? What transport layer do you use?
# REST API
? Would you like to generate CRUD entry points?
# Yes
```

のように答えてください。

次に環境変数を使うために

```
npm install @nestjs/config
```

して

```diff ts:app.module.ts
import { ConfigModule } from '@nestjs/config';
@Module({
-  imports: [PuppeteerModule],
+  imports: [PuppeteerModule, ConfigModule.forRoot()],
 controllers: [AppController],
 providers: [AppService],
})
```

app.module.ts に追加する。（https://qiita.com/piggydev/items/e76adcc3a65364f98537）

```
npm i puppeteer
```

TODO
Puppeteer で Google 検索するコードを書く

#

# まとめ

# 参考
