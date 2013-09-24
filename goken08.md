#Go研 Vol.8

* 発表者：[@tenntenn](https://twitter.com/tenntenn)

## reflectパッケージと良く使うパターン

Go言語でリフレクションを行なうには、`reflect`パッケージを使用する。
`reflect`パッケージは標準パッケージでも多く使われていて、`encoding`パッケージがよい例である。

* `reflect`パッケージ: http://golang.org/pkg/reflect
* `encoding`パッケージ: http://golang.org/pkg/encoding

JSONをエンコード／デコードする`encoding/json`パッケージでは、

```
type Hoge struct {
    Field1 string `json:"field1"`
    Field2 int    `json:"field2"`
    Field3 string `json:"-"`
}
```

のような構造体を作っておき、リフレクションすることでエンコードしてJSONにしたり、JSONをデコードして構造体に落とし込んでいる。また、デコード結果を変数に格納する時も、汎用的にするために`interface{}`型で返すのではなく、ポインタを`interface{}`型で引数として受け取って値を設定している。

```
var hoge Hoge
// func Unmarshal([]byte, interface{}) error
err := json.Unmarshal(jsonStr, &hoge)
```

リフレクションのパターンには以下のようなものがあり、以降でそれぞれ説明する。

* 出力引数として`interface{}`を受け取るパターン
* `struct`のリフレクションパターン
* `channel`のリフレクションパターン
* `func`のリフレクションパターン

まず`reflect`パッケージの基礎となる`Value`型と`Type`型について説明し、その後、各パターンについて説明する。

## `Value`型と`Type`型

`reflect`パッケージを使いこなすには以下の2つの型を知っておく必要がある。

* `Value`型: http://golang.org/pkg/reflect/#Value
* `Type`型: http://golang.org/pkg/reflect/#Type

### `Value`型

`Value`型は以下のように定義(コメントは省略)されている。

https://code.google.com/p/go/source/browse/src/pkg/reflect/value.go#61

```src/pkg/reflect/value.go#61
type Value struct {
    typ *rtype
    val unsafe.Pointer
    flag
}
```

`rtype`型は後述する`Type`型の実体である。
つまり、`typ`は型情報である。
`val`はこの`Value`オブジェクトの表す値の実体へのポインタである。
`flag`は以下のようにメタ情報を保持している。

* 下位4ビット: フラグ用
    * flagRO: 読み込み専用か
    * flagIndir: ポインタかどうか
        * typ.size > ptrSizeならフラグが立つらしい
    * flagAddr: `v.CanAddr`が`true`かどうか
    * flagMethod: メソッドかどうか
* 次の5ビット: 値の種類
    * `Kind`メソッドで取得できる値
* 残りの23ビット: メソッドの為の領域?
    * この部分についてはよくわからない

なお、`flag`型は以下のように定義されている。
ビット操作の為のマスクやシフト幅などが`iota`を使ってうまく行なわれている([参考](http://qiita.com/tenntenn/items/0a3af58b225eeae29088))。

https://code.google.com/p/go/source/browse/src/pkg/reflect/value.go#96

```src/pkg/reflect/value.go#96
type flag uintptr
const (
        flagRO flag = 1 << iota 
        flagIndir              
        flagAddr              
        flagMethod
        // 以下はフラグ操作の為に使用
        flagKindShift        = iota // 4 (4ビット分シフトが必要)
        flagKindWidth        = 5 // there are 27 kinds
        flagKindMask    flag = 1<<flagKindWidth - 1
        flagMethodShift      = flagKindShift + flagKindWidth
)
```

`Value`型の値を取得するには`reflect.ValueOf()`を使う。

```
v := reflect.ValueOf(100)
```

`Value.Kind()`でその値の種類が取得できる。この値は`flag`の中に保持されている値である。

http://golang.org/pkg/reflect/#Kind

`Value`型の持つメソッドは実際の値によって呼べるものと呼べないものがある。
例えば、`int`型の値にスライス用のメソッドを呼び出すと`panic`を起こす。
それぞれの値の種類に対するメソッドの説明は後述する。

### `Type`型

`Type`型は、型を表すインタフェースで、以下のように定義されている。
メソッドやフィールドの情報等が取得できるのが分かる。

https://code.google.com/p/go/source/browse/src/pkg/reflect/type.go#31

```src/pkg/reflect/type.go#31
type Type interface {
        Align() int
        FieldAlign() int
        Method(int) Method
        MethodByName(string) (Method, bool)
        NumMethod() int
        Name() string
        PkgPath() string
        Size() uintptr
        String() string
        Kind() Kind
        Implements(u Type) bool
        AssignableTo(u Type) bool
        ConvertibleTo(u Type) bool
        Bits() int
        ChanDir() ChanDir
        IsVariadic() bool
        Elem() Type
        Field(i int) StructField
        FieldByIndex(index []int) StructField
        FieldByName(name string) (StructField, bool)
        FieldByNameFunc(match func(string) bool) (StructField, bool)
        In(i int) Type
        Key() Type
        Len() int
        NumField() int
        NumIn() int
        NumOut() int
        Out(i int) Type
        common() *rtype
        uncommon() *uncommonType
}
```

上記の中の非公開メソッドの`common()`は`*rtype`という型の値を返している。
この`rtype`という型が、`Type`インタフェースの実体であり、以下のように定義されている。
共通部分はこの型で定義しているようだ。

https://code.google.com/p/go/source/browse/src/pkg/reflect/type.go#243

```src/pkg/reflect/type.go#243
type rtype struct {
        size          uintptr        // size in bytes
        hash          uint32         // hash of type; avoids computation in hash tables
        _             uint8          // unused/padding
        align         uint8          // alignment of variable with this type
        fieldAlign    uint8          // alignment of struct field with this type
        kind          uint8          // enumeration for C
        alg           *uintptr       // algorithm table (../runtime/runtime.h:/Alg)
        gc            unsafe.Pointer // garbage collection data
        string        *string        // string form; unnecessary but undeniably useful
        *uncommonType                // (relatively) uncommon fields
        ptrToThis     *rtype         // type for pointer to this type, if used in binary or has methods
}
```

`uncommon()`は`*uncommonType`という型の値を返すが、`uncommonType`は以下のように定義されている。
名前やメソッドがある場合のみこの型の値が保持されるようだ。

https://code.google.com/p/go/source/browse/src/pkg/reflect/type.go#267

```src/pkg/reflect/type.go#267
type uncommonType struct {
        name    *string  // name of type
        pkgPath *string  // import path; nil for built-in types like int, string
        methods []method // methods associated with type
}
```

`Type`型は`Value.Type()`メソッドか`reflect.TypeOf`関数で取得できる。

```
v  := reflect.ValueOf(100)
t  := v.Type()
t2 := reflect.TypeOf(100)
```

それぞれの型のスライスやチャネルを作る関数がある。

* [`func ChanOf(dir ChanDir, t Type) Type`](http://golang.org/pkg/reflect/#ChanOf)
* [`func MapOf(key, elem Type) Type`](http://golang.org/pkg/reflect/#MapOf)
* [`func PtrTo(t Type) Type`](http://golang.org/pkg/reflect/#PtrTo)
* [`func SliceOf(t Type) Type`](http://golang.org/pkg/reflect/#SliceOf)

ちなみに、`Tyape`型から`New`を使ってオブジェクトが作れる。

http://play.golang.org/p/-kY7RBBLvr

```
package main

import (
    "fmt"
    "reflect"
)

type MyStruct struct {
    field1 string
}

func main() {
    var msp *MyStruct
    fmt.Println(msp)

    msv := reflect.New(reflect.TypeOf(msp).Elem())
    mspv := reflect.ValueOf(&msp)
    mspv.Elem().Set(msv)

    fmt.Println(msp)
}
```

## 出力引数として`interface{}`を受け取るパターン

### リフレクションを使って、変数に値を設定する

よく考えれば当たり前だが、一度ポインタにしないと変数の値を変えることができない。

http://play.golang.org/p/grX4uYh2VO

```
package main

import (
    "fmt"
    "reflect"
)

func main() {
    n := 100

    // ダメ
    nv := reflect.ValueOf(n)
    fmt.Println(nv.CanSet())

    // ポインタを使う
    npv := reflect.ValueOf(&n)
    fmt.Println(npv.Elem().CanSet())
    npv.Elem().SetInt(200)

    fmt.Println(n)
}
```

ポインタの`Value`型の値を取得後、`Elem`でポインタの指す、実体の`Value`型の値を取得している。
ちなみに、`reflect.Indirect`でもポインタの実体取得できる。
`Elem`メソッドの実装を見てみる。

https://code.google.com/p/go/source/browse/src/pkg/reflect/value.go#780

```
fl := v.flag&flagRO | flagIndir | flagAddr
```

まず、`flagRO`で論理積をとっているので、元の`v`のフラグのうち、`flagRO`の値だけ利用している。
つまり、元の`v`が読込み限定だった場合は、`v.Elem()`で取得したものも読込み限定となる。
次に、`flagIndir`と`flagAddr`が論理和で追加されている。
ここでこれらのフラグを付けている理由は、`CanSet`の実装を見るとわかる。

https://code.google.com/p/go/source/browse/src/pkg/reflect/value.go#330

```
func (v Value) CanSet() bool {
        return v.flag&(flagAddr|flagRO) == flagAddr
}
```

`CanSet`が`true`を返すためには、`flagRO`が立っておらず、`flagAddr`が立っている必要がある。
つまり、`Elem`で取得した値は`flagAddr`が立っているため、読込み限定でなければ、`Set`を使って値を設定できる。

一方で、`ValueOf`などで取得したものは、`ValueOf`の実装を見ると、`flagAddr`が立っていないことが分かる。

https://code.google.com/p/go/source/browse/src/pkg/reflect/value.go#2129

```
if typ.size > ptrSize {
    fl |= flagIndir
}
```

### ポインタを`interface{}`型で引数に受け取る

利用側がキャストをせずに、任意の型の値を生成したい場合がある。
Javaなどでは、ジェネリクスを使って、設定する値の型を指定することで、実現できる。
Go言語には、ジェネリクスがないため、出力引数として、`interface{}`型でポインタを受け取り、設定するパターンがある。
ここで気をつけたいのは、`interface{}`型で受け取ると、想定しているポインタ型ではないものが渡される可能性があるということだ。
`Kind`メソッドなどを使って、型のチェックを行なうと良いだろう。

http://play.golang.org/p/qsTwrL11mu

```reflect_pattern1.go
package main

import (
    "fmt"
    "reflect"
)

func set(p, v interface{}) error {
    pv := reflect.ValueOf(p)
    if pv.Kind() != reflect.Ptr {
        return fmt.Errorf("p must be pointer.")
    }

    vv := reflect.ValueOf(v)
    if pv.Elem().Kind() != vv.Kind() {
        return fmt.Errorf("p type and v type do not mutch")
    }

    pv.Elem().Set(vv)

    return nil
}

func main() {
    var hoge int
    fmt.Println(hoge)
    set(&hoge, 100)
    fmt.Println(hoge)
    fmt.Println(set(&hoge, 10.4))
}

```

## `struct`のリフレクションパターン
### フィールド情報を取得する
#### `Value`型から取得する

`Value`型からフィールドに関する情報を取得するメソッドは以下の通りである。
`Value`型からフィールド情報を取得すると、実際にその`struct`オブジェクトのフィールドに設定されている値が`Value`型で取得できる。

* [`func (v Value) Field(i int) Value`](http://golang.org/pkg/reflect#Field)
* [`func (v Value) FieldByIndex(index []int) Value`](http://golang.org/pkg/reflect#FieldByIndex)
* [`func (v Value) FieldByName(name string) Value`](http://golang.org/pkg/reflect#FieldByName)
* [`func (v Value) FieldByNameFunc(match func(string) bool) Value`](http://golang.org/pkg/reflect#FieldByNameFunc)
* [`func (v Value) NumField() int`](http://golang.org/pkg/reflect#NumField)

http://play.golang.org/p/VF1zJOITSr

```
package main

import (
    "fmt"
    "reflect"
)

type MyStruct struct {
    field1 string
    field2 MyStruct2
}

type MyStruct2 struct {
    field int
}

func main() {
    ms := MyStruct{"str", MyStruct2{100}}
    v := reflect.ValueOf(ms)

    // ms.field1
    fmt.Println(v.Field(0))

    // ms.field2.field
    fmt.Println(v.FieldByIndex([]int{1, 0}))

    // ms.field1
    fmt.Println(v.FieldByName("field1"))

    fmt.Println(v.FieldByNameFunc(func(name string) bool {
        return name == "field1"
    }))

    fmt.Println(v.NumField())
}
```

値を設定するには、通常の変数と同様にフィールドを保持している構造体のポインタを使用する必要がある。
また、フィールドが公開されていない場合(厳密には`PkgPath`に値が設定されていないフィールド)は、`flagRO`フラグがたつため、`CanSet`が`false`となり`Set`できない。

http://play.golang.org/p/SSW28W5bXn

```
package main

import (
    "fmt"
    "reflect"
)

type Hoge struct {
    N int
}

func main() {
    h := Hoge{10}
    hpv := reflect.ValueOf(&h)
    hpv.Elem().FieldByName("N").SetInt(200)

    fmt.Println(h)
}
```

#### `Type`型から取得する


一方、`Type`型からフィールドの情報を取得するメソッドは以下の通りである。
メソッド名や引数は`Value`型の場合と同じであるが、戻り値が`StructField`型になっている事に注意したい。
`Type`型は型情報なので、当然ながら実際の値の情報ではなく型に関する情報を保持している。
そのため、フィールドに関する型情報は`StructField`として定義している。

* `Field(i int) StructField`
* `FieldByIndex(index []int) StructField`
* `FieldByName(name string) (StructField, bool)`
* `FieldByNameFunc(match func(string) bool) (StructField, bool)`
* `NumField() int`

http://play.golang.org/p/EWOO-7Psza

```
package main

import (
    "fmt"
    "reflect"
)

type MyStruct struct {
    field1 string
    field2 MyStruct2
}

type MyStruct2 struct {
    field int
}

func main() {
    ms := MyStruct{"str", MyStruct2{100}}
    t := reflect.TypeOf(ms)

    // ms.field1
    fmt.Println(t.Field(0))

    // ms.field2.field
    fmt.Println(t.FieldByIndex([]int{1, 0}))

    // ms.field1
    fmt.Println(t.FieldByName("field1"))

    fmt.Println(t.FieldByNameFunc(func(name string) bool {
        return name == "field1"
    }))

    fmt.Println(t.NumField())
}
```

`StructField`型からフィールドに設定された、タグ情報を取得することができる。
`encoding/json`パッケージなどでは、このようにタグを使用している。

http://play.golang.org/p/xyCMeD5yuv

```
package main

import (
    "fmt"
    "reflect"
)

type Hoge struct {
    N int `json:"n"`
}

func main() {
    h := Hoge{10}
    t := reflect.TypeOf(h)
    n, _ := t.FieldByName("N")
    fmt.Println(n.Tag.Get("json"))
}
```

### メソッド情報を取得する
#### `Value`型から取得する
`Value`型からメソッドに関する情報を取得するメソッドは以下の通りである。
`Value`型から取得した取得したメソッドは、`Value.Call`メソッドで呼び出すことができる。

* [`func (v Value) Method(i int) Value`](http://golang.org/pkg/reflect/#Value.Method)
* [`func (v Value) MethodByName(name string) Value`](http://golang.org/pkg/reflect/#Value.MethodByName)
* [`func (v Value) NumMethod() int`](http://golang.org/pkg/reflect/#Value.NumMethod)

http://play.golang.org/p/9vp1cvCSMW

```
package main

import (
    "fmt"
    "reflect"
)

type MyStruct struct {
    field1 string
}

func (ms *MyStruct) method1() {
}

func (ms *MyStruct) Method2() {
    fmt.Println("method2!")
}

func main() {
    ms := &MyStruct{"str"}
    v := reflect.ValueOf(ms)

    // ms.method1
    fmt.Println(v.Method(0))

    // ms.Method2
    fmt.Println(v.MethodByName("Method2"))

    fmt.Println(v.NumMethod())

    v.MethodByName("Method2").Call([]reflect.Value{})
}
```

#### `Type`型から取得する
`Type`型からメソッドに関する情報を取得するメソッドは以下の通りである。

* `Method(int) Method`
* `MethodByName(string) (Method, bool)`
* `NumMethod() int`

http://play.golang.org/p/7mwSlX7YS_

```
package main

import (
    "fmt"
    "reflect"
)

type MyStruct struct {
    field1 string
}

func (ms *MyStruct) method1() {
}

func (ms *MyStruct) Method2() {
}

func main() {
    ms := &MyStruct{"str"}
    t := reflect.TypeOf(ms)

    // ms.method1
    fmt.Println(t.Method(0))

    // ms.Method2
    fmt.Println(t.MethodByName("Method2"))

    fmt.Println(t.NumMethod())
}
```

## `channel`のリフレクションパターン

### `channel`を作る
`reflect.MakeChan`関数を使うと、任意の型のchannelを作ることができる。
`Value`型にはchannel用のメソッドがありそれぞれの役割は以下のとおりである。

* `Value.Send`: チャネルにデータを送る`ch<-100`
* `Value.TrySend`: ブロックなしで、チャネルにデータを送る。送れなかったら戻り値が`false`
* `Value.Recv`: チャネルからデータを受け取る`<-ch`
* `Value.TryRecv`: ブロックなしで、チャネルからデータを受け取る。受け取れなかったら第2戻り値が`false`
* `Value.Len`: バッファの中にある容量
* `Value.Cap`: channelの容量

http://play.golang.org/p/xWHNPSJTUH

```
package main

import (
    "fmt"
    "reflect"
)

func main() {
    var ch chan int
    c := reflect.MakeChan(reflect.TypeOf(ch), 0)
    reflect.ValueOf(&ch).Elem().Set(c)

    go func() {
        ch <- 100
    }()
    fmt.Println(<-ch)
}
```

また、任意の型をやりとりするチャネルの型を取得したい場合は、`reflect.ChanOf`が使える。

http://play.golang.org/p/iqCeByVCoD

```
package main

import (
    "fmt"
    "reflect"
)

func main() {
    ct := reflect.ChanOf(reflect.SendDir, reflect.TypeOf(1))
    fmt.Println(ct)
}
```

第1引数の`ChanDir`は方向を表すフラグで以下のように定義されている。

```
type ChanDir int
const (
        RecvDir ChanDir             = 1 << iota // <-chan
        SendDir                                 // chan<-
        BothDir = RecvDir | SendDir             // chan
)
```

### `reflect.Select`を使う

`reflect.Select`を使用すると、任意の個数のcaseからなるselect文を実行できる。
`reflect.Select`関数は以下のように定義されて、引数に`Select`型のスライスを渡す。
戻り値は、

* `chosen`: 選んだcase
* `recv`: channelから受けっとた値
* `recvOK`: channelから受け取れたかどうか?（閉じてないか） 

である。

```
func Select(cases []SelectCase) (chosen int, recv Value, recvOK bool)
```

`Select`型は、以下のように定義されている。

```
type SelectCase struct {
        Dir  SelectDir // direction of case
        Chan Value     // channel to use (for send or receive)
        Send Value     // value to send (for send)
}
```

`SelectDir`は以下のように定義されており、caseで使用するchannelの方向または`default`caseかどうかを表す。

```
type SelectDir int
const (
        SelectSend    // case Chan <- Send
        SelectRecv    // case <-Chan:
        SelectDefault // default
)
```

チャネルの使い方をまとめたサンプル

https://github.com/golang-samples/reflect/blob/master/chan/main.go

## `func`のリフレクションパターン

### `func`をリフレクションする

メソッドと同様`Value.Call`メソッドで呼び出すことができる。

http://play.golang.org/p/0vsMpUNvOv

```
package main

import (
    "fmt"
    "reflect"
)

func main() {
    f := func(n int) {
        fmt.Println(n)
    }
    fv := reflect.ValueOf(f)
    fmt.Println(fv)

    fv.Call([]reflect.Value{reflect.ValueOf(100)})
}
```

型情報からは、引数や戻り値について取得できる。

* `Type.NumIn() int`: 引数の数
* `Type.In(i int) Type`:  `i`番目の引数の型を取得する
* `Type.NumOut() int`: 戻り値の数
* `Type.Out(i int) Type`: `i`番目の戻り値の型を取得する

http://play.golang.org/p/L290-TbjEu

```
package main

import (
    "fmt"
    "reflect"
)

func main() {
    f := func(n int) {
        fmt.Println(n)
    }
    ft := reflect.TypeOf(f)
    fmt.Println(ft)

    fmt.Println(ft.NumIn())
    for i := 0; i < ft.NumIn(); i++ {
        fmt.Println(ft.In(i))
    }

    fmt.Println(ft.NumOut())
    for i := 0; i < ft.NumOut(); i++ {
        fmt.Println(ft.Out(i))
    }
}
```

### 関数を作る

`reflect.MakeFunc`関数を使えば、動的に関数を作ることができる。
残念ながら、メソッドは動的には作れない。
関数を作る手順は以下のとおりである。

* `func(in []Value) []Value`型で関数を作る
* `MakeFunc`を使って関数を型情報を渡し、`Value`型に変換する
* 上記の変換の際に使用した型の変数に、`Value.Set`を使って設定する

合成関数を作る例を以下に示す。

http://play.golang.org/p/o5Vw9unsFL

```
package main

import (
    "fmt"
    "reflect"
)

func Compose(f, g, fptr interface{}) {
    fv := reflect.ValueOf(f)
    gv := reflect.ValueOf(g)

    fgv := func(in []reflect.Value) []reflect.Value {
        return gv.Call(fv.Call(in))
    }

    fn := reflect.ValueOf(fptr).Elem()

    v := reflect.MakeFunc(fn.Type(), fgv)

    fn.Set(v)
}

func main() {
    square := func(x int) int {
        return x * x
    }

    var fourthPow func(x int) int

    Compose(square, square, &fourthPow)
    fmt.Println(fourthPow(2))
}
```
