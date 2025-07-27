package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Skill-issue-coding/colly-crawler/models"
	"github.com/gocolly/colly/v2"
)

func main() {
	program := models.Program{}
	var mutex = sync.Mutex{}
	var wg sync.WaitGroup

	outputDir := "data"
	outputFile := "program_data.json"

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
		print("found program name\n")
		parts := strings.SplitN(strings.TrimSpace(e.Text), ",", 2)
		mutex.Lock()
		if len(parts) > 0 {
			program.Name = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				program.Credits = strings.TrimSpace(parts[1])
			}
		}
		mutex.Unlock()
	})

	c.OnHTML("div.tab-pane#curriculum", func(e *colly.HTMLElement) {
		print("found curriculum tab\n")
		var currentSemester *models.Semester

		e.ForEach("*", func(_ int, el *colly.HTMLElement) {
			if el.Name == "h3" {

				semester := models.Semester{
					Name:    strings.TrimSpace(el.Text),
					Courses: []models.Course{},
				}

				currentSemester = &semester
				mutex.Lock()
				program.Semesters = append(program.Semesters, semester)
				mutex.Unlock()

			}
			if el.Name == "a" && isCourseLink(el.Attr("href")) && currentSemester != nil {
				semesterName := currentSemester.Name // Capture current semester name
				wg.Add(1)

				url := e.Request.AbsoluteURL(el.Attr("href"))

				courseCollector := c.Clone()

				course := &models.Course{Url: url} // shared pointer between handlers

				courseCollector.OnHTML("h1", func(h *colly.HTMLElement) {
					parts := strings.SplitN(strings.TrimSpace(h.Text), ",", 2)
					if len(parts) > 0 {
						course.Name = strings.TrimSpace(parts[0])
						if len(parts) > 1 {
							course.Credits = strings.TrimSpace(parts[1])
						}
					}
				})

				courseCollector.OnHTML("p.subtitle", func(h *colly.HTMLElement) {
					code := strings.TrimSpace(h.Text)
					if match, _ := regexp.MatchString(`^([A-Za-z]{3}\d{3}|[A-Za-z]{4}\d{2})$`, code); match && len(code) == 6 {
						course.Code = code
					}

				})

				// Once scraping of both elements is done, attach the course
				courseCollector.OnScraped(func(_ *colly.Response) {
					mutex.Lock()
					defer mutex.Unlock()
					for i := range program.Semesters {
						if program.Semesters[i].Name == semesterName {
							program.Semesters[i].Courses = append(program.Semesters[i].Courses, *course)
							break
						}
					}
					wg.Done()
				})

				if err := courseCollector.Visit(url); err != nil {
					// if !strings.Contains(err.Error(), "already visited") {} // Ignore already visited errors/logs
					log.Printf("Failed to visit course URL %s: %v\n", url, err)
					wg.Done()
				}

			}
		})

	})

	// Start the scraping process
	program.Url = "https://studieinfo.liu.se/program/6CMEN/5712#overview"
	err := c.Visit(program.Url)
	if err != nil {
		log.Fatal(err)
	}

	c.Wait()
	wg.Wait() // Wait for all goroutines to finish

	// Convert the program data to JSON
	jsonData, err := json.MarshalIndent(program, "", "  ")
	if err != nil {
		log.Fatal("JSON marshaling error:", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	filePath := filepath.Join(outputDir, outputFile)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		log.Fatal("Failed to write JSON file:", err)
	}

	fmt.Println("\nScraping complete! Data saved to data/program_data.json")
	fmt.Printf("Found %d semesters:", len(program.Semesters))
	for i, s := range program.Semesters {
		fmt.Printf("Semester %d %s (%d courses):\n", i+1, s.Name, len(s.Courses))
	}
}

func isCourseLink(href string) bool {
	return strings.Contains(href, "/kurs/") || strings.Contains(href, "/course/")
}
