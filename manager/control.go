package manager

import (
	"mini-spider/crawler"
	"sync"
)

// Controller is the brain for all Fetchers
type Controller struct {
	fetchers []crawler.Crawler
	quit     <-chan struct{}
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

func NewController() {}

type manager interface {}
