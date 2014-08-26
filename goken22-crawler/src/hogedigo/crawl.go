package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"code.google.com/p/go.net/html"
)

/*
	goken課題：gopher crawler
	usage: crawl [-d <depth>] [-o <output dir>] <url>
	再帰クロールをgoroutineでパラレルにした。
	redirect先チェック。
 */
var attrNameMap map[string]string = map[string]string{"a": "href", "img": "src"}

type Tree struct {
	Url      string `json:",omitempty"`
	Error    string `json:",omitempty"`
	Children []*Tree
	Tat      int64 `json:",omitempty`
}

type syncset struct {
	s map[string]struct{}
	sync.Mutex
}

func newSyncset() *syncset {
	var set syncset
	set.s = make(map[string]struct{})
	return &set
}

func (set *syncset) put(aUrl string) bool {
	lurl := strings.ToLower(aUrl)
	set.Lock()
	defer set.Unlock()
	if _, exists := set.s[lurl]; exists {
		return false
	}
	set.s[lurl] = struct{}{}
	return true
}

var outdir string

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var depth int
	flag.IntVar(&depth, "d", 10, "depth")
	flag.StringVar(&outdir, "o", "out", "output dir")
	flag.Parse()

	os.MkdirAll(outdir, os.ModeDir|0766)

	aUrl := flag.Arg(0)
	if aUrl == "" {
		log.Fatalf("usage: crawl [-d <depth>] [-o <output dir>] <url>")
	}

	set := newSyncset()
	ch := pcrawl(aUrl, depth-1, set)

	result, err := json.MarshalIndent(<-ch, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
		return
	}

	fmt.Printf("done. result=%s\n", result)
}

func pcrawl(aUrl string, depth int, set *syncset) <-chan *Tree {

	ch := make(chan *Tree)
	go func() {
		ch <- _pcrawl(aUrl, depth, set)
	}()
	return ch
}

func _pcrawl(aUrl string, depth int, set *syncset) *Tree {

	tree := Tree{Url: aUrl}

	fmt.Println(tree.Url)

	purl, err := url.Parse(tree.Url)
	if err != nil {
		tree.Error = err.Error()
		return &tree
	}

	start := time.Now().UnixNano()
	defer func(tree *Tree) {
		tree.Tat = time.Now().UnixNano() - start
	}(&tree)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if purl.Host != req.URL.Host || !set.put(req.URL.String()){
				return errors.New("redirect ignored. url:" + req.URL.String())
			}
			return nil
		},
	}
	resp, err := client.Get(tree.Url)
	if err != nil {
		tree.Error = err.Error()
		return &tree
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "image/") {
		if err := writeImage(purl.Path, resp.Body); err != nil {
			tree.Error = err.Error()
		}
		return &tree
	}

	if !strings.HasPrefix(contentType, "text/html") {
		return &tree
	}

	resources := parseHTML(resp.Body)

	futureChildren := make([]<-chan *Tree, 0)

	for r := range resources {
		rurl, err := purl.Parse(r.src)
		if err != nil {
			tree.Error = err.Error()
			return &tree
		}

		if !set.put(rurl.String()) {
			continue
		}

		if purl.Host != rurl.Host {
			continue
		}

		if r.tag == "a" {
			lpath := strings.ToLower(rurl.Path)
			if !strings.HasSuffix(lpath, ".html") && !strings.HasSuffix(lpath, ".htm") && !strings.HasSuffix(lpath, "/") {
				continue
			}
			if depth <= 0 {
				childCh := make(chan *Tree)
				go func() {
					childCh <- &Tree{Url: rurl.String()}
				}()
				futureChildren = append(futureChildren, childCh)
				continue
			}
		}

		childCh := pcrawl(rurl.String(), depth-1, set)
		futureChildren = append(futureChildren, childCh)
	}

	for _, ch := range futureChildren {
		tree.Children = append(tree.Children, <-ch)
	}

	return &tree
}

type resource struct {
	tag string
	src string
}

func parseHTML(r io.Reader) <-chan resource {
	ch := make(chan resource)

	go _parseHTML(r, ch)

	return ch
}

func _parseHTML(r io.Reader, ch chan<- resource) {

	defer func() {
		close(ch)
	}()

	z := html.NewTokenizer(r)

	findAttr := func(name string) string {
		lname := strings.ToLower(name)
		moreAttr := true
		for moreAttr {
			var key, val []byte
			key, val, moreAttr = z.TagAttr()
			if strings.ToLower(string(key)) == lname {
				return strings.Split(string(val), "#")[0]
			}
		}
		return ""
	}

	for {
		tokenType := z.Next()
		switch tokenType {
		case html.ErrorToken:
			return
		case html.StartTagToken, html.SelfClosingTagToken:
			tagName, hasAttr := z.TagName()
			if !hasAttr {
				continue
			}
			ltag := strings.ToLower(string(tagName))
			attrName, ok := attrNameMap[ltag]
			if !ok {
				continue
			}
			if attr := findAttr(attrName); attr != "" {
				ch <- resource{ltag, attr}
			}
		default:
		}
	}
}

func writeImage(aUrl string, r io.Reader) error {
	fname := strings.Replace(aUrl, "/", ".", -1)
	if fname[0] == '.' {
		fname = fname[1:]
	}

	f, err := os.Create(path.Join(outdir, fname))
	if err != nil {
		return err
	}
	defer f.Close()

	bufw := bufio.NewWriter(f)
	defer bufw.Flush()

	bufr := bufio.NewReader(r)

	if _, err := bufw.ReadFrom(bufr); err != nil {
		return err
	}
	return nil
}
