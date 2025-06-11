package main

import (
	"fmt"

	"github.com/gocolly/colly/v2"
)

func main() {
	fmt.Println("Hello, Scraper!")

	c := colly.NewCollector(
		colly.AllowedDomains("studieinfo.liu.se"),
	)

	c.OnHTML("h2.overview-label", func(e *colly.HTMLElement) {
		fmt.Println("h2-tags: ", e.Text)
	})

	c.Visit("https://studieinfo.liu.se/program/6CMEN/5712#overview")
}
