# Goken vol.15

## 課題

apache bench clone


## 共通

- この README.md ファイルがある場所を $GOPATH とする
- src 以下に自分のディレクトリを掘る。その下の使い方は自由。
- このリポジトリの master へのコミットは基本一人 1 コミットにまとめる。
- コミット権無い人は 1 コミットにした PR を下さい、当日マージします。
- main.go 的なファイルの先頭にコメントで概要とか工夫点とかを書く。


(あとでやりながら色々変える)


## 仕様


```
$ go run main.go -n 100 -c 10 http://example.com
total time: 100 [ms]
average time: 10 [ms]
req per sec: 1000 [#/seq]
```

-n : リクエストの回数
-c : クライアントの数(多重度)
出力: 合計時間と平均時間 rps のみ。

指定された URL に対して、 HTTP/1.1 の GET を投げる。
ヘッダは最小限。テスト, カバレッジは任意。
主眼は、並列リクエストの取り回し。


## 補足(apache bench とは)

Apache HTTP Server に付属するベンチマークツールで、
簡単に言うと、 URL に対して沢山の GET を並列で投げられるものです。
(オプションは多々あり、また似たようなツールも多々あります)

[http://jxck.tumblr.com/post/17479634129/lion-ab](Lion で ab コマンドがエラー)

Apache HTTP Server をインストールすると入りますが、
Ubuntu などの場合は、 apache2-utils を入れればサーバ入れなくても入ります。

```sh
$ sudo apt-get install apache2-utils
```

ちなみにちゃんと動かすとこんな感じ。これの簡易版を作るのが目的です。

```sh
$ ab -n 100 -c 10 http://example.coms is ApacheBench, Version 2.3 <$Revision: 655654 $>
Copyright 1996 Adam Twiss, Zeus Technology Ltd, http://www.zeustech.net/
Licensed to The Apache Software Foundation, http://www.apache.org/

Benchmarking example.com (be patient).....done


Server Software:        ECS
Server Hostname:        example.com
Server Port:            80

Document Path:          /
Document Length:        1270 bytes

Concurrency Level:      10
Time taken for tests:   3.478 seconds
Complete requests:      100
Failed requests:        0
Write errors:           0
Total transferred:      161000 bytes
HTML transferred:       127000 bytes
Requests per second:    28.76 [#/sec] (mean)
Time per request:       347.753 [ms] (mean)
Time per request:       34.775 [ms] (mean, across all concurrent requests)
Transfer rate:          45.21 [Kbytes/sec] received

Connection Times (ms)
              min  mean[+/-sd] median   max
              Connect:      142  179 150.7    158    1232
              Processing:   140  159   8.8    158     196
              Waiting:      138  158   8.9    158     196
              Total:        290  337 151.8    315    1393

Percentage of the requests served within a certain time (ms)
  50%    315
  66%    323
  75%    326
  80%    328
  90%    335
  95%    344
  98%   1387
  99%   1393
 100%   1393 (longest request)
```

## その他

時間内に読めるよう、規模を大きくしすぎないように
その辺は空気を読みましょうw

もし、課題を修正した方が良い場合は、 issue とか PR とかで立ててください。
