package main

import (
	"fmt"

	"github.com/gocolly/colly/v2"
)

func main() {

	// Create a new collector
	c := colly.NewCollector(
		colly.AllowedDomains("studieinfo.liu.se"),
	)

	// Testing the collector with h2 on first page
	c.OnHTML("h2.overview-label", func(e *colly.HTMLElement) {
		fmt.Println("h2-tags: ", e.Text)
	})

	// Print visited URLs
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL.String())
	})

	// Test to print content on "Programplan"
	/* c.OnHTML("div.tab-pane#curriculum", func(e *colly.HTMLElement) {
		fmt.Println("\nCurriculum Content Found!")
		fmt.Println("----------------------")
		fmt.Println(e.Text)
	}) */

	// Test to print button name "Programplan"
	c.OnHTML("button.tabs-nav__link[data-bs-target='#curriculum']", func(e *colly.HTMLElement) {
		fmt.Println("\nFound Curriculum Button:")
		fmt.Println("Button Text:", e.Text)
		fmt.Println("Data Target:", e.Attr("data-bs-target"))
	})

	c.Visit("https://studieinfo.liu.se/program/6CMEN/5712#overview")
}
