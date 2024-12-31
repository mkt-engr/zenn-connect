---
title: "Nextjsでnuqsを用いて検索条件をURLのクエリパラメタを型安全に保存する検索機能の実装（サーバーコンポーネントでAPIコールするよ）" # 記事のタイトル
emoji: "😊" # アイキャッチとして使われる絵文字（1文字だけ）
type: "tech" # tech: 技術記事 / idea: アイデア記事
topics: ["Nextjs","nuqs","Server Component"] # タグ。["markdown", "rust", "aws"]のように指定する
published: false # 公開設定（falseにすると下書き）
---

# 概要

URLのクエリパラメタを型安全に扱うことができるライブラリ[nuqs](https://nuqs.47ng.com/)をNextjsで利用し、検索条件をURLのクエリパラメタを型安全に保存する検索機能を作成してみました。

Nextjsで検索機能を実装しようとしたとき、まず最初に浮かぶのが`useState`で検索条件を保持してAPIをコールする手法かと思います。ただ、その場合だと

- リロードすると検索結果がリセットされてしまう
- 検索結果を他人にシェアすることができない

といったデメリットがあります。そこで検索条件をURLのクエリパラメタに保存することでこれらのデメリットを解消することができました。



GitHub のリポジトリは[こちら](https://github.com/mkt-engr/playground-nuqs/tree/main)です。


## 主に利用したライブラリとそのバージョン

主に利用したライブラリとそのバージョンはこちらです。

|ライブラリ|バージョン|
|---|---|
|next|15.1.2|
|nuqs|^2.2.3|
|typescript|^5|

ダミーのAPIには[DummyJson](https://dummyjson.com/docs/products#products-search)を用いました。

# 実装手順

# 結論

# 参考