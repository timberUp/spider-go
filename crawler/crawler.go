package crawler

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"mini-spider/config"
)

type Url struct {
	link  string
	depth int
}

type Crawler struct {
	Client        *http.Client
	Urls          []string
	MaxDepth      int
	Interval      int
	OutputDir     string
	TargetPattern *regexp.Regexp
	ThreadCount   int
	taskQueue     chan Url
	quit          chan struct{}
	urlMap        map[string]bool
	wg            sync.WaitGroup
	mu            sync.RWMutex
}

// NewCrawler new a Crawler instance
func NewCrawler(cfg config.Config) (*Crawler, error) {
	if err := config.CheckConfig(&cfg); err != nil {
		return nil, err
	}

	file, err := os.Open(cfg.UrlListFile)
	if err != nil {
		logrus.Errorf("failed to read urlListFile, err: %v", err)
		return nil, err
	}
	defer file.Close()
	var urls []string
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&urls)
	if err != nil {
		logrus.Errorf("failed to load urlListFile, err: %v", err)
		return nil, err
	}

	reg, err := regexp.Compile(cfg.TargetUrl)
	if err != nil {
		logrus.Errorf("failed to compile targetUrl pattern, err: %v", err)
		return nil, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: time.Duration(cfg.CrawlTimeout) * time.Second,
	}

	return &Crawler{
		Client:        client,
		Urls:          urls,
		MaxDepth:      cfg.MaxDepth,
		Interval:      cfg.CrawlInterval,
		OutputDir:     cfg.OutputDirectory,
		TargetPattern: reg,
		ThreadCount:   cfg.ThreadCount,
		taskQueue:     make(chan Url, cfg.ThreadCount),
		quit:          make(chan struct{}),
		urlMap:        map[string]bool{},
		wg:            sync.WaitGroup{},
		mu:            sync.RWMutex{},
	}, nil
}

// initTaskQueue enqueues seed urls into taskQueue
// assuming all seed urls are valid and not duplicated
func (c *Crawler) initTaskQueue() {
	c.wg.Add(1)
	defer c.wg.Done()
	for _, u := range c.Urls {
		select {
		case <-c.quit:
			logrus.Warnf("crawler quit unexpectedly")
			return
		default:
			c.mu.Lock()
			if _, ok := c.urlMap[u]; !ok {
				c.urlMap[u] = true
				c.taskQueue <- Url{
					link:  u,
					depth: c.MaxDepth,
				}
			}
			c.mu.Unlock()
			logrus.Infof("[%s] sent to taskQueue", u)
		}
	}
	return
}

// Start launch a Crawler
func (c *Crawler) Start() error {
	logrus.Infof("crawler starting...")

	go c.initTaskQueue()

	c.wg.Add(c.ThreadCount)
	for i := 0; i < c.ThreadCount; i++ {
		w := worker{
			id:            i,
			client:        c.Client,
			wg:            &c.wg,
			mu:            &c.mu,
			urlPattern:    c.TargetPattern,
			urlMap:        c.urlMap,
			taskQueue:     c.taskQueue,
			quit:          c.quit,
			crawlInterval: time.Duration(c.Interval) * time.Second,
			lastFetch:     time.Now(),
			count:         0,
			outputDir:     c.OutputDir,
		}
		go w.start()
	}

	return nil
}

func (c *Crawler) Stop() {
	logrus.Info("stopping crawler...")
	close(c.quit)
	c.wg.Wait()
	close(c.taskQueue) // taskQueue should close after all workers quit, or panic
	logrus.Info("all workers quit, crawler gracefully shut down")
}

// Loop listens for termination signal to shut down Crawler
func (c *Crawler) Loop() {
	logrus.Info("loop for termination signal")

	buf := make([]byte, 1<<20)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	for {
		select {
		case s := <- sig:
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				logrus.Info("==== received SIGINT/SIGTERM, exiting ===\n***")
			case syscall.SIGQUIT:
				stackLen := runtime.Stack(buf, true)
				logrus.Infof("==== received SIGQUIT ===\n*** goroutine dump...\n%s\n***end", buf[:stackLen])
			}
			c.Stop()
			return
		default:
			// TODO: quit when all tasks consumed?
		}
	}
}
