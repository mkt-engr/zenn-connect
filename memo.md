#

https://zenn.dev/zenn/articles/markdown-guide

🎉 Done!
早速コンテンツを作成しましょう

👇 新しい記事を作成する
$ zenn new:article

👇 新しい本を作成する
$ zenn new:book

👇 投稿をプレビューする
$ zenn preview

## セットアップ方法

https://zenn.dev/zenn/articles/install-zenn-cli

## 執筆方法

https://zenn.dev/zenn/articles/zenn-cli-guide

```sh
$ npx zenn new:article --slug what-is-slug
# => articles/what-is-slug.md`が作成される
```

詳細

```sh
npx zenn new:article --slug 記事のスラッグ --title タイトル --type idea --emoji ✨
```

## 画像の挿入方法

参考：https://zenn.dev/zenn/articles/deploy-github-images
`zenn-connect/images`の下に画像を置く

- 画像の指定方法
  ![](/images/USER_type.png)
  絶対パスで行う。

# テスト

ソースツリーでコマンドを使えるか
