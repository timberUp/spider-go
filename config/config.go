package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

const (
	DefaultOutputDir     = "../output"
	DefaultMaxDepth      = 0
	DefaultCrawlInterval = 1
	DefaultCrawlTimeout  = 1
	DefaultThreadCount   = 8
)

type Config struct {
	Spider `gcfg:"spider"`
}

type Spider struct {
	UrlListFile     string `gcfg:"urlListFile"`
	OutputDirectory string `gcfg:"outputDirectory"`
	MaxDepth        int    `gcfg:"maxDepth"`
	CrawlInterval   int    `gcfg:"crawlInterval"`
	CrawlTimeout    int    `gcfg:"crawlTimeout"`
	TargetUrl       string `gcfg:"targetUrl"`
	ThreadCount     int    `gcfg:"threadCount"`
}

// CheckConfig validates config and set default value
func CheckConfig(cfg *Config) error {
	if len(strings.TrimSpace(cfg.UrlListFile)) == 0 {
		return fmt.Errorf("empty path for seed url")
	}
	if len(strings.TrimSpace(cfg.OutputDirectory)) == 0 {
		logrus.Warn("found empty output dir, set to default folder ../output")
		cfg.OutputDirectory = DefaultOutputDir
	}
	if err := os.MkdirAll(cfg.OutputDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory, err: %v", err)
	}
	if cfg.MaxDepth < 0 {
		logrus.Warn("maxDepth should >= 0, set to default value 0")
		cfg.MaxDepth = DefaultMaxDepth
	}
	if cfg.CrawlInterval < 0 {
		logrus.Warn("crawlInterval should >= 0, set to default value 1")
		cfg.CrawlInterval = DefaultCrawlInterval
	}
	if cfg.CrawlTimeout < 0 {
		logrus.Warn("crawlTimeout should > 0, set to default value 1")
		cfg.CrawlTimeout = DefaultCrawlTimeout
	}
	if len(strings.TrimSpace(cfg.TargetUrl)) == 0 {
		return fmt.Errorf("please specify targetUrl parttern")
	}
	if cfg.ThreadCount <= 0 {
		logrus.Warn("threadCount should > 0, set to default value 8")
		cfg.ThreadCount = DefaultThreadCount
	}
	return nil
}
