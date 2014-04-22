# Goken vol.17

## 課題

assert


## 共通

- この README.md ファイルがある場所を $GOPATH とする
- src 以下に自分のディレクトリを掘る。その下の使い方は自由。
- このリポジトリの master へのコミットは基本一人 1 コミットにまとめる。
- コミット権無い人は 1 コミットにした PR を下さい、当日マージします。
- main.go 的なファイルの先頭にコメントで *実行方法* 、概要、工夫点とかを書く。


## 仕様

いわゆる、 testing に使う assert を作ってみましょう。

- 提供するメソッドの種類
- 引数などのインタフェース
- テストの落とし方
- SetUp(), TearDown() の提供
- assert, expected, shoud などの語彙体系
- go test コマンドで走るようにするか
- testing パッケージと連携するか
- ベンチやカバレッジを提供するか

などなどは、 *すべて自由* です。
仕様はお任せします。自分が使うことを考えて作ってみてください。


参考
- https://github.com/junit-team/junit/wiki/Assertions
- http://nodejs.org/api/assert.html
- http://docs.ruby-lang.org/ja/2.1.0/class/Test=3a=3aUnit=3a=3aAssertions.html


今回は、多少大きくなってもしょうが無いと思います。
実際にテストを走らせてデモができるように、サンプルを作っておいて下さい。


## 観点

- 使いやすさ
- 見た目
- saiias さんが作ったコードのテストを書く。


## その他

時間内に読めるよう、規模を大きくしすぎないように
その辺は空気を読みましょうw

もし、課題を修正した方が良い場合は、 issue とか PR とかで立ててください。
