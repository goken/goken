# goken07

## GO'CIRCUIT

- Goで分散処理を書くためのフレームワーク
- 任意のホストで処理を実行できる

例: "host25.datacenter.net"でfunc()を実行

````
feedback := make(chan int)
circuit.Spawn("host25.datacenter.net", func() {
	feedback <- 1
})
<-feedback
println("Roundtrip complete.")
````

- MapReduce処理を書いたりとか...
http://www.gocircuit.org/trend.html

## セットアップ
http://www.gocircuit.org/build.html

### 事前準備

Go Compilerをビルドする.

````
$ hg clone -u go1.1.1 https://code.google.com/p/go
$ export GOROOT=$HOME/go
$ cd $GOROOT/src
$ ./all.bash
````

GO'CIRCUITのソースコードを取得する

````
$ hg clone -u release https://code.google.com/p/gocircuit $GOCIRCUIT
$ export GOPATH=$GOPATH:$GOCIRCUIT
````

以下の環境変数を設定する
note: GO'CIRCUITのビルドには、ZookeeperのCライブラリが必要(GO'CIRCUITに同梱)

````
$ export ZKINCLUDE=$GOCIRCUIT/misc/starter-linux-osx/zookeeper/include
$ export ZKLIB=$GOCIRCUIT/misc/starter-kit-linux/zookeeper/lib
````

### GO'CIRCUITのビルド

cgo用の環境変数を設定する.
GoからCのコードを呼ぶために必要.(ZookeeperのC Libraryを使ってるから)
cgo -> http://golang.org/cmd/cgo/

note: 環境によっては、"-lm"オプションをつける必要あり.

````
$ export CGO_CFLAGS="-I$ZKINCLUDE"
$ export CGO_LDFLAGS="$ZKLIB/libzookeeper_mt.a -lm"
````

GO'CIRCUITをビルドする.

````
$ cd $GOCIRCUIT/src/circuit/cmd
$ go install ./...

$ export PATH=$PATH:$GOCIRCUIT/bin
````

circuit applicationを動かすために、Zookeeperが必要なので用意する.

````
$ curl -O http://ftp.jaist.ac.jp/pub/apache/zookeeper/zookeeper-3.4.5/zookeeper-3.4.5.tar.gz
$ tar xvf zookeeper-3.4.5.tar.gz
$ export PATH=$PATH:$HOME/zookeeper-3.4.5/bin
$ cp $HOME/zookeeper-3.4.5/conf/zoo_sample.cfg $HOME/zookeeper-3.4.5/conf/zoo.cfg
$ mkdir -p /tmp/zookeeper
$ zkServer.sh start
````

## 動かしてみる

note: 内部でsshでログインするので、パスワードなしでログインできるようにしておく.

### "Hello, world!"

http://www.gocircuit.org/hello.html

Tutorialの"Hello, world!"のビルド

````
$ export GOPATH=$GOPATH:$GOCIRCUIT/tutorials/hello
$ cd $GOCIRCUIT/tutorials/hello
$ CIR=app.config 4crossbuild
````

note: buildに失敗する場合は、"Build"に以下のプロパティを追加する.

````
"CGO_LDFLAGS": "{{ repo }}/misc/starter-kit-{{ os }}/zookeeper/lib/libzookeeper_mt.a -lm"
````

### "Hello, world!"の解説

src/hello/cmd/hello-spawn/main.go
- main packageでimportされた"circuit/load/cmd"でGO'CIRCUITの初期化処理を行う.
- main関数からSpawnを呼び出し、localhostでx.App{}に定義された関数を引数"world!"で実行する.

src/hello/worker/worker.go
- main packageでimportされた"circuit/load/worker"でGO'CIRCUITの初期化処理を行う.
- "circuit/load/worker"の初期化処理でBlockするので、worker.goのmain関数は実行されない

src/hello/x/hello.go
- init関数でcircuit.RegisterFunc関数を呼び出す


ビルドについて

GO'CIRCUITには専用のbuild, deploy toolが付属するので、それをを使用する.
note: ZookeeperのCライブラリに依存しているため、cross-compileできない。
各コマンドはapp.configに従って動作する.

- "4corssbuild"は"app.config"の"Host"に記載されたHostにsshでログインして、"4build"を実行する.
- "4build"は"app.config"の"Build"に従った設定で、go buildを実行する.
- "4crossbuild"はビルド後の実行ファイルをrsyncで取得する.("ShipDir"で指定したディレクトリにコピー)

"Hello, world!"のTutorialは上記手順のみでおｋ.(localhost内で別プロセスを起動するので)
リモートホストで実行する場合は、以下手順も必要.

- "4deploy"コマンドで実行ファイルをリモートホストにコピーする.

````
CIR=app.config 4deploy < host_list
````

- ローカルで"hello-spawn"をbuildする

````
$ cd $GOCIRCUIT/tutorials/hello/src/cmd/hello-spawn
$ go build
````

4corssbuild, 4build, 4deploy以外にもモニタリング用のコマンドなどがある。


## Codeを読む

Point
- 主にSpawnしてリモートプロセスを起動する部分について確認しました

http://www.gocircuit.org/spawn.html

### circuit/use/circuit/bind.go

Spawn呼び出しは、linkに委譲しているだけ.
linkは任意のオブジェクトを格納するキーペア.

````
func Bind(v interface{}) {
	link.Set(v)
}

func get() runtime {
	return link.Get().(runtime)
}

func Spawn(host string, anchor []string, fn Func, in ...interface{}) (retrn []interface{}, addr Addr, err error) {
	return get().Spawn(host, anchor, fn, in...)
}

func RunInBack(fn func()) {
	get().RunInBack(fn)
}
````

### circuit/load/cmd/init.go

GO'CIRCUITの初期化処理を行う.

役割によって、3パターンの初期化が定義されている.

config.Main -> Spawnを呼び出す親プロセス
config.Daemonizer -> workerプロセスを起動するプロセス
config.Worker -> workerプロセス

````
	switch config.Role {
	case config.Main:
		start(false, config.Config.Zookeeper, config.Config.Deploy, config.Config.Spark)
	case config.Worker:
		start(true, config.Config.Zookeeper, config.Config.Deploy, config.Config.Spark)
	case config.Daemonizer:
		workerBackend.Daemonize(config.Config)
	default:
		log.Println("Circuit role unrecognized:", config.Role)
		os.Exit(1)
	}
````

### circuit/sys/transport/transport.go

親プロセスとworkerプロセス間の通信処理は、TCP上をencoding/gob packageで
Goのオブジェクトを送受信している.

````
func (t *Transport) loop() {
	for {
		c, err := t.listener.AcceptTCP()
		if err != nil {
			panic(err) // Best not to be quiet about it
		}
		t.link(c, nil)
	}
}
````

### circuit/sys/transport/msg.go

プロセス間通信のメッセージ定義.

````
func init() {
	gob.Register(&welcomeMsg{})
	gob.Register(&openMsg{})
	gob.Register(&connMsg{})
	gob.Register(&linkMsg{})
}
````

### circuit/sys/runtime/runtime.go

GO'CIRCUITの状態を管理している


````
func New(t circuit.Transport) *Runtime {
	r := &Runtime{
		dialer: t,
		exp:    makeExpTabl(types.ValueTabl),
		imp:    makeImpTabl(types.ValueTabl),
		live:   make(map[circuit.Addr]struct{}),
		prof:   prof.New(),
	}
	r.srv.Init()
	go func() {
		for {
			r.accept(t)
		}
	}()
	r.Listen("acid", acid.New())
	return r
}
````

### circuit/sys/lang/types/register.go

workerプロセスで実行する関数を登録する.

````
func RegisterFunc(fn interface{}) {
	t := makeType(fn)
	if len(t.Func) != 1 {
		panic("fn type must have exactly one method: " + strconv.Itoa(len(t.Func)))
	}
	FuncTabl.Add(t)
}
````

### Spawn呼び出し

Spawn(circuit/use/circuit/bind.go)
->
Spawn(circuit/sys/lang/func.go)
->
Spawn(circuit/use/worker/worker.go)
->
Spawn(circuit/sys/worker/spawn.go)

1. sshでリモートホストにログインして、workerプロセスを起動する
2. workerプロセスがworkerID, PID, Addressを返す
3. 親プロセスがworkerプロセスにopenMsgを送信
4. 親プロセスがworkerプロセスにgoMsgを送信
5. workerプロセスが関数を実行する

circuit/sys/lang/runtime.go(77-105L)
````
go func() {
		req, err := conn.Read()
		if err != nil {
			log.Println("unexpected eof conn", err.Error())
			return
		}

		switch q := req.(type) {
		case *goMsg:
			r.serveGo(q, conn)
		case *dialMsg:
			r.serveDial(q, conn)
		case *callMsg:
			r.serveCall(q, conn)
		case *dropPtrMsg:
			r.serveDropPtr(q, conn)
		case *getPtrMsg:
			r.serveGetPtr(q, conn)
		case *dontReplyMsg:
			// Don't reply. Intentionally don't close the conn.
			// It will close when the process dies.
		default:
			log.Printf("unknown request %v", req)
		}
	}()
````

RunInBack内で関数をコールしている場合は、関数終了までblockする.

circuit/sys/lang/func.go(34-40L)
````
func (r *Runtime) RunInBack(fn func()) {
	r.dwg.Add(1)
	go func() {
		defer r.dwg.Done()
		fn()
	}()
}
````

最後に、workerプロセスから親プロセスにreturnMsgを送信
circuit/sys/lang/func.go(61L)


## 読んでないところ

- File Systemとか
Zookeeper上にprocfsのようなものを構築し、プロセス情報とか確認できるようにしているらしい.
http://www.gocircuit.org/anchor.html

