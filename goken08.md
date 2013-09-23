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

        // Methods applicable only to some types, depending on Kind.
        // The methods allowed for each kind are:
        //
        //      Int*, Uint*, Float*, Complex*: Bits
        //      Array: Elem, Len
        //      Chan: ChanDir, Elem
        //      Func: In, NumIn, Out, NumOut, IsVariadic.
        //      Map: Key, Elem
        //      Ptr: Elem
        //      Slice: Elem
        //      Struct: Field, FieldByIndex, FieldByName, FieldByNameFunc, NumField

        // Bits returns the size of the type in bits.
        // It panics if the type's Kind is not one of the
        // sized or unsized Int, Uint, Float, or Complex kinds.
        Bits() int

        // ChanDir returns a channel type's direction.
        // It panics if the type's Kind is not Chan.
        ChanDir() ChanDir

        // IsVariadic returns true if a function type's final input parameter
        // is a "..." parameter.  If so, t.In(t.NumIn() - 1) returns the parameter's
        // implicit actual type []T.
        //
        // For concreteness, if t represents func(x int, y ... float64), then
        //
        //      t.NumIn() == 2
        //      t.In(0) is the reflect.Type for "int"
        //      t.In(1) is the reflect.Type for "[]float64"
        //      t.IsVariadic() == true
        //
        // IsVariadic panics if the type's Kind is not Func.
        IsVariadic() bool

        // Elem returns a type's element type.
        // It panics if the type's Kind is not Array, Chan, Map, Ptr, or Slice.
        Elem() Type

        // Field returns a struct type's i'th field.
        // It panics if the type's Kind is not Struct.
        // It panics if i is not in the range [0, NumField()).
        Field(i int) StructField

        // FieldByIndex returns the nested field corresponding
        // to the index sequence.  It is equivalent to calling Field
        // successively for each index i.
        // It panics if the type's Kind is not Struct.
        FieldByIndex(index []int) StructField

        // FieldByName returns the struct field with the given name
        // and a boolean indicating if the field was found.
        FieldByName(name string) (StructField, bool)

        // FieldByNameFunc returns the first struct field with a name
        // that satisfies the match function and a boolean indicating if
        // the field was found.
        FieldByNameFunc(match func(string) bool) (StructField, bool)

        // In returns the type of a function type's i'th input parameter.
        // It panics if the type's Kind is not Func.
        // It panics if i is not in the range [0, NumIn()).
        In(i int) Type

        // Key returns a map type's key type.
        // It panics if the type's Kind is not Map.
        Key() Type

        // Len returns an array type's length.
        // It panics if the type's Kind is not Array.
        Len() int

        // NumField returns a struct type's field count.
        // It panics if the type's Kind is not Struct.
        NumField() int

        // NumIn returns a function type's input parameter count.
        // It panics if the type's Kind is not Func.
        NumIn() int

        // NumOut returns a function type's output parameter count.
        // It panics if the type's Kind is not Func.
        NumOut() int

        // Out returns the type of a function type's i'th output parameter.
        // It panics if the type's Kind is not Func.
        // It panics if i is not in the range [0, NumOut()).
        Out(i int) Type

        common() *rtype
        uncommon() *uncommonType
}
```
