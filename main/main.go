package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/gcfg.v1"

	"mini-spider/config"
	"mini-spider/crawler"
)

const SpiderVersion = "1.0.0"

var (
	fset *flag.FlagSet
	h    bool
	v    bool
	c    string
	l    string
)

func init() {
	fset = flag.NewFlagSet("cmd-tool", flag.ExitOnError)

	fset.BoolVar(&h, "h", false, "this help")
	fset.BoolVar(&v, "v", false, "mini-spider version")
	fset.StringVar(&c, "c", "../config/spider.conf", "config file for mini-spider")
	fset.StringVar(&l, "l", "../log", "log for mini-spider")

	fset.Usage = usage
}

func main() {
	_ = fset.Parse(os.Args[1:])
	if h {
		fset.Usage()
		os.Exit(0)
	}
	if v {
		fmt.Println(SpiderVersion)
		os.Exit(0)
	}

	// new logger
	err := initLogger(l, "mini_spider.log", logrus.InfoLevel, true)
	if err != nil {
		logrus.Errorf("failed to create logFile, err: %v", err)
		os.Exit(-1)
	}

	// read config
	var cfg config.Config
	if err := gcfg.ReadFileInto(&cfg, c); err != nil {
		logrus.Errorf("parse config error, err: %v", err)
		os.Exit(-1)
	}
	logrus.Info("read config success")

	c, err := crawler.NewCrawler(cfg)
	if err != nil {
		logrus.Errorf("failed to new Crawler, err: %v", err)
		os.Exit(-1)
	}

	if err = c.Start(); err != nil {
		logrus.Errorf("failed to start crawler, err: %v", err)
		os.Exit(-1)
	}

	c.Loop()
}

func initLogger(path, fileName string, level logrus.Level, hasStdOut bool) error {
	f, err := os.OpenFile(filepath.Join(path, fileName), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetOutput(f)
	logrus.SetLevel(level)
	if hasStdOut {
		mw := io.MultiWriter(os.Stdout, f)
		logrus.SetOutput(mw)
	}
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr,
		`Usage: ./mini-spider
	[-h help]	this help
	[-c conf]	config path
	[-l log]	log path
	[-v]		version`)
}
