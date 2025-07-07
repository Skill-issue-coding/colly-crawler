package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

type Course struct {
	Name    string `json:"name"`
	Credits string `json:"credits"`
	Url     string `json:"url"`
}

type Semester struct {
	Name    string   `json:"name"`
	Courses []Course `json:"courses"`
}

type Program struct {
	Name      string     `json:"name"`
	Credits   string     `json:"credits"`
	Url       string     `json:"url"`
	Semesters []Semester `json:"semsesters"`
}

func main() {
	program := Program{}

	var mutex = sync.Mutex{} // Mutex to protect shared data

	var wg sync.WaitGroup // WaitGroup to wait for all goroutines to finish

	// Create a new collector
	c := colly.NewCollector(
		colly.AllowedDomains("studieinfo.liu.se"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.54 Safari/537.36"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		RandomDelay: 500 * time.Microsecond,
	})

	c.OnHTML("h1", func(e *colly.HTMLElement) {
		parts := strings.SplitN(strings.TrimSpace(e.Text), ",", 2)
		mutex.Lock()
		defer mutex.Unlock()
		if len(parts) > 0 {
			program.Name = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				program.Credits = strings.TrimSpace(parts[1])
			}
		}
	})

	c.OnHTML("div.tab-pane#curriculum", func(e *colly.HTMLElement) {
		currentSemester := Semester{}
		var currentCourses []Course

		e.ForEach("*", func(_ int, el *colly.HTMLElement) {
			if el.Name == "h3" {
				if currentSemester.Name != "" {
					mutex.Lock()
					program.Semesters = append(program.Semesters, currentSemester)
					mutex.Unlock()
				}
				currentSemester = Semester{
					Name:    strings.TrimSpace(el.Text),
					Courses: []Course{},
				}
				currentCourses = []Course{}
			}
			if el.Name == "a" {
				wg.Add(1)

				url := e.Request.AbsoluteURL(el.Attr("href"))

				courseCollector := c.Clone()
				courseCollector.OnHTML("h1", func(h *colly.HTMLElement) {
					defer wg.Done()
					parts := strings.SplitN(strings.TrimSpace(h.Text), ",", 2)
					course := Course{Url: url}
					if len(parts) > 0 {
						course.Name = strings.TrimSpace(parts[0])
						if len(parts) > 1 {
							course.Credits = strings.TrimSpace(parts[1])
						}
					}

					mutex.Lock()
					currentCourses = append(currentCourses, course)
					mutex.Unlock()
				})

				go func() {
					if err := courseCollector.Visit(url); err != nil {
						log.Printf("Failed to visit course URL %s: %v\n", url, err)
						wg.Done()
					}
				}()
			}
		})

	})

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

		e.ForEach("a[href]", func(_ int, a *colly.HTMLElement) {
			href := a.Attr("href")
			if strings.HasPrefix(href, "/kurs/") {
				c.Visit(e.Request.AbsoluteURL(href)) // Visit course links
			}
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
