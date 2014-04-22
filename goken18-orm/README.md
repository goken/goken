# Goken vol.18

## 課題

ORM

## 仕様

ORM を作る。

## 観点

ORM のクエリビルダの部分だけ作る。
DSL を Go でやったらどうなるか？

こんなの

```
db.select("username").from("users").where("id = 1")
// "select username from users where id = 1"
```


## 仕様

- DSL から、 SQL(ANSI 99 準拠) を生成
- 結果として、 SQL 文字列を返す
- 対応する文(select, update,,,)、句(where, groupby,,,)は任意
- プレースホルダ対応すること(where id=? とか)
- サンプルを用意すること(最低限以下のテーブルに対する CRUD)

```
users (id number, name varchar, age number, email varchar)
```


## 共通

- この README.md ファイルがある場所を $GOPATH とする
- src 以下に自分のディレクトリを掘る。その下の使い方は自由。
- このリポジトリの master へのコミットは基本一人 1 コミットにまとめる。
- コミット権無い人は 1 コミットにした PR を下さい、当日マージします。
- main.go 的なファイルの先頭にコメントで *実行方法* 、概要、工夫点とかを書く。


## その他

時間内に読めるよう、規模を大きくしすぎないように
その辺は空気を読みましょうw

もし、課題を修正した方が良い場合は、 issue とか PR とかで立ててください。
