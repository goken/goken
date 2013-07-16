# Docker Overview

## こんな人にお薦め

- Linux 使ってるけど vagrant とか Virtual Box とか時間かかって面倒。
- LXC 使ってるけどコンテナの管理面倒。
- VM 作りまくって何が何だかわからなくなった。
- Go のソース読みたい

## 事前準備

note: Ubuntu 12.04 (linux-image-3.8.0-26-generic) 使ってます。 Kernel 3.5 以前では安定しない模様。

### Ubuntu 12.04 / Kernel 3.8.0-26-generic の準備

````
$ cd path/to/this_dir
$ vagrant box add precise64 http://dl.dropbox.com/u/1537815/precise64.box
$ vagrant up

vagrant$ sudo apt-get update
vagrant$ sudo apt-get install -q -y linux-image-3.8.0-26-generic
vagrant$ exit

$ vagrant reload
...
The following SSH command responded with a non-zero exit status.
Vagrant assumes that this means the command failed!

mount -t vboxsf -o uid=`id -u vagrant`,gid=`id -g vagrant` v-root /vagran
````

`mount -t vboxsf ...` でエラーがでるがとりあえず無視(Kernelをアップデートしたので、vboxのリビルドが必要)

### Docker のインストール

````
$ vagrant ssh

vagrant$ sudo apt-get install -q -y curl
vagrant$ curl get.docker.io | sudo sh -x
vagrant$ ps aux | grep docker
root      3409  0.0  0.1 273064  5064 ?        Ssl  07:40   0:00 /usr/local/bin/docker -d
````

note: この手順だとサービスの自動起動を登録しないので、リブートしたときは `start dockerd` で起動させる

## いろいろ試す

### 最初のコマンド

Docker で /bin/echo をコンテナで起動する

````
vagrant$ docker pull base
vagrant$ docker run base /bin/echo Hello World
Hellow World
````

### コンテナのホスト名を確認

コンテナ環境が毎回変わっていることを確認

````
vagrant$ docker run base /bin/hostname
550cf7ca6e98
vagrant$ docker run base /bin/hostname
f2aa47f6ea1c
````

### IPアドレスは....?

````
vagran$ docker run base /sbin/ip addr | grep eth0
21: eth0: <NO-CARRIER,BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state DOWN qlen 1000
    inet 172.16.42.7/24 brd 172.16.42.255 scope global eth0
vagran$ docker run base /sbin/ip addr | grep eth0
24: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UP qlen 1000
    inet 172.16.42.8/24 brd 172.16.42.255 scope global eth0
vagran$ docker run base /sbin/ip addr | grep eth0
27: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UNKNOWN qlen 1000
    inet 172.16.42.9/24 brd 172.16.42.255 scope global eth0
````

どんどん 172.16.42.0/24 の範囲内でインクリメントされていきます。

````
510: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UNKNOWN qlen 1000
    inet 172.16.42.254/24 brd 172.16.42.255 scope global eth0
512: eth0: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state UNKNOWN qlen 1000
    inet 172.16.42.2/24 brd 172.16.42.255 scope global eth0
514: eth0: <NO-CARRIER,BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc pfifo_fast state DOWN qlen 1000
````

## もうちょっと試す

### /bin/bash を実行する

インタラクティブシェルを起動する(tty にアタッチするので root 権限が必要)。

````
vagrant$ docker run -i -t base /bin/bash
root@b51e604d11b6:/# ps aux
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.1  0.0  18056  1940 ?        S    04:23   0:00 /bin/bash
root        17  0.0  0.0  15528  1124 ?        R+   04:23   0:00 ps aux
````

コンテナ内のプロセスしか見えない。

### コンテナをバックグラウンドで動かす

````
vagrant$ sudo docker run -i -t -d base /bin/bash
4774e1bbbd18
vagratn$ docker ps
ID                  IMAGE               COMMAND             CREATED              STATUS              PORTS
4774e1bbbd18        base:latest         /bin/bash           About a minute ago   Up About a minute
````

普通のpsコマンドでも見ることができる

````
vagrant$ ps aux | grep bash
vagrant   1132  0.0  0.2  29364  8356 pts/0    Ss   07:34   0:00 -bash
root     21154  0.0  0.0  21160  1168 pts/3    Ss   08:04   0:00 lxc-start -n 3dc2f42b0b4ebc0eacada31f2f71703a899dcefa58ac621b3ec3afe4d7b4e610 -f /var/lib/docker/containers/3dc2f42b0b4ebc0eacada31f2f71703a899dcefa58ac621b3ec3afe4d7b4e610/config.lxc -- /sbin/init -g 172.16.42.1 -e TERM=xterm -e HOME=/ -e PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin -- /bin/bash
````

### コンテナにアタッチする

````
vagrant$ docker attach 4774e1bbbd18
[Enter]
root@4774e1bbbd18:/#
root@4774e1bbbd18:/# ps aux
USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root         1  0.0  0.0  18064  1880 ?        S    08:06   0:00 /bin/bash
root        11  0.0  0.0  15532  1128 ?        R+   08:07   0:00 ps aux

root@4774e1bbbd18:/# Ctrl+p Ctrl+q  # dettach する
````

### Node.js サーバーを動かす

#### まずはアプリとDockerfileを作る

````
vagrant$ mkdir nodeapp
vagrant$ cd nodeapp
vagrant$ mkdir src
vagrant$ vi app.js
var PORT = 8080;
var server = require('http').createServer(function(req, res){
  res.end('Hello World\n');
});
server.listen(PORT)
console.log('Running on port:' + PORT);
````

````
vagrant$ vi Dockerfile
FROM    base
RUN     apt-get update -o "Acquire::http::proxy=xxxxx"
RUN     apt-get install -q -y -o "Acquire::http::proxy=xxxxx" nodejs

ADD     /home/vagrant/nodeapp/src /src
EXPOSE  8080
CMD     ["/usr/bin/nodejs", "/src/app.js"]
````

(proxy 環境でやるときは `apt-get` に `-o "Acquire::http::proxy=xxxxx"` オプションをつける)

Dockerfile で使えるコマンド: http://docs.docker.io/en/latest/use/builder/#id1

- FROM: ベースにするイメージ
- RUN: ビルド時に実行するコマンド
- ADD: ビルド時にファイル/ディレクトリをコピーするコマンド
- EXPOSE: イメージ実行時にNATするポート
- CMD: イメージ実行時に実行するコマンド

#### Docker のイメージとしてビルドする

````
vagrant$ docker build -t yssk22/node-hello .
Uploading context 10240 bytes
Step 1 : FROM base
 ---> b750fe79269d
Step 2 : RUN apt-get update
 ---> Running in 9b336a266d7c
 ---> 64bcd9d40c26
Step 3 : RUN apt-get install -q -y nodejs
 ---> Running in 3554ed296982
 ---> aa27b6706638
Step 4 : ADD . /src
 ---> 659e961143c5
Step 5 : EXPOSE 8080
 ---> Running in b7af7365a58c
 ---> 5b6e9916cc39
Step 6 : CMD ["/usr/bin/nodejs", "/src/app.js"]
 ---> Running in fcfdd87a6b97
 ---> c01203ff6be0
Successfully built c01203ff6be0
````

````
vagrant$ docker images
REPOSITORY          TAG                 ID                  CREATED              SIZE
base                latest              b750fe79269d        3 months ago         24.65 kB (virtual 180.1 MB)
base                ubuntu-12.10        b750fe79269d        3 months ago         24.65 kB (virtual 180.1 MB)
base                ubuntu-quantal      b750fe79269d        3 months ago         24.65 kB (virtual 180.1 MB)
base                ubuntu-quantl       b750fe79269d        3 months ago         24.65 kB (virtual 180.1 MB)
yssk22/node-hello   latest              c01203ff6be0        About a minute ago   12.29 kB (virtual 309.8 MB)
````

#### 起動する

````
# docker run -d yssk22/node-hello
923c56817dac
# docker ps
ID                  IMAGE                      COMMAND                CREATED             STATUS              PORTS
344d6d56fb46        yssk22/node-hello:latest   /usr/bin/nodejs /src   2 seconds ago       Up 1 seconds        49167->8080
````

note: docker ps で表示されない場合は -d を削除して interactive で起動してみるとよい

ホストのランダムポートからEXPOSEで指定したポートへのマッピング済みなのでホストから確認できる

````
# curl http://localhost:49167/
Hello World
````

## 気になるところを試す

### ネットワーク

````
vagrant$ netstat -an | grep 49167
tcp        0      0 127.0.0.1:49167         0.0.0.0:*               LISTEN
vagrant$ sudo lsof | grep 49167
docker    1362        root   16u     IPv4              18332      0t0        TCP localhost:49167 (LISTEN)
vagrant$ ps aux | grep docker
root      1362  0.9  0.4 276900  9528 pts/2    Ssl+ 04:47   0:20 /usr/local/bin/docker -d
vagrant$ sudo iptables -t nat -L
Chain PREROUTING (policy ACCEPT)
target     prot opt source               destination
DOCKER     all  --  anywhere             anywhere             ADDRTYPE match dst-type LOCAL

Chain INPUT (policy ACCEPT)
target     prot opt source               destination

Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination
DOCKER     all  --  anywhere            !127.0.0.0/8          ADDRTYPE match dst-type LOCAL

Chain POSTROUTING (policy ACCEPT)
target     prot opt source               destination
MASQUERADE  all  --  10.0.3.0/24         !10.0.3.0/24
MASQUERADE  all  --  172.16.42.0/24      !172.16.42.0/24

Chain DOCKER (2 references)
target     prot opt source               destination
DNAT       tcp  --  anywhere             anywhere             tcp dpt:49153 to:172.16.42.6:8080
````

DNATでランダムポートからEXPOSEしたポートにトラフィックを流している模様。

## Docker を支える技術

### LXC / cgroup / namespacing

#### LXC

Linux Container。Linux にコンテナ環境を提供する。

- アプリケーションコンテナ: プロセスを動かす環境
- システムコンテナ: OSを動かす環境 (カーネルはコンテナ間でシェア)

#### cgroup

ユーザー定義のプロセスグループにリソース(CPU時間, メモリ, IO, ...etc)を割り当てる仕組み。

#### namespacing

プロセスグループを隔離する仕組み

- 他のプロセスグループとPID空間を分離する
- 他のプロセスグループとネットワーク空間を分離する
- 他のプロセスグループとファイルシステム空間を分離する
- ...etc

これにより、コンテナ毎に"同じポートやファイルシステムツリーで"サーバーを立てることが可能。

### AUFS

Another Union File System. Linux で Union Mount を可能にするファイルシステム。

### Union Mount

1つのマウントポイントで複数のディスクデバイスを扱えるようにする。

````
/mnt/disk
  + /dev/sda1
  + /dev/sdb1
````

複数のデバイスは階層化させることができるので、

````
/mnt/disk
  /dev/sda1 このディスクにベースのイメージを保存する
  /dev/sdb1 このディスクに起動後に更新のあったファイルを書き込む
````

といった感じで使うことができる(Dockerの実装は未確認)

- Docker ではこれを利用してコンテナイメージファイルの履歴管理をする
  - `docker commit`: 更新をコミットして元のイメージを更新することも出来る
  - `docker diff`: コンテナとイメージのdiffを取得できる
  - `docker history`: イメージの更新履歴を取得できる

## とりあえずソース読もう

commit: 9b57f9187b84c5cdb92cb50271988aa4c51e8b95 より

- GIT_REPO_ROOT/ が docker パッケージ
- GIT_REPO_ROOT/docker/ が main パッケージの構成
  main 内で

  ````
  import (

        "github.com/dotcloud/docker"
        "github.com/dotcloud/docker/utils"

  )
  ````

  のようなことをしている

````
vagrant$ $ docker -h
Usage of docker:
  -D=false: Debug mode
  -H=[tcp://127.0.0.1:4243]: tcp://host:port to bind/connect to or unix://path/to/socket to use
  -api-enable-cors=false: Enable CORS requests in the remote api.
  -b="": Attach containers to a pre-existing network bridge
  -d=false: Daemon mode
  -dns="": Set custom dns servers
  -g="/var/lib/docker": Path to graph storage base dir.
  -p="/var/run/docker.pid": File containing process PID
  -r=false: Restart previously running containers
````

となるので、コマンドラインもサーバーも同じmainからスイッチする模様。

### ブリッジ設定あれこれ

先のネットワークの確認では 172.16.42.0/24 というネットワークが勝手に作られ、利用されていた。これはどうなっているのか。

````
  -b="": Attach containers to a pre-existing network bridge
````

が気になるので読み進める。

#### ./docker/docker.go: main()関数。

各種引数でオプションを処理。Dockerそのものに設定ファイルのようなものは存在しない。必要に応じてdockerパッケージのパッケージ変数を設定していく。

````
  bridgeName := flag.String("b", "", "Attach containers to a pre-existing network bridge")
...
  if *bridgeName != "" {
    docker.NetworkBridgeIface = *bridgeName
  } else {
    docker.NetworkBridgeIface = docker.DefaultNetworkBridge
  }
````

-b オプションで利用するブリッジを設定可能、デフォルトでは docker.DefaultNetworkBridge (docker0) が使われる。

#### ./runtime.go: Docker 環境を管理するコード。Runtime 構造体およびその関数。

````
 22 type Runtime struct {
 23         root           string
 24         repository     string
 25         containers     *list.List
 26         networkManager *NetworkManager
 27         graph          *Graph
 28         repositories   *TagStore
 29         idIndex        *utils.TruncIndex
 30         capabilities   *Capabilities
 31         kernelVersion  *utils.KernelVersionInfo
 32         autoRestart    bool
 33         volumes        *Graph
 34         srv            *Server
 35         Dns            []string
 36 }
````

#### ./network.go ネットワーク管理。NetworkManager 構造体およびその関数。

````
539 // Network Manager manages a set of network interfaces
540 // Only *one* manager per host machine should be used
541 type NetworkManager struct {
542         bridgeIface   string
543         bridgeNetwork *net.IPNet
544
545         ipAllocator   *IPAllocator
546         portAllocator *PortAllocator
547         portMapper    *PortMapper
548 }
````

NetworkManager は専用にBridgeを利用する(渡された名前のものがなければ作る)

````
564 func newNetworkManager(bridgeIface string) (*NetworkManager, error) {
565         addr, err := getIfaceAddr(bridgeIface)
566         if err != nil {
567                 // If the iface is not found, try to create it
568                 if err := CreateBridgeIface(bridgeIface); err != nil {
569                         return nil, err
570                 }
571                 addr, err = getIfaceAddr(bridgeIface)
572                 if err != nil {
573                         return nil, err
574                 }
575         }
````

決め打ちで "172.16.42.1/24", "10.0.42.1/24", "192.168.42.1/24" が順に使われる模様。ルーティング情報が設定されていなければ勝手に使ってしまうので、注意。

````
115 func CreateBridgeIface(ifaceName string) error {
116         // FIXME: try more IP ranges
117         // FIXME: try bigger ranges! /24 is too small.
118         addrs := []string{"172.16.42.1/24", "10.0.42.1/24", "192.168.42.1/24"}
119
120         var ifaceAddr string
121         for _, addr := range addrs {
122                 _, dockerNetwork, err := net.ParseCIDR(addr)
123                 if err != nil {
124                         return err
125                 }
126                 if err := checkRouteOverlaps(dockerNetwork); err == nil {
127                         ifaceAddr = addr
128                         break
129                 } else {
130                         utils.Debugf("%s: %s", addr, err)
131                 }
132         }
133         if ifaceAddr == "" {
134                 return fmt.Errorf("Could not find a free IP address range for interface '%s'. Please configure its address     manually and run 'docker -b %s'", ifaceName, ifaceName)
135         }
136         utils.Debugf("Creating bridge %s with network %s", ifaceName, ifaceAddr)
````

ifaceName と ifaceAddr が決まったら ip コマンドでブリッジを作ってMASQUERADE設定をかける

````
137
138         if output, err := ip("link", "add", ifaceName, "type", "bridge"); err != nil {
139                 return fmt.Errorf("Error creating bridge: %s (output: %s)", err, output)
140         }
141
142         if output, err := ip("addr", "add", ifaceAddr, "dev", ifaceName); err != nil {
143                 return fmt.Errorf("Unable to add private network: %s (%s)", err, output)
144         }
145         if output, err := ip("link", "set", ifaceName, "up"); err != nil {
146                 return fmt.Errorf("Unable to start network bridge: %s (%s)", err, output)
147         }
148         if err := iptables("-t", "nat", "-A", "POSTROUTING", "-s", ifaceAddr,
149                 "!", "-d", ifaceAddr, "-j", "MASQUERADE"); err != nil {
150                 return fmt.Errorf("Unable to enable network bridge NAT: %s", err)
151         }
152         return nil
153 }
````

ここまでがブリッジ用インターフェースの操作。別にPythonでもRubyでもできる。 `ip(args ...string)` とか `iptales(args ...string)` は `exec.Command(path, args ...string)` 呼んでるだけ。

#### (参考): ルーティングの確認

````
 96 func checkRouteOverlaps(dockerNetwork *net.IPNet) error {
 97         output, err := ip("route")
 98         if err != nil {
 99                 return err
100         }
101         utils.Debugf("Routes:\n\n%s", output)
102         for _, line := range strings.Split(output, "\n") {
103                 if strings.Trim(line, "\r\n\t ") == "" || strings.Contains(line, "default") {
104                         continue
105                 }
106                 if _, network, err := net.ParseCIDR(strings.Split(line, " ")[0]); err != nil {
107                         return fmt.Errorf("Unexpected ip route output: %s (%s)", err, line)
108                 } else if networkOverlaps(dockerNetwork, network) {
109                         return fmt.Errorf("Network %s is already routed: '%s'", dockerNetwork.String(), line)
110                 }
111         }
112         return nil
113 }
````

### EXPOSE?

Dockerfile に `EXPOSE  8080` と書いてポートマッピングをすることができたが、これはどう動いているのか?

#### ./container.go

LXC のコンテナ操作を司るファイル。 `Container` という名前の通りの struct と `Config` という struct で管理される(使い分けは不明で、 Container インスタンスが Config インスタンスへの参照を持つ。

````
715 func (container *Container) allocateNetwork() error {
716         iface, err := container.runtime.networkManager.Allocate()
717         if err != nil {
718                 return err
719         }
720         container.NetworkSettings.PortMapping = make(map[string]PortMapping)
721         container.NetworkSettings.PortMapping["Tcp"] = make(PortMapping)
722         container.NetworkSettings.PortMapping["Udp"] = make(PortMapping)
723         for _, spec := range container.Config.PortSpecs {
724                 nat, err := iface.AllocatePort(spec)
725                 if err != nil {
726                         iface.Release()
727                         return err
728                 }
729                 proto := strings.Title(nat.Proto)
730                 backend, frontend := strconv.Itoa(nat.Backend), strconv.Itoa(nat.Frontend)
731                 container.NetworkSettings.PortMapping[proto][backend] = frontend
732         }
````

#### ./network.go

ポートのアサイン

````
459 func (iface *NetworkInterface) AllocatePort(spec string) (*Nat, error) {
460         nat, err := parseNat(spec)
461         if err != nil {
462                 return nil, err
463         }
464
465         if nat.Proto == "tcp" {
466                 extPort, err := iface.manager.tcpPortAllocator.Acquire(nat.Frontend)
467                 if err != nil {
468                         return nil, err
469                 }
470                 backend := &net.TCPAddr{IP: iface.IPNet.IP, Port: nat.Backend}
471                 if err := iface.manager.portMapper.Map(extPort, backend); err != nil {
472                         iface.manager.tcpPortAllocator.Release(extPort)
473                         return nil, err
474                 }
475                 nat.Frontend = extPort
````

そして、NATの設定を加えて、さらにTCP Proxyを起動する(TCP Proxyなんで? 127.0.0.1 用にみえる)

````
226 func (mapper *PortMapper) Map(port int, backendAddr net.Addr) error {
227         if _, isTCP := backendAddr.(*net.TCPAddr); isTCP {
228                 backendPort := backendAddr.(*net.TCPAddr).Port
229                 backendIP := backendAddr.(*net.TCPAddr).IP
230                 if err := mapper.iptablesForward("-A", port, "tcp", backendIP.String(), backendPort); err != nil {
231                         return err
232                 }
233                 mapper.tcpMapping[port] = backendAddr.(*net.TCPAddr)
234                 proxy, err := NewProxy(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: port}, backendAddr)
235                 if err != nil {
236                         mapper.Unmap(port, "tcp")
237                         return err
238                 }
239                 mapper.tcpProxies[port] = proxy
240                 go proxy.Run()
````

こちらは単なるプロキシ。

````
 98 func (proxy *TCPProxy) Run() {
 99         quit := make(chan bool)
100         defer close(quit)
101         utils.Debugf("Starting proxy on tcp/%v for tcp/%v", proxy.frontendAddr, proxy.backendAddr)
102         for {
103                 client, err := proxy.listener.Accept()
104                 if err != nil {
105                         utils.Debugf("Stopping proxy on tcp/%v for tcp/%v (%v)", proxy.frontendAddr, proxy.backendAddr, err.Error())
106                         return
107                 }
108                 go proxy.clientLoop(client.(*net.TCPConn), quit)
109         }
110 }
111
````

### サブコマンドってどうなってるの?

````
vagrant$ docker
Usage: docker [OPTIONS] COMMAND [arg...]

Commands:
    attach    Attach to a running container
    build     Build a container from a Dockerfile
    ....
````

となっているがサブコマンドの実装は? (オプションの実装は `flag` パッケージ使ってるだけ)

#### ./commands.go

````
func (cli *DockerCli) CmdAttach(args ...string) error {
````

のような形で `CmdXXXX(args ...string) error` の関数を実装している。 `reflect` パッケージで呼び出す。

````
36 func (cli *DockerCli) getMethod(name string) (reflect.Method, bool) {
37   methodName := "Cmd" + strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
38   return reflect.TypeOf(cli).MethodByName(methodName)
39 }
````

コマンド自体は HTTP で Docker サーバーとやりとりをする。 下記は `attach` コマンドでの例。

````
1148         if err := cli.hijack("POST", "/containers/"+cmd.Arg(0)+"/attach?"+v.Encode(), container.Config.Tty, cli.in, cli.out); err != nil {
1149                 return err
1150         }
1151         return nil
````

`hijack` 関数は内部で httputil.NewClientConn のインスタンスに対して `Hijack()` を呼ぶ。 http://golang.org/pkg/net/http/httputil/#ClientConn.Hijack によれば、keep-alive のロジックでIOを扱うらしい(WebSocketとかいうことはしてない)。 `CmdLogs` なども同様に `Hijack()` を使って実装されているので、 docker クライアント-サーバー間は、

- HTTP Keep-alive が有効な経路であること
- HTTP を許可しても、途中で接続を切ってしまうようなFWがいないこと

などが必要とされる模様。 `attach` のサーバー側の実装は server.go にある (Container の stdin/out をプロキシしてるだけ)。


