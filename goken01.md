Go研 vol.1 まとめ
==================

##参加者

* [Jxck](https://twitter.com/Jxck_)
* [tenntenn](https://twitter.com/tenntenn)
* [manji0112](https://twitter.com/manji0112)
* [hogedigo](https://twitter.com/hogedigo)

##今回の概要

* 開催日：2013年05月15日(水)
* connpass：http://connpass.com/event/2355/
* 発表者：hogedigo
* バージョン：go1.1
* パッケージ：io

##話題に上がった点

### EOF
EOFを表すerrorには、io.EOFとio. ErrUnexpectedEOFがある。io.EOFは期待しているEOFに使用する。一方で、io. ErrUnexpectedEOFは期待しないEOFの場合に使用する。io.EOFについては、error型ではあるがエラーという扱いはあまりしないようだ。

### io.Readerインタフェース
io.Reader.Readメソッドは、n( > 0), io.EOFが返される場合があるため、単にEOFをチェックするだけでループを抜けると最後のnバイト分飛ばしてしまう。

### io.Seeker.Seekメソッドの引数whence
io.Seekerインタフェースは以下のように定義されている。

	type Seeker interface {
    		Seek(offset int64, whence int) (ret int64, err error)
	}

Seekメソッドは、第2引数にwhence（どこから）として、0〜2の値をとる。それぞれの値の意味は以下の通りである。

* 0: ファイルの先頭から
* 1: 現在のオフセットの位置から
* 2: ファイルの末尾から

これらの値は、定数で宣言されていない。そのため、なぜ以下のように定義しないのかという点が議論に挙がった（[参考](http://play.golang.org/p/GqU3Yot0S3)）。

	type Whence int
	const (
		HEAD Whence = iota
		CURRENT
		END
	)

### io.ReaderFromインタフェースとio.WriterToインタフェース

io.ReaderFromインタフェースは、以下のように定義されている。

	type ReaderFrom interface {
    		ReadFrom(r Reader) (n int64, err error)
	}

また、io.WriterToインタフェースは、以下のように定義されている。

	type WriterTo interface {
    		WriteTo(w Writer) (n int64, err error)
	}

名前から想像すると、io.ReaderFromは読み込み処理を行なうように思える。しかし、実際には引数で渡されたio.Readerから読み込みを行い、その後何かしらの処理を行なうというインタフェースである。例えば、以下のような、io.Copy関数ではこのインタフェースを実装していると期待されているのは、第1引数のdstであり、これはio.Writerインタフェースを実装している。つまり、io.ReaderFromインタフェースは、io.WriterなどのWriterが書き込むデータの提供元を保証するインタフェースとなっている。また、反対にio.WriterToインタフェースもReaderが読み込んだデータの提供先を保証するインタフェースとなっている。なお、io.Copy関数では、srcがio.WriterToを実装していることを期待されている。

	func Copy(dst Writer, src Reader) (written int64, err error)

### 引数のWriter（出力先）とReader（入力元）の順序

上述のio.Copy関数のように、Go言語では、出力先と入力元の順序が他の言語と逆になっていることが多い。例えば、http.HandlerのServeHTTPメソッドもhttp.ResponseWriterが第1引数で、http.Requestが第2引数である。

### io.ReaderAtインタフェース

io.ReaderAtインタフェースは以下のように定義されている。

	type ReaderAt interface {
    		ReadAt(p []byte, off int64) (n int, err error)
	}

第2引数はoffsetであり、読み込み開始位置がoffset分ずれて読み込まれる。
一方、io.Seeker.Seekメソッドでは、次の読み込み／書き込み開始位置をoffset分ずらす役割があるが、第1戻り値で新しいoffsetを返すことができる。

また、ReadAtメソッドは、n < len(p)のとき、全データを読み込むかエラーが起きるまでブロックする。一方、io.Reader.Readメソッドはブロックしない。
つまり、ReadAtメソッドは、戻り値のnがn < len(p)のとき、エラーにてその理由を返す。
呼び出し側は、同じソースからReadAtメソッドを使用してパラレルに読み込んでよい。

### io.WriterAtインタフェース

io.WriterAtインタフェースは以下のように定義されている。

	type WriterAt interface {
    		WriteAt(p []byte, off int64) (n int, err error)
	}

オーバーラップしていなければ、io.WriterAt.WriteAtメソッドは、パラレルに書き込んで良い。

### io.ByteScanner.UnreadByteメソッド

io.ByteScannerインタフェースは以下のように定義されている。

	type ByteScanner interface {
		ByteReader
		UnreadByte() error
	}

UnreadByteメソッドは、読み込んだバイトを読み込まなかった事にし、次のReadByte()メソッドで前回読み込んだバイトと同じものを返す。
[net/textproto/reader.go](http://golang.org/src/pkg/net/textproto/reader.go:159)のskipSpace()で、空白を飛ばしていき、空白以外が来たら1バイト戻って、最後の空白の次から処理をするという使われ方をしている。

### stringerWriterインタフェースについて

io.stringerWriterインタフェースは、[io/io.go](http://golang.org/src/pkg/io/io.go)でWriteStringを実装しているかどうかをチェックする為だけに使われている。

### io.Copyの戻り値のint64について

io.Copyの戻り値のint64を超えるようなコピーが行なわれた場合、オーバーフローし、マイナスの値をとることもあるのではないか？
例えば、あるReaderから入ってきたデータをそのままあるWriterに受け流すような処理があり、その処理が常時行なわれるようなものだった場合、そのうちint64の範囲を超えるのではないか？
ちなみに、オーバーフローを起こしても、panicにはならない。

### io.LimitedReaderはデコレータパターン

io.LimitedReader型は以下のように定義されている。

	type LimitedReader struct {
		R Reader // underlying reader
		N int64  // max bytes remaining
	}

実際の読み込む処理については、LimitedReader.Rに[任せており](https://code.google.com/p/go/source/browse/src/pkg/io/io.go#387)、これはデザインパターンのデコレータパターンになっている。

### io.TeeReader関数
io.TeeReader関数は以下の様な関数である。

	func TeeReader(r Reader, w Writer) Reader

この関数は、Unix/Linuxコマンドのteeと同じような働きをする。
戻り値で返されたReaderのReadメソッドで読み込むと、その内容がwで指定したWriterに書き込まれる。
実際の読み込みには引数rで指定したReaderが使用される。

### bufioパッケージ
bufio は buffer を使ってシステムコールを減らす。
（詳細を忘れました）

### _test.goファイル

_test.goに書いたものはコンパイル対象から外れる。
src/pkg/runtime/以下にある_linux.goなどのファイルはコンパイル時に環境によってうまく使用するファイルを選ぶのか？
たとえば、linuxの場合は_linux.goをコンパイルに含めるが、windowsのものは含めないとか。

追記：
鵜飼さんから、Google+のコミュニティで_test.goや_linux.goが説明してあるページを教えてもらいました。
https://plus.google.com/117100596700604439455/posts/UDyCVWhAXKB


