package scraper

import (
	"time"

	"github.com/gocolly/colly/v2"
)

func NewCollector() *colly.Collector {

	c := colly.NewCollector(
		colly.AllowedDomains("studieinfo.liu.se"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,               // Parallel scraping channels (1 for now to avoid ip bans),
		Delay:       1 * time.Second, // Delay between requests, change later
		RandomDelay: 500 * time.Microsecond,
	})

	return c
}
