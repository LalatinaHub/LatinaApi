package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/LalatinaHub/LatinaApi/api/router"
	"github.com/LalatinaHub/LatinaApi/common/account"
	"github.com/LalatinaHub/LatinaApi/common/account/converter"
	"github.com/LalatinaHub/LatinaApi/common/helper"
	latinasub "github.com/LalatinaHub/LatinaSub-go"
	"github.com/go-co-op/gocron"
)

var (
	filterCron = os.Getenv("FILTER_CRON")
	loc, _     = time.LoadLocation("Asia/Jakarta")
)

func cronJob() {
	schedule := gocron.NewScheduler(loc)
	schedule.SetMaxConcurrentJobs(1, gocron.RescheduleMode)

	if filterCron == "" {
		filterCron = "00 */6 * * *"
	}

	schedule.Cron(filterCron).Tag("filter").Do(func() {
		nodes := strings.Split(converter.ToRaw(account.Get("")), "\n")
		if len(nodes) > 100 {
			fmt.Println("Filtering accounts ...")
			helper.LogFuncToFile(func() {
				latinasub.Start(nodes, true)
			}, "scrape.log")
		} else {
			fmt.Println("No accounts found!")
			schedule.RunByTag("scrape")
		}
	})

	schedule.Every(1).Day().At("09:00").Tag("scrape").Do(func() {
		fmt.Println("Scraping accounts ...")
		helper.LogFuncToFile(func() {
			latinasub.Start([]string{}, true)
		}, "scrape.log")
	})

	schedule.StartAsync()
	// schedule.RunByTag("scrape")
}

func main() {

	// Set cron job
	cronJob()

	// Start server
	router.Start()
}
