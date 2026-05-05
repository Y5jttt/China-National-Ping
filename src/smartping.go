package main

import (
	"flag"
	"fmt"
	"github.com/jakecoffman/cron"
	"github.com/smartping/smartping/src/funcs"
	"github.com/smartping/smartping/src/g"
	"github.com/smartping/smartping/src/http"
	"os"
	"runtime"
	"time"
)

var Version = "0.8.0"

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// Set timezone to Beijing
	time.LoadLocation("Asia/Shanghai")
	version := flag.Bool("v", false, "show version")
	flag.Parse()
	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}
	g.ParseConfig(Version)
	go funcs.ClearArchive()
	c := cron.New()
	c.AddFunc("*/60 * * * * *", func() {
		go funcs.Ping()
		go funcs.Mapping()
		if g.Cfg.Mode["Type"] == "cloud" {
			go funcs.StartCloudMonitor()
		}
	}, "ping")
	c.Start()
	http.StartHttp()
}
