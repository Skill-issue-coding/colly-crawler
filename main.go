package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly/v2"
)

func main() {

	// Create a new collector
	c := colly.NewCollector(
		colly.AllowedDomains("studieinfo.liu.se"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36"),
	)

	// Print visited URLs
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL.String())
	})

	var currentProgram, currentCredits string

	// Capture h1 separately (async)
	c.OnHTML("h1", func(e *colly.HTMLElement) {
		parts := strings.Split(strings.TrimSpace(e.Text), ",")
		if len(parts) >= 2 {
			currentProgram = strings.TrimSpace(parts[0])
			currentCredits = strings.TrimSpace(parts[1])
		}
	})

	// Handle "Programplan" page
	c.OnHTML("div.tab-pane#curriculum", func(e *colly.HTMLElement) {
		fmt.Println("\n Curriculum section:")

		// Print URL
		currentURL := e.Request.URL.String()
		fmt.Println("\nCurrent Page URL:", currentURL)

		// Print program and credits
		if currentProgram != "" {
			fmt.Printf("Program: %s\nCredits: %s\n", currentProgram, currentCredits)
		}

		// Loop and print all semsesters
		e.ForEach("h3", func(_ int, h3 *colly.HTMLElement) {
			fmt.Printf("H3: %s\n", h3.Text)
		})
	})

	c.Visit("https://studieinfo.liu.se/program/6CMEN/5712#overview")
}
