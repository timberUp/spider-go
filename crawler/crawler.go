package crawler

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"sync"
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

func (c *Crawler) Start() error  {
	return nil
}
