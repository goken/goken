Go研 vol0. まとめ
==================

##参加者

* [Jxck](https://twitter.com/Jxck_)
* [tenntenn](https://twitter.com/tenntenn)
* [manji0112](https://twitter.com/manji0112)
* [hogedigo](https://twitter.com/hogedigo)
* [yssk22](https://twitter.com/yssk22)

##今回の概要

* 開催日：2013年04月24日(水)
* connpass：http://connpass.com/event/2222/
* 発表者：Jxck
* バージョン：go1.0.3
* パッケージ：net/http

##話題に上がった点

###ServerMux型のexplicitは何なのか？
ServerMux型は

	type ServeMux struct {
		mu sync.RWMutex
		m  map[string]muxEntry
	}

	type muxEntry struct {
		explicit bool
		h        Handler
	}

と宣言されている。
このmuxEntry型のexplicitというフィールドは何の為に使用しているのか？という点が話題に挙がった。

ServerMux.Handleの定義（server.go:944）は以下のようになっている。

	// Handle registers the handler for the given pattern.
	// If a handler already exists for pattern, Handle panics.
	func (mux *ServeMux) Handle(pattern string, handler Handler) {
 		mux.mu.Lock()
		defer mux.mu.Unlock()
		
		if pattern == "" {
			panic("http: invalid pattern " + pattern)
		}
		if handler == nil {
			panic("http: nil handler")
		}
		if mux.m[pattern].explicit {
			panic("http: multiple registrations for " + pattern)
		}

		mux.m[pattern] = muxEntry{explicit: true, h: handler}

		// Helpful behavior:
		// If pattern is /tree/, insert an implicit permanent redirect for /tree.
		// It can be overridden by an explicit registration.
		n := len(pattern)
		if n > 0 && pattern[n-1] == '/' && !mux.m[pattern[0:n-1]].explicit {
		mux.m[pattern[0:n-1]] = muxEntry{h: RedirectHandler(pattern, StatusMovedPermanently)}
		}
	}

これを見ると、すでに指定したパタンにハンドラが設定してあるかどうかをチェックするフラグとして使用されている。
しかし、mapはキーが「ある」か「ない」かを[チェックできる](http://play.golang.org/p/Msxq9bIhMn)。
そのため、このフィールドが指す意味が良くわからなかった。
ちなみに、explicitを設定しているのは、上記の

	mux.m[pattern] = muxEntry{explicit: true, h: handler}

の部分だけだった。

追記：
鵜飼さんから、Google+のコミュニティでexplicitの意味を教えていただきました。
https://plus.google.com/u/0/117100596700604439455/posts/bowpjrpButJ

###ResponseWriterがinterfaceな理由

http.Handler.ServeHTTPの引数には、http.Request型とhttp.ResponseWriter型が使われている。http.Request型がstructであるのに対し、http.ResponseWriter型がinterfaceである理由について、話題に挙がった。

http.ResponseWriter型がinterfaceである理由は、Handlerのテストをするためではないかと推測できる。理由は、net/http/httptestパッケージのhttptest.ResponseRecorder型を見れば分かる。httptest.ResponseRecorder型では、Handlerの返したレスポンスを記録し、あとで取得することができる。そして、返したレスポンスを取得し、想定していたものと同じものがレスポンスされたかどうかチェックすることができる。

このように、レスポンスについては、本番用は通常通りクライアントサイドにレスポンスする。一方で、単体テストでは、httptest.ResponseRecorderの内部にレスポンスを保持することで、テストを容易にしている。

### importについて

import文について以下の2点が話題に挙がった。

#### import . "パッケージ"
「.」を使うことで、Javaのstaticインポートのようなことができる。つまり、「.」で別名を付けたパッケージのメンバはパッケージ名無しで呼び出すことができる（[参考](http://play.golang.org/p/--dWV6PHYA)）。

#### import _ "パッケージ"
「_」をつけることで、未使用のパッケージがあってもコンパイルエラーを出さなくて済む（[参考](http://play.golang.org/p/EcgE0plkD9)）。
また、「_」を付けてインポートしても、そのパッケージのinit関数は呼ばれる（[参考](https://github.com/golang-samples/basic/tree/master/import/underscore)）。

###http.DefaultServeMuxについて
http.Server型の定義は以下のようになっている(server.go:997)。

 	// A Server defines parameters for running an HTTP server.
	type Server struct {
		Addr           string        // TCP address to listen on, ":http" if empty
		Handler        Handler       // handler to invoke, http.DefaultServeMux if nil
		ReadTimeout    time.Duration // maximum duration before timing out read of the request
		WriteTimeout   time.Duration // maximum duration before timing out write of the response
		MaxHeaderBytes int           // maximum size of request headers, DefaultMaxHeaderBytes if 0
		TLSConfig      *tls.Config   // optional TLS config, used by ListenAndServeTLS
	}

http.ListenAndServe関数では、このうちのhttp.Server.Addrをhttp.Server.Handlerを以下のように設定している。

	func ListenAndServe(addr string, handler Handler) error {
		server := &Server{Addr: addr, Handler: handler}
		return server.ListenAndServe()
	}

ここで設定されたHandlerがリクエストを捌く際に使用される（[参考](http://www.slideshare.net/takuyaueda967/go-webgocon-2013-spring)）。この値がnilである場合は、代わりにhttp.DefaultServeMuxが使用される。なお、http.Handle関数などは、このhttp.DefaultServeMuxに対してのラッパーである。そのため、http.Server.Handlerがnilでない場合は、http.Server.Handlerに設定した値が優先されるため、http.Handleやhttp.HandleFuncでいくらハンドラを設定しても意味がない。
