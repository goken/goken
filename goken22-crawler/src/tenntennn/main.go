package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
)

type Task interface {
	Do()
	Wait() error
}

type task struct {
	fnc     func(line string) error
	done    chan struct{}
	url     string
	err     error
	crawler *Crawler
}

func newTask(downloadURL string, crawler *Crawler, fnc func(line string) error) *task {
	return &task{
		fnc:     fnc,
		url:     downloadURL,
		done:    make(chan struct{}),
		crawler: crawler,
	}
}

func (t *task) Do() {
	t.done = make(chan struct{})
	defer func() {
		t.done <- struct{}{}
	}()

	resp, err := http.Get(downloadURL)
	if err != nil {
		t.err = err
		return
	}
	defer resp.Body.Close()

	scanner := bufio.Scanner(resp.Body)
	for scanner.Scan() {
		if err := t.fnc(t, scanner.Text()); err != nil {
			t.err = err
			return
		}
	}
	if err := scanner.Err(); err != nil {
		t.err = err
		return
	}
}

func (t *task) Wait() error {
	<-t.done
	return t.err
}

type Crawler struct {
	workerCh chan Task
}

func NewCrawler(numWorker int) *Crawler {
	return &Crawler{
		workerCh: make(chan Task, numWorker),
	}
}

func (c *Crawler) Crawl(downloadURL string) {
	go func() {
		for task := range c.workerCh {
			go task.Do()
		}
	}()
	task := c.NewHtmlTask(downloadURL)
	c.workerCh <- task
	if err := task.Wait(); err != nil {
		log.Println("ERROR: " + err.Error())
	}
}

const (
	aRegexp     = regexp.MustCompile(`<a.+ href="([^"]+)".*>`)
	imgRegexp   = regexp.MustCompile(`<img.+ src="([^"]+)".*>`)
	cssRegexp   = regexp.MustCompile(`<link.+ href="([^"]+\.css)[^"]*".*>`)
	urlRegexp   = regexp.MustCompile(`url\("?([^"()?]+)\?.+?"?\)`)
	imageRegexp = regexp.MustCompile(`png|jpe?g`)
)

func fullURL(htmlURL, relUrl string) (bool, string) {
	parsedHtmlURL := url.Parse(htmlURL)
	if err != nil {
		return false, ""
	}

	parsedURL, err := url.Parse(relUrl)
	if err == nil {
		if parsedURL.Scheme != parsedHtmlURL.Scheme || parsedURL.Host != parsedHtmlURL.Host {
			return false, ""
		}

		return true, relUrl
	}

	absURL := parsedHtmlURL.Scheme + "://" + parsedHtmlURL.HOST + "/" + relUrl
	return true, absURL
}

func htmlParse(t *task, line string) error {
	// imgタグ
	if match := imgRegexp.FindStringSubmatch(line); len(match) > 0 {
		if ok, imgUrl := fullURL(match[1]); !ok {
			return nil
		}

		task := t.crawler.NewImageTask(imgUrl)
		t.crawler.workerCh <- task
		return task.Wait()
	}

	// aタグ
	if match := aRegexp.FindStringSubmatch(line); len(match) > 0 {
		if ok, htmlUrl := fullURL(match[1]); !ok {
			return nil
		}

		task := t.crawler.NewHtmlTask(htmlUrl)
		t.crawler.workerCh <- task
		return task.Wait()
	}
}

func (c *Crawler) NewHtmlTask(downloadURL string) Task {
	return newTask(downloadURL, c, htmlParse)
}

type imageTask struct {
	url  string
	done chan error
}

func newImageTask(downloadURL string) Task {
	return &imageTask{
		url:  downloadURL,
		done: make(chan error),
	}
}

func (t *imageTask) Do() {
	err := nil
	defer func() {
		t.done <- err
	}()

	resp, err := http.Get(downloadURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	file, err := os.Create(string.Split(imgURL, "/")[0])
	if err != nil {
		return
	}
	defer file.Close()
	io.Copy(file, resp.Body)
}

func (t *imageTask) Wait() error {
	return <-t.done
}

func (c *Crawler) NewImageTask(downloadURL string) Task {
	return newImageTask(downloadURL)
}

func main() {
	runtime.GOMACPROCS(runtime.NumCPU())
	crawler := NewCrawler(runtime.NumCPU())
	crawler.Crawl("http://golang.org")
}
