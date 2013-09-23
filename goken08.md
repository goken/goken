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
ビット操作の為のマスクやシフト幅などが`iota`を使ってうまく行なわれている。

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
