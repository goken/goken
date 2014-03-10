# Goken vol.16

## 課題

grep clone


## 共通

- この README.md ファイルがある場所を $GOPATH とする
- src 以下に自分のディレクトリを掘る。その下の使い方は自由。
- このリポジトリの master へのコミットは基本一人 1 コミットにまとめる。
- コミット権無い人は 1 コミットにした PR を下さい、当日マージします。
- main.go 的なファイルの先頭にコメントで *実行方法* 、概要、工夫点とかを書く。


## 仕様

```
$ go run grep.go mattn CONTRIBUTORS
Yasuhiro Matsumoto <mattn.jp@gmail.com>
```

arg1: 検索単語
arg2: 検索対象のファイル
out: 検索単語の含まれる行全て

arg1 は *単一の単語* のみとし、正規表現やスペース区切りの複数の単語などは想定しない。
arg2 は *単一のファイル* のみとし、ワイルドカードやディレクトリ指定は想定しない。
その他 grep の持つオプションなどは実装しない。


## 観点

- パフォーマンス!!
- 正規表現の取り回し
- ファイルの扱い
- readline の実装

## ベンチ

今回は、この仕様の中で速度を競ってみましょう。
コマンドとしての速度が欲しいので、測定は test コマンドを用います。
以下の sh で行うので全員 goken-grep という名前で
Mac 64bit 用バイナリを吐いて入れておいて下さい。
(当日だれかのマシンで全員のものを流して比較します)

```sh
for i in `seq 1 1000`
do
  ./goken-grep mattn CONTRIBUTORS 1>/dev/null 2>/dev/null
done
```

```sh
$ time ./test.sh
./test.sh  0.87s user 1.43s system 95% cpu 2.425 total
```

## その他

時間内に読めるよう、規模を大きくしすぎないように
その辺は空気を読みましょうw

もし、課題を修正した方が良い場合は、 issue とか PR とかで立ててください。
