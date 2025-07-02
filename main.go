package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
)

type Semester struct {
	Name string `json:"name"`
}

type Program struct {
	Name      string     `json:"name"`
	Credits   string     `json:"credits"`
	Url       string     `json:"url"`
	Semesters []Semester `json:"semsesters"`
}

func main() {
	program := Program{}

	// Create a new collector
	c := colly.NewCollector(
		colly.AllowedDomains("studieinfo.liu.se"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36"),
	)

	// Print visited URLs
	c.OnRequest(func(r *colly.Request) {
		program.Url = r.URL.String()
		fmt.Println("Visiting:", r.URL.String())
	})

	// Capture h1 separately (async)
	c.OnHTML("h1", func(e *colly.HTMLElement) {
		parts := strings.SplitN(strings.TrimSpace(e.Text), ",", 2)
		if len(parts) >= 2 {
			program.Name = strings.TrimSpace(parts[0])
			if len(parts) >= 2 {
				program.Credits = strings.TrimSpace(parts[1])
			}
		}
	})

	// Handle "Programplan" page
	c.OnHTML("div.tab-pane#curriculum", func(e *colly.HTMLElement) {
		e.ForEach("h3", func(_ int, h3 *colly.HTMLElement) {
			program.Semesters = append(program.Semesters, Semester{
				Name: strings.TrimSpace(h3.Text),
			})
		})
	})

	// Error handling
	c.OnError(func(r *colly.Response, err error) {
		log.Println("Error:", err)
	})

	// Start the scraping process
	err := c.Visit("https://studieinfo.liu.se/program/6CMEN/5712#overview")
	if err != nil {
		log.Fatal(err)
	}

	// Convert the program data to JSON
	jsonData, err := json.MarshalIndent(program, "", "  ")
	if err != nil {
		log.Fatal("JSON marshaling error:", err)
	}

	// Save to file
	err = os.WriteFile("program_data.json", jsonData, 0644)
	if err != nil {
		log.Fatal("File write error:", err)
	}

	fmt.Println("\nScraping complete! Data saved to program_data.json")
	fmt.Println(string(jsonData)) // Print to console for verification
}
