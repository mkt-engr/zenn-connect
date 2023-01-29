---
title: "Prismaで複数のスキーマを扱う方法"
emoji: "🦁"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [NestJS,Prisma]
published: false
---

# 概要
:::message alert
今回紹介している方法はPreview機能なのでご利用の際は自己責任でお願いいたします。
:::
NestJSでバッチが複数ありそれらのバッチの起動時間を異なるスキーマで同名のテーブルで持つ必要が出てきた

最終的に作成したテーブルの状態をDBeaverで確認したものが以下の画像です。
![DBeaverで複数のスキーマに同一名のテーブルがあることを確認できる画像](/images/prisma-multi-schema/multi-schema-same-table-name.png)

prisma.schemaを複数作ってPrismaインスタンスを複数作成することで複数のスキーマのテーブル扱えるようにした記事を書いている方はいらっしゃいました。しかしNestJSでPrismaインスタンスをシングルトンとして複数作成する方法がわからなかったので今回の方法でマルチスキーマに対応しました。
（もし今回のPreview機能を使わなくてもマルチスキーマに対応できる方法があればコメントに書いていただきたいです！）



GitHubのリポジトリはこちらでブランチは
TODO:ブランチ名を載せる
です。
https://github.com/mkt-engr/play-prisma

## 利用したライブラリとそのバージョン
TODO:書く


# 結論
schema.prismaを以下のように書いてください。
重要な行はハイライトしています。
```diff js:src/prisma/schema.prisma
generator client {
  provider        = "prisma-client-js"
+  previewFeatures = ["multiSchema"] //複数スキーマを扱う
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
+  schemas  = ["schema01", "schema02"] //スキーマの種類の設定
}

model BatchExecutionTime01 {
  id       Int      @id @default(autoincrement())
  updateAt DateTime @default(now())

+  @@map("batch_execution_time") //batch_execution_timeテーブルにマッピングする
+  @@schema("schema01") //schema01スキーマにマッピングする
}

model BatchExecutionTime02 {
  id       Int      @id @default(autoincrement())
  updateAt DateTime @default(now()) @updatedAt

+  @@map("batch_execution_time") //batch_execution_timeテーブルにマッピングする
+  @@schema("schema02") //schema02スキーマにマッピングする
}

```

まずはPrismaに複数のスキーマに対応させたいので`client`セクションのところでその設定を行います。
```js
generator client {
  provider        = "prisma-client-js"
  previewFeatures = ["multiSchema"] //複数スキーマを扱う
}
```

次に利用するスキーマ名を`db`セクションに記載します。今回は「schema01」と「schema02」を設定します。
```js
datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
  schemas  = ["schema01", "schema02"] //スキーマの種類の設定
}
```

最後にマッピングしたいスキーマとテーブル名を`model`セクションに記載します。

```js
model BatchExecutionTime01 {
  id       Int      @id @default(autoincrement())
  updateAt DateTime @default(now())

  @@map("batch_execution_time") //batch_execution_timeテーブルにマッピングする
  @@schema("schema01") //schema01スキーマにマッピングする
}
```

prismaの場合`@@map()`が記載されていない場合はmodelがそのままテーブル名になります。しかしschema.prismaでは同一名のモデルを作成することができません。そこで`@@map()`を使ってマッピングしたいテーブルを明記します。

言わずもがなですが`@@schema()`でマッピングしたいスキーマを記載します。

次のセクションでは
- Dockerでデータベースを立てる
- NestJSのテンプレートを作成する
- 上記のスキーマを用いてupdateの操作をする
  
ところまで書きます。結論だけ知りたい方はこのセクションまで読めば問題ないです！

# TODO:
##  Dockerでデータベースを立てる
Dockerについて詳しく説明はしません。利用したdocker-compose.yamlは以下の通りです。
```yaml
version: '3.8'
services:

  postgres:
    container_name: play-prisma
    image: postgres:15
    restart: always
    environment:
      POSTGRES_USER: pguser
      POSTGRES_PASSWORD: pgpassword
      POSTGRES_DB: db-with-prisma
    volumes:
      - postgres:/var/lib/postgresql/data
    ports:
      - 5432:5432

volumes:
  postgres:
```

docker-composeでデータベースを立てます。
```sh
docker-compose up -d
```
docker-composeをバックグラウンドで実行したい場合は`-d`をつけて実行してください。
##  NestJSのテンプレート作成
## 異なるスキーマで同一のテーブル名のデータを操作する


# 参考文献
https://www.prisma.io/blog/nestjs-prisma-rest-api-7D056s1BmOL0
https://www.prisma.io/docs/guides/database/multi-schema