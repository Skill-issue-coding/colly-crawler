package scraper

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/Skill-issue-coding/colly-crawler/models"
	"github.com/gocolly/colly/v2"
)

func ScrapeCourse(url string, semesterName string, c *colly.Collector, wg *sync.WaitGroup, courseChan chan<- scrapedCourse) {
	defer wg.Done()
	course := &models.Course{Url: url} // shared pointer between handlers

	c.OnHTML("h1", func(h *colly.HTMLElement) {
		fmt.Println("Scraping course name:", h.Text)
		parts := strings.SplitN(strings.TrimSpace(h.Text), ",", 2)
		if len(parts) > 0 {
			course.Name = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				course.Credits = strings.TrimSpace(parts[1])
			}
		}
	})

	c.OnHTML("p.subtitle", func(h *colly.HTMLElement) {
		fmt.Println("Scraping course code:", h.Text)
		code := strings.TrimSpace(h.Text)
		if match, _ := regexp.MatchString(`^([A-Za-z]{3}\d{3}|[A-Za-z]{4}\d{2})$`, code); match && len(code) == 6 {
			course.Code = code
		}

	})

	// Once scraping of both elements is done, attach the course
	c.OnScraped(func(_ *colly.Response) {
		fmt.Printf("Finished scraping %s\n", url)
		courseChan <- scrapedCourse{
			Course:       *course,
			SemesterName: semesterName,
		}
	})

	if err := c.Visit(url); err != nil {
		// if !strings.Contains(err.Error(), "already visited") {} // Ignore already visited errors/logs
		log.Printf("Failed to visit course URL %s: %v\n", url, err)
	}
	c.Wait()
}
