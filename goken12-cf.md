# Go CLI って

CloudFoundry において、アプリケーションのデプロイや起動停止などを行うためのAPIにアクセスするコマンドラインツール。

歴史的経緯:

````
CF v1: Ruby
CF v2: Ruby + xxxx
CF v2: Go
```

なんでGoにしたのさ?

- http://blog.cloudfoundry.com/2013/11/09/announcing-cloud-foundry-cf-v6/
- https://groups.google.com/a/cloudfoundry.org/forum/#!msg/vcap-dev/Xa56uBdjD1U/iAHPSscfMDAJ

*We've encountered quite a bit of pain from users installing a ruby interpreter just to use the cf gem*

- Windows とか Windows とか Windows だと、大抵Gemのインストールでこける。
- OS X でも make を必要とする拡張ライブラリ系でこける。
- Rubyのバージョンあげただけでこける(1.9.2 -> 1.9.3 でこけた)。
- インストールされたように見えて動かないとかもよくある(Windows上のtar.gzとか)

同じような話: http://wazanova.jp/post/67207541956/go

# ソースコードリーティング

https://github.com/cloudfoundry/cli

## 構成

````
cli + bin : ビルド用のユーティリティスクリプトがいくつか。
    |
    + installer
    |
    + src + cf : 後で読む
          |
          + code.github.com : [*1]
          |
          + fileutils: ヘルパ関数 [*2]
          |
          + fixtures: テストデータ
          |
          + generic: ヘルパ関数 [*2]
          |
          + githubc.com: [*1]
          |
          + glob: ヘルパ関数 [*2]
          |
          + main: メイン,後で読む
          |
          + testhelpers: また assert 再発明!
          |
          + terminal: ヘルパ関数 [*2]

````

- [*1] ライブラリの依存で依存先の互換性が壊れることを考慮して、code.google.com とか
- [*2] ユーティリティ/ヘルパ関数のパッケージングはどうするのが流儀?


## main/cf.go

https://github.com/codegangsta/cli を使ってコマンドライン引数とかの解析。このライブラリは便利!

実行中に panic が起こったら

## cf/commands/runner.go

cli ライブラリを使ってコマンドを実行する部分。

````
type Command interface {
        GetRequirements(reqFactory requirements.Factory, c *cli.Context) (reqs []requirements.Requirement, err error)
        Run(c *cli.Context)
}
type Runner interface {
        RunCmdByName(cmdName string, c *cli.Context) (err error)
}

type ConcreteRunner struct {
        cmdFactory Factory
        reqFactory requirements.Factory
}

func NewRunner(cmdFactory Factory, reqFactory requirements.Factory) (runner ConcreteRunner) {
        runner.cmdFactory = cmdFactory
        runner.reqFactory = reqFactory
        return
}

func (runner ConcreteRunner) RunCmdByName(cmdName string, c *cli.Context) (err error) {
        cmd, err := runner.cmdFactory.GetByCmdName(cmdName)
        if err != nil {
                fmt.Printf("Error finding command %s\n", cmdName)
                os.Exit(1)
                return
        }

        requirements, err := cmd.GetRequirements(runner.reqFactory, c)
        if err != nil {
                return
        }

        for _, requirement := range requirements {
                success := requirement.Execute()
                if !success {
                        err = errors.New("Error in requirement")
                        return
                }
        }

        cmd.Run(c)
        return
}
````

## cf/commands/login.go

login コマンドの実装


````
type Login struct {
        ui            terminal.UI
        config        *configuration.Configuration
        configRepo    configuration.ConfigurationRepository
        authenticator api.AuthenticationRepository
        endpointRepo  api.EndpointRepository
        orgRepo       api.OrganizationRepository
        spaceRepo     api.SpaceRepository
}

func (cmd Login) Run(c *cli.Context) {
        oldUserName := cmd.config.Username()

        apiResponse := cmd.setApi(c)
        if apiResponse.IsNotSuccessful() {
                cmd.ui.Failed("Invalid API endpoint.\n%s", apiResponse.Message)
                return
        }

        apiResponse = cmd.authenticate(c)
        if apiResponse.IsNotSuccessful() {
                cmd.ui.Failed("Unable to authenticate.")
                return
        }

        userChanged := (cmd.config.Username() != oldUserName && oldUserName != "")

        apiResponse = cmd.setOrganization(c, userChanged)
        if apiResponse.IsNotSuccessful() {
                cmd.ui.Failed(apiResponse.Message)
                return
        }

        apiResponse = cmd.setSpace(c, userChanged)
        if apiResponse.IsNotSuccessful() {
                cmd.ui.Failed(apiResponse.Message)
                return
        }

        cmd.ui.ShowConfiguration(cmd.config)
        return
}

````


# Gorouter って?

CloudFoundry において、 HTTPリクエストの(Hostヘッダベースの)ルーティングを行うエンジン。

歴史的経緯:

````
CF v1: Nginx => Ruby => Application Server
CF v1: Nginx(+ Lua + Ruby) => Application Server
CF v2: Go => Applicaiton Server
````

## (参考)

CloudFoundry は元々Rubyで作られていたが、Go に置き換えることを表明している。以下、Go で書かれている https://github.com/cloudfoundry 配下のリポジトリ

- cli: CLI of CloudFoundery - Ruby -> Go デプロイが楽!
- gonit: monit in golang - パフォーマンス問題を解決(したい)
- gosteno: logging library - ロガー
- gorouter: Routing of CloudFoundry - HTTP以外のプロトコルも(主にWebSocket上で)サポートしたいのでNginxなくしたい(え!?)
- gibson: Client Library fo gorouter - gorouter いじるツール
- loggregator: logging aggregator in golang - 機能としてなかったので作った系(fluentd 提案したのにorz)
- hm9000: Health Manager of CloudFoundry - 元々全面書き直しが必要だった?
- gosigar: Hyperic API client in golang.
- gordon: ...
-

車輪の再発明的なのも多いですが(笑)。

# ルーティングの仕組みの概要

## ルーティングテーブルのアップデート (追加)

1. アプリケーションサーバー上のAgentが、メッセージバスに `router.register` トピックにメッセージを飛ばす。メッセージに、ドメイン名、IP、そしてPortを含む。
2. Router が `router.register` トピックのメッセージを受け取り、{ドメイン名 => [(IP, Port)]} というマッピングを作る

なお、Router が `router.start` トピックのメッセージを送った場合にも、Agent がこのメッセージに対して、 `router.register` を応答するように設計されている。このメッセージのやりとりはメッセージバスの接続直後に行われる。

## ルーティングテーブルのアップデート (削除)

1. アプリケーションサーバー上のAgentが、メッセージバスに `router.unregister` トピックにメッセージを飛ばす。メッセージに、ドメイン名、IP、そしてPortを含む。
2. Router が `router.unregister` トピックのメッセージを受け取り、{ドメイン名 => [(IP, Port)]} というマッピングを削除する

## HTTPリクエストの処理

1. HTTPリクエストを受け取り、Host ヘッダを参照する
2. ルーティングテーブルを参照し、(IP, Port) のマップを取得し，ランダムに (IP, Port) を選択する
3. HTTPリクエストを (IP, Port) にフォワードする

実際には、Cookie を使って、Sticky を実現していたりもするけど割愛。

# ソースコード構成

https://github.com/cloudfoundry/gorouter

````
./router/main.go   # main() 関数
./router.go        # 全体をRouter構造体として管理
./server/server.go # HTTP Server と
./proxy/proxy.go   # Proxy の実装
./registry/registry.go # ルーティングテーブルの実装
````

## router.go

### NewRouter(c *config.Config) *Router

各種構造体の初期化。ルーティングテーブルは `StartPruningCycle()` を起動して自動削除処理ループを回しておく(後述)。

### Run()

メッセージバスへの接続とサブスクライブ設定(後述)。サブスクライブ設定をした後、しばらく待ってから(ルーティングテーブルのアップデート待ち)、プロキシサーバーを起動させてlistenを開始する。

#### RegisterComponent()

自分自身の情報をメッセージバスに送る。これによって、監視を自動化したりできる(今回は割愛)

#### SubscribeRegister()

> 2. Router が `router.register` トピックのメッセージを受け取り、{ドメイン名 => [(IP, Port)]} というマッピングを作る

この部分。"ドメイン名" は内部的に route.Uri という構造体で表現されていて、1個のメッセージで複数入れることが可能。


````
type registryMessage struct {
        Host string            `json:"host"`
        Port uint16            `json:"port"`
        Uris []route.Uri       `json:"uris"`
        Tags map[string]string `json:"tags"`
        App  string            `json:"app"`

        PrivateInstanceId string `json:"private_instance_id"`
}

func (r *Router) SubscribeRegister() {
        r.subscribeRegistry("router.register", func(registryMessage *registryMessage) {
                log.Debugf("Got router.register: %v", registryMessage)

                for _, uri := range registryMessage.Uris {
                        r.registry.Register(
                                uri,
                                makeRouteEndpoint(registryMessage),
                        )
                }
        })
}
````

(参考) https://github.com/cloudfoundry/gorouter/blob/da4613cf591a20e67c640884580468a35b97005f/route/uris.go 正規化するための `ToLower()` が追加実装された単なるstring

````
type Uri string

func (u Uri) ToLower() Uri {
        return Uri(strings.ToLower(string(u)))
}
````

#### SubscribeUnregister()

> 2. Router が `router.unregister` トピックのメッセージを受け取り、{ドメイン名 => [(IP, Port)]} というマッピングを削除する

この部分。

````
func (r *Router) SubscribeUnregister() {
        r.subscribeRegistry("router.unregister", func(registryMessage *registryMessage) {
                log.Debugf("Got router.unregister: %v", registryMessage)

                for _, uri := range registryMessage.Uris {
                        r.registry.Unregister(
                                uri,
                                makeRouteEndpoint(registryMessage),
                        )
                }
        })
}
````

#### SendStartMessage()

> なお、Router が `router.start` トピックのメッセージを送った場合にも、Agent がこのメッセージに対して、 `router.register` を応答するように設計されている。このメッセージのやりとりはメッセージバスの接続直後に行われ

この部分。

````
func (r *Router) SendStartMessage() {
        b, err := r.greetMessage()
        if err != nil {
                panic(err)
        }

        // Send start message once at start
        r.mbusClient.Publish("router.start", b)
}
````

#### HandleGreetings()

`router.greet` トピックのメッセージに対してレスポンスを返す。

#### ScheduleFlushApps(), flushApps(t time.Time)

アクティブな(一定時間内にアクセスのあった)アプリケーションをチェックして、メッセージバスに送信する関数。この目的は、CloudFoundryの別のコンポーネントが「アクティブなアプリケーションのみを残し、それ以外は停止させる」という仕様を満たすためで、この仕様により、開発環境などで頻繁に作成されるHello Worldアプリによるリソース枯渇を防ぐことができる(かもしれない)。

定期的に検査するための処理で `timeNewTicker()` を使っている。

## registry/registry.go

https://github.com/cloudfoundry/gorouter/blob/da4613cf591a20e67c640884580468a35b97005f/registry/registry.go

> {ドメイン名 => [(IP, Port)]} というマッピング

の実装の実態。実際には２つのmapを用いている。

### NewRegistry(c *config.Config, mbus yagnats.NATSClient) *Registry

ルーティングテーブルの初期化。

`r.byUri` が {URI => (IP, Port)の集合} のマッピングを持ち、 `r.table` が {tableKey => tableEntry} のマッピングを持つ。 `tableKey` は URIおよびエンドポイントを"IP:PORT"表現した文字列の組。 `tableEntry` はエンドポイントとその最終更新時刻。

ルーティングテーブルとして機能するのは `r.byUri` のほうで、 `r.table` のほうは、古くなったエンドポイントを自動削除するために使用される。

例:

````
r.byUri : {
   "www.example.com" => {
      "192.168.1.1:1234" => {(Endpoint)App1},
      "192.168.1.2:1234" => {(Endpoint)App2},
      "192.168.1.3:1234" => {(Endpoint)App1},
   }
}

r.table : {
  ("www.example.com", "192.168.1.1:1234") => ((Endpoint)App1, "2013/01/01 12:00:00")
  ("www.example.com", "192.168.1.2:1234") => ((Endpoint)App2, "2013/01/02 12:00:00")
  ("www.example.com", "192.168.1.3:1234") => ((Endpoint)App1, "2013/01/01 12:00:00")
}
````


````
type tableKey struct {
        addr string
        uri  route.Uri
}

type tableEntry struct {
        endpoint  *route.Endpoint
        updatedAt time.Time
}

func NewRegistry(c *config.Config, mbus yagnats.NATSClient) *Registry {
        r := &Registry{}

        r.Logger = steno.NewLogger("router.registry")

        r.ActiveApps = stats.NewActiveApps()
        r.TopApps = stats.NewTopApps()

        r.byUri = make(map[route.Uri]*route.Pool)

        r.table = make(map[tableKey]*tableEntry)

        r.pruneStaleDropletsInterval = c.PruneStaleDropletsInterval
        r.dropletStaleThreshold = c.DropletStaleThreshold

        r.messageBus = mbus

        return r
}
````

### Register(uri route.Uri, endpoint *route.Endpoint)

ルーティングテーブルへの登録。 `r.table` を先に参照し、そこに登録されているエンドポイント情報があれば、それを使って `r.byUri` に登録する。

````
func (registry *Registry) Register(uri route.Uri, endpoint *route.Endpoint) {
        registry.Lock()
        defer registry.Unlock()

        // (snip)

        entry, found := registry.table[key]
        if found {
                endpointToRegister = entry.endpoint
        } else {
                endpointToRegister = endpoint
                entry = &tableEntry{endpoint: endpoint}

                registry.table[key] = entry
        }

        pool, found := registry.byUri[uri]
        if !found {
                pool = route.NewPool()
                registry.byUri[uri] = pool
        }

        pool.Add(endpointToRegister)

        // (snip)
}
````

### Unregister(uri route.Uri, endpoint *route.Endpoint), unregisterUri(key tableKey)

ルーティングテーブルからの削除。 `r.byUri` からエントリを削除し(プールが空になればそれも削除)、さらに、`r.table` の情報も削除する。

````
func (registry *Registry) Unregister(uri route.Uri, endpoint *route.Endpoint) {
        registry.Lock()
        defer registry.Unlock()

        uri = uri.ToLower()

        key := tableKey{
                addr: endpoint.CanonicalAddr(),
                uri:  uri,
        }

        registry.unregisterUri(key)
}

func (registry *Registry) unregisterUri(key tableKey) {
        entry, found := registry.table[key]
        if !found {
                return
        }

        endpoints, found := registry.byUri[key.uri]
        if found {
                endpoints.Remove(entry.endpoint)

                if endpoints.IsEmpty() {
                        delete(registry.byUri, key.uri)
                }
        }

        delete(registry.table, key)
}
````

### Lookup(uri route.Uri) (*route.Endpoint, bool)

ルーティングテーブルの参照(Uriを使う)。

````
func (r *Registry) Lookup(uri route.Uri) (*route.Endpoint, bool) {
        r.RLock()
        defer r.RUnlock()

        pool, ok := r.lookupByUri(uri)
        if !ok {
                return nil, false
        }

        return pool.Sample()
}
````

### StartPruningCycle(), checkAndPrune()

古くなったエンドポイントを削除するためのルーチンの起動。

````
func (registry *Registry) StartPruningCycle() {
        go registry.checkAndPrune()
}

func (r *Registry) checkAndPrune() {
        if r.pruneStaleDropletsInterval == 0 {
                return
        }

        tick := time.Tick(r.pruneStaleDropletsInterval)
        for {
                select {
                case <-tick:
                        log.Debug("Start to check and prune stale droplets")
                        r.PruneStaleDroplets()
                }
        }
}
````

#### PruneStaleDroplets(), pruneStaleDroplets(), isEntryStale(entry *tableEntry)

古いエンドポイントの削除処理の実装。

````
func (registry *Registry) PruneStaleDroplets() {
        if registry.isStateStale() {
                log.Info("State is stale; NOT pruning")
                registry.pauseStaleTracker()
                return
        }

        registry.Lock()
        defer registry.Unlock()

        registry.pruneStaleDroplets()
}


func (registry *Registry) pruneStaleDroplets() {
        for key, entry := range registry.table {
                if !registry.isEntryStale(entry) {
                        continue
                }

                log.Infof("Pruning stale droplet: %v, uri: %s", entry, key.uri)
                registry.unregisterUri(key)
        }
}

func (r *Registry) isEntryStale(entry *tableEntry) bool {
        return entry.updatedAt.Add(r.dropletStaleThreshold).Before(time.Now())
}
````

