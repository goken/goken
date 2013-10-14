# Go研 Vol.8
- 発表者: [@oshothebig](http://twitter.com/oshothebig/)

## encoding/binaryの概要
- 複数の値とバイト列の相互の変換を行う
    - 複数の値は固定長の値として解釈される
    - 固定値の値として扱えるもの
        - 数値型: int8, uint8, int16, float32, complex64, …
        - 配列
        - 固定長の値のみから構成される構造体
    - エンディアンの指定が可能（ビッグエンディアン、リトルエンディアン）
- varintのエンコード/デコードする
    - varintとは、その値に合わせてバイト列の長さの変わる整数値のこと（[仕様](https://developers.google.com/protocol-buffers/docs/encoding?hl=ja&csw=1): Protocol Buffer由来）

## `Read`と`Write`
`binary`パッケージで主に使う関数は`Read`と`Write`である。ある値をbyteスライスに変換する場合は、`Read`を使い、byteスライスから値に戻したい場合は`Write`を使う。`order`を指定することで、エンディアンの指定が出来る。

```go
func Read(r io.Reader, order ByteOrder, data interface{}) error
func Write(w io.Writer, order ByteOrder, data interface{}) error
```

- 例
    - `Read`
        - ビッグエンディアン: <http://play.golang.org/p/ECSFzRrtx0>
        - リトルエンディアン: <http://play.golang.org/p/Utw-Q74GLU>
    - `Write`
        - ビッグエンディアン: <http://play.golang.org/p/RXVEwvrTcR>
        - リトルエンディアン: <http://play.golang.org/p/VkgfdNuLa9>

`data`に渡すことが出来る型は、`Read`の場合、固定長の値へのポインタもしくは固定長の値から構成されるスライスである。`Write`の場合、それらに加えて固定長の値そのものを渡すことが出来る。`data`には、構造体も渡せるが固定長の方ではなくてはいけない。そのため、単体のスライスは扱えるにも関わらず、スライスを含んだ構造体は`binary`パッケージでは扱えないことに注意する必要がある。

## ByteOrderインターフェイス
バイトオーダの変換を行う振る舞いが`ByteOrder`インターフェイスとして定義されている。符号無し整数をbyteスライスに変換するメソッド(`PutUintXX([]byte, uintXX)`)、byteスライスを符号無し整数に変換するメソッド(`UintXX([]byte) uintXX`)をもつ。

```go
type ByteOrder interface {
    Uint16([]byte) uint16
    Uint32([]byte) uint32
    Uint64([]byte) uint64
    PutUint16([]byte, uint16)
    PutUint32([]byte, uint32)
    PutUint64([]byte, uint64)
    String() string
}
```

リトルエンディアンとビッグエンディアンを表す`ByteOrder`インターフェイスの実装が用意されている。
<http://golang.org/pkg/encoding/binary/#pkg-variables>
```go
var BigEndian bigEndian
var LittleEndian littleEndian
```

`bigEndian`型と`littleEndian`型が何かを見てみると空の構造体として定義されている。
<http://golang.org/src/pkg/encoding/binary/binary.go>

```go
type bigEndian struct {}
type littleEndian struct{}
```

それらに対して`ByteOrder`インターフェイスのメソッドが定義されている。`Read()`および`Write()`に渡す`ByteOrder`は`BigEndian`もしくは`LittleEndian`を渡し、エンディアンの指定を行うのが普通である。変態的なエンディアン（ミドルエンディアンというがあるらしい: [Wikipedia](http://ja.wikipedia.org/wiki/%E3%82%A8%E3%83%B3%E3%83%87%E3%82%A3%E3%82%A2%E3%83%B3)）を使用する場面に遭遇しない限り、`BigEndian`もしくは`LittleEndian`で事足りる。

## コードを読む
エンコード、デコードともに[Go研 vol.8](https://github.com/goken/goken/blob/master/goken08-reflect.md)で勉強した`reflect`パッケージが多用されている。

### エンコードの仕組み
[encoding/binary/binary.go](http://golang.org/src/pkg/encoding/binary/binary.go?s=5194:5258#L179)

```
func Write(w io.Writer, order ByteOrder, data interface{}) error {
	// Fast path for basic types.
	var b [8]byte
	var bs []byte
	switch v := data.(type) {
	case *int8:
		bs = b[:1]
		b[0] = byte(*v)
	case int8:
		bs = b[:1]
		b[0] = byte(v)
	case *uint8:
		bs = b[:1]
		b[0] = *v
	case uint8:
		bs = b[:1]
		b[0] = byte(v)
	case *int16:
		bs = b[:2]
		order.PutUint16(bs, uint16(*v))
	case int16:
		bs = b[:2]
		order.PutUint16(bs, uint16(v))
	case *uint16:
		bs = b[:2]
		order.PutUint16(bs, *v)
	case uint16:
		bs = b[:2]
		order.PutUint16(bs, v)
	case *int32:
		bs = b[:4]
		order.PutUint32(bs, uint32(*v))
	case int32:
		bs = b[:4]
		order.PutUint32(bs, uint32(v))
	case *uint32:
		bs = b[:4]
		order.PutUint32(bs, *v)
	case uint32:
		bs = b[:4]
		order.PutUint32(bs, v)
	case *int64:
		bs = b[:8]
		order.PutUint64(bs, uint64(*v))
	case int64:
		bs = b[:8]
		order.PutUint64(bs, uint64(v))
	case *uint64:
		bs = b[:8]
		order.PutUint64(bs, *v)
	case uint64:
		bs = b[:8]
		order.PutUint64(bs, v)
	}
	if bs != nil {
		_, err := w.Write(bs)
		return err
	}

	// Fallback to reflect-based encoding.
	v := reflect.Indirect(reflect.ValueOf(data))
	size, err := dataSize(v)
	if err != nil {
		return errors.New("binary.Write: " + err.Error())
	}
	buf := make([]byte, size)
	e := &encoder{order: order, buf: buf}
	e.value(v)
	_, err = w.Write(buf)
	return err
}
```

整数型（符号付き、符号無しともに）および整数型のポインタは、型スイッチ文によって判断され`ByteOrder`インターフェイスを用いてbyteスライスに変換される。それ以外の型の場合は、リフレクションを用いて得られる型情報を元にbyteスライスに変換される。

`v := reflect.Indirect(reflect.ValueOf(data))`をすることで、値がポインタ型の場合はそのポインタが指し示す値の`reflect.Value`オブジェクトを得る。`dataSize(v)`で、値の長さを取得し、その長さ分のbyteスライスを生成し`encoder`で`reflect.Value`オブジェクトが示す値をbyteスライスに変換する。

```
type coder struct {
	order ByteOrder
	buf   []byte
}

type encoder coder
```

`encoder`は、`ByteOrder`型のオブジェクトとbyteスライスからなる構造体である。

```go
func (e *encoder) value(v reflect.Value) {
	switch v.Kind() {
	case reflect.Array:
		l := v.Len()
		for i := 0; i < l; i++ {
			e.value(v.Index(i))
		}

	case reflect.Struct:
		t := v.Type()
		l := v.NumField()
		for i := 0; i < l; i++ {
			// see comment for corresponding code in decoder.value()
			if v := v.Field(i); v.CanSet() || t.Field(i).Name != "_" {
				e.value(v)
			} else {
				e.skip(v)
			}
		}

	case reflect.Slice:
		l := v.Len()
		for i := 0; i < l; i++ {
			e.value(v.Index(i))
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v.Type().Kind() {
		case reflect.Int8:
			e.int8(int8(v.Int()))
		case reflect.Int16:
			e.int16(int16(v.Int()))
		case reflect.Int32:
			e.int32(int32(v.Int()))
		case reflect.Int64:
			e.int64(v.Int())
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		switch v.Type().Kind() {
		case reflect.Uint8:
			e.uint8(uint8(v.Uint()))
		case reflect.Uint16:
			e.uint16(uint16(v.Uint()))
		case reflect.Uint32:
			e.uint32(uint32(v.Uint()))
		case reflect.Uint64:
			e.uint64(v.Uint())
		}

	case reflect.Float32, reflect.Float64:
		switch v.Type().Kind() {
		case reflect.Float32:
			e.uint32(math.Float32bits(float32(v.Float())))
		case reflect.Float64:
			e.uint64(math.Float64bits(v.Float()))
		}

	case reflect.Complex64, reflect.Complex128:
		switch v.Type().Kind() {
		case reflect.Complex64:
			x := v.Complex()
			e.uint32(math.Float32bits(float32(real(x))))
			e.uint32(math.Float32bits(float32(imag(x))))
		case reflect.Complex128:
			x := v.Complex()
			e.uint64(math.Float64bits(real(x)))
			e.uint64(math.Float64bits(imag(x)))
		}
	}
}
```

数値型の場合は、`encoder`の対応するメソッド(`encoder.uint8()`など)を用いてbyteスライスに変換される。配列およびスライスの場合は、各要素について`encoder.value()`が再帰的に呼び出される。

構造体型の場合は、フィールドによって動作が変わる。

- 設定可能なフィールド、もしくは、ブランクフィールドではない場合: 再帰的に`encoder.value()`が呼び出される
- それ以外の場合（つまり、ブランクフィールドの場合）: 0でパディングされる

- 疑問
    - 数値型の場合、`v.Kind()`型スイッチした後、`v.Type().Kind()`しているが、得られる値が違うのか？

### デコードの仕組み
[encoding/binary/binary.go](http://golang.org/src/pkg/encoding/binary/binary.go?s=3732:3795#L121)

```go
func Read(r io.Reader, order ByteOrder, data interface{}) error {
	// Fast path for basic types.
	if n := intDestSize(data); n != 0 {
		var b [8]byte
		bs := b[:n]
		if _, err := io.ReadFull(r, bs); err != nil {
			return err
		}
		switch v := data.(type) {
		case *int8:
			*v = int8(b[0])
		case *uint8:
			*v = b[0]
		case *int16:
			*v = int16(order.Uint16(bs))
		case *uint16:
			*v = order.Uint16(bs)
		case *int32:
			*v = int32(order.Uint32(bs))
		case *uint32:
			*v = order.Uint32(bs)
		case *int64:
			*v = int64(order.Uint64(bs))
		case *uint64:
			*v = order.Uint64(bs)
		}
		return nil
	}

	// Fallback to reflect-based decoding.
	var v reflect.Value
	switch d := reflect.ValueOf(data); d.Kind() {
	case reflect.Ptr:
		v = d.Elem()
	case reflect.Slice:
		v = d
	default:
		return errors.New("binary.Read: invalid type " + d.Type().String())
	}
	size, err := dataSize(v)
	if err != nil {
		return errors.New("binary.Read: " + err.Error())
	}
	d := &decoder{order: order, buf: make([]byte, size)}
	if _, err := io.ReadFull(r, d.buf); err != nil {
		return err
	}
	d.value(v)
	return nil
}
```

整数型（符号あり、符号無し双方）のポインタは、型スイッチ文によって判断され`ByteOrder`インターフェイスを用いてbyteスライスから対応する整数型の値に変換される。それ以外の型の場合は、リフレクションを用いて得られる型情報をbyteスライスから与えられた型に変換される。このとき、ポインタ型かスライスではないとエラーが返る。`decoder.value()`を用いてデコードが行われる。

```go
type decoder coder
```

`decoder`は、`encoder`と同じく`coder`から定義される。

```go
func (d *decoder) value(v reflect.Value) {
	switch v.Kind() {
	case reflect.Array:
		l := v.Len()
		for i := 0; i < l; i++ {
			d.value(v.Index(i))
		}

	case reflect.Struct:
		t := v.Type()
		l := v.NumField()
		for i := 0; i < l; i++ {
			// Note: Calling v.CanSet() below is an optimization.
			// It would be sufficient to check the field name,
			// but creating the StructField info for each field is
			// costly (run "go test -bench=ReadStruct" and compare
			// results when making changes to this code).
			if v := v.Field(i); v.CanSet() || t.Field(i).Name != "_" {
				d.value(v)
			} else {
				d.skip(v)
			}
		}

	case reflect.Slice:
		l := v.Len()
		for i := 0; i < l; i++ {
			d.value(v.Index(i))
		}

	case reflect.Int8:
		v.SetInt(int64(d.int8()))
	case reflect.Int16:
		v.SetInt(int64(d.int16()))
	case reflect.Int32:
		v.SetInt(int64(d.int32()))
	case reflect.Int64:
		v.SetInt(d.int64())

	case reflect.Uint8:
		v.SetUint(uint64(d.uint8()))
	case reflect.Uint16:
		v.SetUint(uint64(d.uint16()))
	case reflect.Uint32:
		v.SetUint(uint64(d.uint32()))
	case reflect.Uint64:
		v.SetUint(d.uint64())

	case reflect.Float32:
		v.SetFloat(float64(math.Float32frombits(d.uint32())))
	case reflect.Float64:
		v.SetFloat(math.Float64frombits(d.uint64()))

	case reflect.Complex64:
		v.SetComplex(complex(
			float64(math.Float32frombits(d.uint32())),
			float64(math.Float32frombits(d.uint32())),
		))
	case reflect.Complex128:
		v.SetComplex(complex(
			math.Float64frombits(d.uint64()),
			math.Float64frombits(d.uint64()),
		))
	}
}
```

数値型の場合は、`decoder`の対応するメソッド(`decoder.uint8()`など)を用いて得られた値をリフレクションを用いて設定している。配列およびスライスの場合は、各要素について`decoder.value()`が再帰的に呼び出される。

構造体型の場合は、フィールドによって動作が変わる。

- 設定可能なフィールド、もしくは、ブランクフィールドではない場合: 再帰的に`decoder.value()`が呼び出される
	- 条件分岐は最適化のため、なるべくStructFieldを参照しないようにしている
- それ以外の場合（つまり、ブランクフィールドの場合）: 長さ分が読み飛ばされる
    - エクスポートされないフィールドがあるとpanicが発生: <http://play.golang.org/p/ccsVEYuKA5>

## 手をつけられなかったところ
Varint関連部分: [varint.go](http://golang.org/src/pkg/encoding/binary/varint.go)

## 参考情報
- ドキュメント: <http://golang.org/pkg/encoding/binary/>
- ソースコード
    - [binary.go](https://code.google.com/p/go/source/browse/src/pkg/encoding/binary/binary.go)
    - [varint.go](https://code.google.com/p/go/source/browse/src/pkg/encoding/binary/varint.go)

## 