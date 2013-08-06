# Heka

## Heka とは

Mozilla が作ってるログ収集システム. 入力・フィルター・出力がプラグインに
なっており、 Go あるいは Lua でプラグインを書くことが可能。

今人気の Fluentd の競合.
単にログを集約して保存するだけでなく、1レコードずつ何らかの処理を
行いたい場合は、メソッド呼び出し速度の差の影響で Fluentd よりも
Heka のほうが速そう.

## 参照

http://heka-docs.readthedocs.org/en/latest/

概要だけ書かれたドキュメント。

https://hekad.readthedocs.org/en/latest/

設定方法や Lua によるプラグインの書き方などが書かれたドキュメント。
必読だけど、これだけ読んでもやっぱりよくわからない.

https://github.com/mozilla-services/heka

Heka 本体のソースコード.

https://github.com/mozilla-services/heka-build

Heka のビルド環境を構築するためのコード。 Go 自体もこれで用意できる。
安定した Heka バイナリを作りたいなら heka-build を使ったほうがいいけど、
ハックするだけなら多分不要なので今回はパス


## 今回のターゲット

実践的な、サービスの裏側で使われるデーモン型のネットワークサーバープログラムの
書き方の調査.

Fluentd を使ってて Heka に乗り換えるつもりがない人は FluenGo を作るのに
必要なノウハウを吸収しよう。 Heka を使いたい人は Heka の内部に詳しくなって
より信頼して使えるようになろう.

具体的なポイント

* シグナルハンドラ
* リロード機構
* I/O のバッファリングと、遅いピア対策
* その他信頼性・安定性のための工夫点


# ビルド

基本的に hekad ドキュメント通りにやればOK. 以下抜粋.

## 準備

Prerequisites (all systems):

* CMake 2.8 or greater http://www.cmake.org/cmake/resources/software.html
* Git http://code.google.com/p/msysgit/downloads/list
* Go 1.1 or greater (1.1.1 recommended) http://code.google.com/p/go/downloads/list
* Mercurial http://mercurial.selenic.com/downloads/
* Protobuf 2.3 or greater http://code.google.com/p/protobuf/downloads/list
* Sphinx (optional - used to generate the documentation) http://sphinx-doc.org/

Prerequisites (Unix):

* make
* gcc
* patch
* dpkg (optional)
* rpmbuild (optional)
* packagemaker (optional)

## ビルド手順

```
$ git clone https://github.com/Mozilla-services/heka
$ cd heka
$ source build.sh
```

`build.sh` は実行するのではなくて source することに注意. (`. build.sh` でも可)
`build/heka/bin` 以下にバイナリが作られる。


# 動かしてみる

設定ファイルは TOML 形式で書く。
TCP で受け取って、すべてのメッセージを Go の log パッケージ経由で標準出力に出力する設定ファイル:

```
[TcpInput]
address = ":5566"

[LogOutput]
message_matcher = "TRUE"
```

heka-py から message を送信してみる.

結構フォーマットが面倒そう?
今回のターゲットからちょっとずれるので後回し.

# Code Reading

## cmd/hekad/main

`hekad` コマンドの main 関数がある. 抜粋:

```go
import (
	"github.com/mozilla-services/heka/pipeline"
)

	// Set up and load the pipeline configuration and start the daemon.
	globals := pipeline.DefaultGlobals()
	pipeconf := pipeline.NewPipelineConfig(globals)
	err = pipeconf.LoadFromConfigFile(*configFile)
	pipeline.Run(pipeconf)
```

どうやら pipeline が本体のようだ.

## pipeline

### Globals

`globals` が持ってるのはグローバル設定の情報だけ

```pipeline/pipeline_runner.go
// Struct for holding global pipeline config values.
type GlobalConfigStruct struct {
	PoolSize            int
	DecoderPoolSize     int
	PluginChanSize      int
	MaxMsgLoops         uint
	MaxMsgProcessInject uint
	MaxMsgTimerInject   uint
	Stopping            bool
	sigChan             chan os.Signal
}

// Creates a GlobalConfigStruct object populated w/ default values.
func DefaultGlobals() (globals *GlobalConfigStruct) {
	return &GlobalConfigStruct{
		PoolSize:            100,
		DecoderPoolSize:     2,
		PluginChanSize:      50,
		MaxMsgLoops:         4,
		MaxMsgProcessInject: 1,
		MaxMsgTimerInject:   10,
	}
}
```

### PipelineConfig

```pipeline/config.go
// Master config object encapsulating the entire heka/pipeline configuration.
type PipelineConfig struct {
	// All running InputRunners, by name.
	InputRunners map[string]InputRunner
	// PluginWrappers that can create Input plugin objects.
	inputWrappers map[string]*PluginWrapper
...
	// Name of host on which Heka is running.
	hostname string
	// Heka process id.
	pid int32
}
```

NewPipelineConfig() も、殆どからの PipelineConfig を作るだけ。

### PipelineConfig.LoadFromConfigFile

```pipeline/config.go
// LoadFromConfigFile loads a TOML configuration file and stores the
// result in the value pointed to by config. The maps in the config
// will be initialized as needed.
//
// The PipelineConfig should be already initialized before passed in via
// its Init function.
func (self *PipelineConfig) LoadFromConfigFile(filename string) (err error) {
```

各セクションに対して `self.loadSection` を呼んでいる.

```pipeline/config.go
// loadSection must be passed a plugin name and the config for that plugin. It
// will create a PluginWrapper (i.e. a factory). For decoders we store the
// PluginWrappers and create pools of DecoderRunners for each type, stored in
// our decoder channels. For the other plugin types, we create the plugin,
// configure it, then create the appropriate plugin runner.
func (self *PipelineConfig) loadSection(sectionName string,
	configSection toml.Primitive) (errcnt uint) {
```

時間があったら追いかける.

### pipeline.Run

先に PipelinePack を見ておく.
これは1メッセージを格納して pipeline に入るデータ構造.
2種類の pool で再利用される.

```pipeline/pipeline_runner.go
// Main Heka pipeline data structure containing raw message data, a Message
// object, and other Heka related message metadata.
type PipelinePack struct {
```

```pipeline/pipeline_runner.go
// Main function driving Heka execution. Loads config, initializes
// PipelinePack pools, and starts all the runners. Then it listens for signals
// and drives the shutdown process when that is triggered.
func Run(config *PipelineConfig) {
...
	for name, output := range config.OutputRunners {
		outputsWg.Add(1)
		if err = output.Start(config, &outputsWg); err != nil {
			log.Printf("Output '%s' failed to start: %s", name, err)
			outputsWg.Done()
			continue
		}
		log.Println("Output started: ", name)
	}
```

こんな感じで、 OutputRunner, FilterRunner を起動し、 pool を作り、
router, InputRunner を起動する.

メインの goroutine はシグナル待ちに入る.

```pipeline/pipeline_runner.go
	// wait for sigint
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGHUP, SIGUSR1)

	for !globals.Stopping {
		select {
		case sig := <-sigChan:
			switch sig {
			case syscall.SIGHUP:
				log.Println("Reload initiated.")
				if err := notify.Post(RELOAD, nil); err != nil {
					log.Println("Error sending reload event: ", err)
				}
			case syscall.SIGINT:
				log.Println("Shutdown initiated.")
				globals.Stopping = true
			case SIGUSR1:
				log.Println("Queue report initiated.")
				go config.allReportsStdout()
			}
		}
	}
```

globals.Stopping になったら、 Inut, decoder, filter, output の順にシャットダウンしていく.

### Input

メッセージを受け取って DecoderRunner に渡す


DecoderRunner は inChan から受け取った PipelinePack を decode し、 router に渡す.

```pipeline/decoders.go
		for pack = range dr.inChan {
			if err = dr.Decoder().Decode(pack); err != nil {
				dr.LogError(err)
				pack.Recycle()
				continue
			}
			pack.Decoded = true
			h.PipelineConfig().router.InChan() <- pack
		}
```

デコーダは、PipelinePack.MsgBytes をデコードして PipelinePack.Message に格納する.

```
func (self *JsonDecoder) Decode(pack *PipelinePack) error {
	return json.Unmarshal(pack.MsgBytes, pack.Message)
}
func (self *ProtobufDecoder) Decode(pack *PipelinePack) error {
	return proto.Unmarshal(pack.MsgBytes, pack.Message)
}
```

### router

参照カウントをインクリメントしながら matcher.inChan に投げまくり. あと放置.

```pipeline/router.go
			case pack, ok = <-self.inChan:
				if !ok {
					break
				}
				atomic.AddInt64(&self.processMessageCount, 1)
				for _, matcher = range self.fMatchers {
					if matcher != nil {
						atomic.AddInt32(&pack.RefCount, 1)
						matcher.inChan <- pack
					}
				}
				for _, matcher = range self.oMatchers {
					if matcher != nil {
						atomic.AddInt32(&pack.RefCount, 1)
						matcher.inChan <- pack
					}
				}
				pack.Recycle()
			}
```

### Matcher

yacc 使ってミニ言語作ってる. (pipeline/message)
MatchRunner は基本的に match したら matchChan に流すだけ.

ただし、100万+1000 でサンプリングしてカウントしてる.

### foRunner
FilterRunner と OutputRunner の両方の interface を実装したのが foRunner

matcher があれば起動し、 inChan で受け取る。
matcher の実態は MatchRunner で、 Pipeline.loadSection で作られてる.

```pipeline/config.go
	if pluginGlobals.Matcher != "" {
		if matcher, err = NewMatchRunner(pluginGlobals.Matcher,
			pluginGlobals.Signer); err != nil {
			self.log(fmt.Sprintf("Can't create message matcher for '%s': %s",
				wrapper.name, err))
			errcnt++
			return
		}
		runner.matcher = matcher
	}
```

filter.Run() か output.Run() を起動し、エラーだったらリトライし、リトライ回数制限超えたら終わる.

### Output

# 感想

全体的に、チャンネルに送信できなかった時のエラー処理がない。
１箇所のアウトプットが詰まった時に全体がつまらないようにする
仕組みがどこにあるのかわからなかった。

Channel を簡単にメッセージキューに使ってる。
ちょっと goroutine + channel 使い過ぎでは?
特に MatchRunner のところ, match したら関数呼び出すだけで良いんじゃないか。

# vim: syntax=markdown
