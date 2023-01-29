---
title: "NestJSでPrismaを用いて複数のスキーマで同一のテーブル名のテーブルを扱う方法"
emoji: "🦁"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [NestJS, Prisma]
published: true
---

# 概要

:::message alert
今回紹介している方法は 2023 年 1 月 29 日（日）時点で Preview 機能なのでご利用の際は自己責任でお願いいたします。
:::
個人的に開発しているアプリケーションではバッチが複数ありそれらのバッチの起動時間を異なるスキーマで同名のテーブルで持つ必要が出てきました。今回のアプリケーションではバックエンドにNestJSとORMとしてPrismaを用いていました。今回の要望に応えるため、どのような手順を踏んだのかを記事にしました。

この記事で最終的に作成したテーブルの状態を DBeaver で確認したものが以下の画像です。
![DBeaverで複数のスキーマに同一名のテーブルがあることを確認できる画像](/images/prisma-multi-schema/multi-schema-same-table-name.png)

prisma.schema を複数作って Prisma インスタンスを複数作成することで複数のスキーマのテーブル扱えるようにした記事を書いている方はいらっしゃいました。しかし NestJS で Prisma インスタンスをシングルトンとして複数作成する方法がわからなかったので今回の方法でマルチスキーマに対応しました。
（もし今回の Preview 機能を使わなくてもマルチスキーマに対応できる方法があればコメントに書いていただきたいです！）

GitHub のリポジトリはこちらです。
https://github.com/mkt-engr/play-prisma/tree/zenn-2023-0129

## 主に利用したライブラリとそのバージョン

主に利用したライブラリとそのバージョンはこちらです。

|ライブラリ|バージョン|
|---|---|
|@nestjs/cli|^8.0.0|
|prisma|^4.9.0|
|typescript|^4.3.5|

# 結論

schema.prisma を以下のように書いてください。
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

まずは Prisma に複数のスキーマに対応させたいので`client`セクションでその設定を行います。

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

prisma の場合`@@map()`が記載されていない場合は model がそのままテーブル名になります。しかし schema.prisma では同一名のモデルを作成することができません。そこで`@@map()`を使ってマッピングしたいテーブルを明記します。

言わずもがなですが`@@schema()`でマッピングしたいスキーマを記載します。

次のセクションでは

- Docker でデータベースを立てる
- モデルの作成と初期データ投入する
- Prisma を用いてデータベースを操作するサービスを作成する
- 上記のスキーマを用いて update の操作をする

ところまで書きます。結論だけ知りたい方は次以降のセクションを読む必要はないです！

# 環境構築

このセクションでは以下の内容を説明します。

- Docker でデータベースを立てる
- モデルの作成と初期データ投入
- Prisma を用いてデータベースを操作するサービスを作成
- 上記のスキーマを用いて update の操作をする


## Docker でデータベースを立てる

Docker について詳しく説明はしません。利用した docker-compose.yaml は以下の通りです。

```yaml
version: "3.8"
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

docker-compose でデータベースを立てます。

```sh
docker-compose up
```

docker-compose をバックグラウンドで実行したい場合は`-d`をつけて実行してください。

## マイグレーションの実行と初期データ投入する

マイグレーションを実行を行いテーブルを作成し、作成されたテーブルに初期データを投入します。
### マイグレーションの実行
結論でも書いた通りschema.prismaでモデルを作成します。
```js:prisma/schema.prisma
generator client {
  provider        = "prisma-client-js"
  previewFeatures = ["multiSchema"] //複数スキーマを扱う
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
  schemas  = ["schema01", "schema02"] //スキーマの種類の設定
}

model BatchExecutionTime01 {
  id       Int      @id @default(autoincrement())
  updateAt DateTime @default(now())

  @@map("batch_execution_time") //batch_execution_timeテーブルにマッピングする
  @@schema("schema01") //schema01スキーマにマッピングする
}

model BatchExecutionTime02 {
  id       Int      @id @default(autoincrement())
  updateAt DateTime @default(now()) @updatedAt

  @@map("batch_execution_time") //batch_execution_timeテーブルにマッピングする
  @@schema("schema02") //schema02スキーマにマッピングする
}
```
モデルが作成できたら以下のコマンドを実行してマイグレーションを実行します。
```
npx prisma migrate dev --name "init"
```

このコマンドは以下の3つのことを実行します。

- schema.prismaをもとにマイグレーションを実行するために必要なSQLを`prisma/migrations`ディレクトリの下に保存する
- `prisma/migrations`ディレクトリの下に保存されたSQLを実行する
- Prisma Clientの作成
### 初期データ投入
初期データを投入するための実行ファイルを作成します。

```
touch prisma/seed.ts
```

次に作成したseedファイルを以下のようにします。
以下のコードを実行することででスキーマが「schema01」の「batch_execution_time」テーブルとスキーマが「schema02」の「batch_execution_time」テーブルにデータがupsertされます。
```ts:prisma/seed.ts
import { PrismaClient } from '@prisma/client';

// initialize Prisma Client
const prisma = new PrismaClient();
async function main() {
  const time01 = await prisma.batchExecutionTime01.upsert({
    where: { id: 1 },
    update: {},
    create: {
      id: 1,
    },
  });
  const time02 = await prisma.batchExecutionTime01.upsert({
    where: { id: 1 },
    update: {},
    create: {
      id: 1,
    },
  });
  console.table({ time01, time02 });
}

// execute the main function
main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    // close Prisma Client at the end
    await prisma.$disconnect();
  });

```

このseedファイルを実行するためのコマンドをpackage.jsonに記載します。
```json:package.json
"scripts": {
  // ...
},
//...
,
"prisma": {
  "seed": "ts-node prisma/seed.ts"
}
```

最後に以下のコマンドを実行してテーブルにデータが挿入されていたらOKです。
```
npx prisma db seed
```
## Prisma を用いてデータベースを操作するサービスを作成する

まずは Nestjs のテンプレートを作成します。

```bash
npx @nestjs/cli new .
```

まず Prisma CLI をインストールします。Prisma CLI を使ってマイグレーションなどを行うことができます。

```bash
npm install -D prisma
```

次に NestJS のテンプレート内で Prisma の初期設定を行います。

```bash
npx prisma init
```

これで`prisma`ディレクトリが作成され、そのディレクトリの中に`schema.prisma`が作成されます。

さらに`.env`ファイルが作成されデータベースの接続情報が記載されます。この接続情報は docker-compose.yaml に書かれたデータベースのユーザーやパスワードに書き換えてください。

```.env
DATABASE_URL="postgresql://pguser:pgpassword@localhost:5432/db-with-prisma"
# DATABASE_URL="postgresql://＜POSTGRES_USER＞:＜POSTGRES_PASSWORD＞@localhost:＜ports＞/＜POSTGRES_DB＞"
```

次にPrisma ServiceとPrisma Module を作成します。Prisma Service では Prisma クライアントの生成（シングルトンとして扱う）とデータベースの接続を行います。これらのファイルは以下のコマンドで生成します。Prisma Module
はNestJSのプロジェクト全体でPrisma Serviceを共有する役割を担います。

```
npx nest generate module prisma
npx nest generate service prisma
```

Prisma Serviceは以下のように書いてください。
```ts:src/prisma/prisma.service.ts
import { INestApplication, Injectable } from '@nestjs/common';
import { PrismaClient } from '@prisma/client';

@Injectable()
export class PrismaService extends PrismaClient {
  async enableShutdownHooks(app: INestApplication) {
    this.$on('beforeExit', async () => {
      await app.close();
    });
  }
}
```
こちらは`PrismaClient`を継承して`PrismaService`を作成しています。追加しているメソッドはGraceful shutdown（正常に処理を終了し、データの整合性を保つためのプロセス）を行うために記載しています。

こちらで作成したメソッドをmain.tsで実行します。
```ts:src/main.ts
import { PrismaService } from './prisma/prisma.service';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  const prismaService = app.get(PrismaService);
  await prismaService.enableShutdownHooks(app);

  await app.listen(3000);
}
bootstrap();
```
Graceful shutdownについて詳しくはこちらのGitHubのIssueをご覧ください。

https://github.com/prisma/prisma/issues/2917#issuecomment-708340112


次に`PrismaService`をアプリケーション全体で共有するために`prisma.module.ts`でexportします。

```ts:src/prisma/prisma.module.ts
import { Module } from '@nestjs/common';
import { PrismaService } from './prisma.service';

@Module({
  providers: [PrismaService],
  exports: [PrismaService], //PrismaServiceをexport
})
export class PrismaModule {}
```
## 異なるスキーマで同一のテーブル名のデータを操作する
バッチの実行時間を記録するコードを実行するためのエンドポイントをNest CLIを用いて作成します。

```
npx nest generate resource
```
これを実行するといくつか質問をされるので以下のように回答してください。今回はバッチの実行時刻を記録する機能を開発している想定なので名前は`batch`としておきます。
1. `What name would you like to use for this resource (plural, e.g., "users")?` batch
2. `What transport layer do you use?` REST API
3. `Would you like to generate CRUD entry points?` Yes

これでコントローラー、サービス、モジュールが作成されます。



次にさきほどexportした`PrismaModule`を`BatchModule`で使えるように以下のようにimportします。

```ts:src/batch/batch.module.ts
import { Module } from '@nestjs/common';
import { BatchService } from './batch.service';
import { BatchController } from './batch.controller';
import { PrismaModule } from 'src/prisma/prisma.module';

@Module({
  controllers: [BatchController],
  providers: [BatchService],
  imports: [PrismaModule], //さきほどexportしたPrismaModuleをimport
})
export class BatchModule {}

```

次にバッチの実行時間を更新するためのエンドポイントを2つ作成します。
|エンドポイント|HTTPメソッド|役割|
|---|---|---|
|/batch/01|POST|スキーマ「schema01」のテーブル「batch_execution_time」を更新する|
|/batch/02|POST|スキーマ「schema02」のテーブル「batch_execution_time」を更新する|

```ts:src/batch/batch.controller.ts
@Controller('batch')
export class BatchController {
  constructor(private readonly batchService: BatchService) {}

  @Post('/01')
  upsertSchema01BatchExecutionTime() {
    return this.batchService.upsertSchema01BatchExecutionTime();
  }

  @Post('/02')
  upsertSchema02BatchExecutionTime() {
    return this.batchService.upsertSchema02BatchExecutionTime();
  }
}
```

次にPrismaをつかってデータベースの操作を行うサービスを作成します。

```ts:src/batch/batch.service.ts
@Injectable()
export class BatchService {
  constructor(private prisma: PrismaService) {}

  //バッチ1実行したときに実行時間を記録
  upsertSchema01BatchExecutionTime() {
    return this.prisma.batchExecutionTime01.upsert({
      where: { id: 1 },
      update: { updateAt: new Date() },
      create: {},
    });
  }

  //バッチ2実行したときに実行時間を記録
  upsertSchema02BatchExecutionTime() {
    return this.prisma.batchExecutionTime01.upsert({
      where: { id: 1 },
      update: { updateAt: new Date() },
      create: {},
    });
  }
}
```

`PrismaService`には型が効いているので`this.prisma.b`とVS Codeで入力すると以下のようにサジェストされます。
![this.prisma.bとしたときにVSCodeで表示されるサジェスト](/images/prisma-multi-schema/vscode-prisma-suggest.png)

この「batchExecutionTime01」と「batchExecutionTime02」はどこからきているかというとschema.prismaのモデル名から来ています。
```js
model BatchExecutionTime01 {
  id       Int      @id @default(autoincrement())
  //省略
}
```

PostmanなどのAPIクライアントで作成したAPIをキックして以下のようなレスポンスが返ってきたらOKです。

![PostmanでAPIをキックする](/images/prisma-multi-schema/execute-upsert-batch-at-postman.png)

# まとめ

schema.prismaに以下のような記述をすれba
複数のスキーマで同一のテーブル名を使うことができる。

- 複数のスキーマを使う
```js
generator client {
  provider        = "prisma-client-js"
  previewFeatures = ["multiSchema"] //複数スキーマを扱う
}
```

- スキーマの指定
```
datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
  schemas  = ["schema01", "schema02"] //スキーマの種類の設定
}
```
- モデルとテーブルのマッピング
```js
model BatchExecutionTime01 {
  id       Int      @id @default(autoincrement())
  updateAt DateTime @default(now())

  @@map("batch_execution_time") //batch_execution_timeテーブルにマッピングする
  @@schema("schema01") //schema01スキーマにマッピングする
}
```
# マルチスキーマの最新情報について

マルチスキーマの最新情報についてはこちらの GitHub の Issue をご確認ください。
https://github.com/prisma/prisma/issues/15077

# 参考文献

https://www.prisma.io/blog/nestjs-prisma-rest-api-7D056s1BmOL0
https://www.prisma.io/docs/guides/database/multi-schema
https://docs.nestjs.com/recipes/prisma#issues-with-enableshutdownhooks