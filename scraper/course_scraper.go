package scraper

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/Skill-issue-coding/colly-crawler/models"
	"github.com/gocolly/colly/v2"
)

func ScrapeCourse(url string, semesterName string, c *colly.Collector, wg *sync.WaitGroup, courseChan chan<- scrapedCourse, programCode string) {
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

	c.OnHTML("section.overview-content", func(e *colly.HTMLElement) {
		var currentLabel string

		e.DOM.Contents().Each(func(i int, s *goquery.Selection) {
			nodeName := goquery.NodeName(s)
			text := strings.TrimSpace(s.Text())

			if nodeName == "h2" {
				currentLabel = text
				return
			}

			if currentLabel != "" && text != "" && nodeName == "#text" {
				switch currentLabel {
				case "Huvudområde":
					course.Overview.Subject = text
				case "Utbildningsnivå":
					course.Overview.Level = text
				case "Kurstyp":
					course.Overview.Type = text
				case "Examinator":
					course.Overview.Examiner = text
				case "Studierektor eller motsvarande":
					course.Overview.Director = text
				case "Undervisningstid":
					fullText := s.Parent().Text()

					// Split by "Rekommenderad"
					parts := strings.Split(fullText, "Rekommenderad")
					if len(parts) >= 2 {
						scheduled := strings.TrimSpace(parts[0])
						recommended := "Rekommenderad" + strings.TrimSpace(parts[1])

						course.Overview.ScheduledHours = extractHours(scheduled)
						course.Overview.SelfStudyHours = extractHours(recommended)
					}
				}
				currentLabel = ""
			}
		})
	})

	c.OnHTML("table.study-guide-table", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			codeCell := row.ChildText("td:nth-of-type(1)")
			vofCell := row.ChildText("td:nth-last-of-type(1)")

			if strings.EqualFold(codeCell, programCode) {
				course.Overview.Period = row.ChildText("td:nth-of-type(4)")
				course.Overview.Block = row.ChildText("td:nth-of-type(5)")
				course.Overview.Language = row.ChildText("td:nth-of-type(6)")
				course.Overview.Campus = row.ChildText("td:nth-of-type(7)")
				course.Overview.VOF = strings.TrimSpace(vofCell)
			}
		})
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

func extractHours(s string) string {
	re := regexp.MustCompile(`(\d+)\s*h\b`)
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return matches[1] + " h"
	}
	return s
}
