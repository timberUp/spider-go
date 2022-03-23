package crawler

import (
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"mini-spider/parser"
)

type Fetcher interface {
	start()
	fetch(url Url) ([]Url, error)
	send(urls []Url)
}

type worker struct {
	id            int
	client        *http.Client
	wg            *sync.WaitGroup
	mu            *sync.RWMutex
	urlPattern    *regexp.Regexp
	urlMap        map[string]bool
	taskQueue     chan Url
	quit          chan struct{}
	crawlInterval time.Duration
	lastFetch     time.Time
	count         int32
	outputDir     string
}

func (w *worker) start() {
	logrus.Info("[fetcher-%03d] started")
	defer w.wg.Done()

	for {
		select {
		case url := <-w.taskQueue:
			logrus.Infof("[fetcher-%03d] received url [%s]", w.id, url.link)
			urls, err := w.fetch(url)
			if err != nil {
				logrus.Warnf("[fetcher-%03d] failed to fetch [%s], err: %v", w.id, url.link, err)
			} else {
				go w.send(urls)
			}
		case <-w.quit:
			logrus.Infof("[fetcher-%03d] downloaded %d urls, quit", w.id, w.count)
			return
		}
	}
}

func (w *worker) fetch(url Url) ([]Url, error) {
	if !w.urlPattern.MatchString(url.link) && url.depth == 0 {
		return nil, nil
	}

	for time.Since(w.lastFetch) < w.crawlInterval {
		// wait
	}
	resp, err := w.client.Get(url.link)
	if err != nil {
		return nil, fmt.Errorf("get url [%s] response failed, err: %v", url.link, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got error when getting content from [%s], httpStatus: %v", url.link, resp.StatusCode)
	}

	if w.urlPattern.MatchString(url.link) {
		if err = parser.Download(url.link, w.outputDir, resp.Body); err != nil {
			logrus.Errorf("failed to download [%s], err: %v", url.link, err)
		} else {
			logrus.Infof("[%s] downloaded", url.link)
			atomic.AddInt32(&w.count, 1)
		}
	}

	var links []Url
	if url.depth != 0 {
		rawUrls, err := parser.Extract(resp)
		if err != nil {
			return nil, fmt.Errorf("parse %s as HTML failed, errL %v", url.link, err)
		}
		for _, l := range rawUrls {
			links = append(links, Url{
				link:  l,
				depth: url.depth - 1,
			})
		}
		logrus.Infof("%d links extracted from url [%s]", len(links), url.link)
	}
	w.lastFetch = time.Now()

	return links, nil
}

func (w *worker) send(urls []Url) {
	w.wg.Add(1)
	defer w.wg.Done()
	for _, u := range urls {
		select {
		case <-w.quit:
			logrus.Warnf("[fetcher-%03d] quit unexpectedly", w.id)
			return
		default:
			w.mu.Lock()
			if _, ok := w.urlMap[u.link]; !ok {
				w.urlMap[u.link] = true
				w.taskQueue <- u
			}
			w.mu.Unlock()
		}
	}
	return
}
