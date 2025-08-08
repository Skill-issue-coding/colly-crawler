package scraper

import (
	"log"
	"strings"
	"sync"

	"github.com/Skill-issue-coding/colly-crawler/models"
	"github.com/gocolly/colly/v2"
)

type scrapedCourse struct {
	Course       models.Course
	SemesterName string
}

func ScrapeProgram(url string) (models.Program, error) {

	program := models.Program{Url: url}
	var mutex sync.Mutex
	var wg sync.WaitGroup
	courseChan := make(chan scrapedCourse)
	var programCode string

	c := NewCollector()

	// scrape first semester and first course for testing now
	foundFirstSemester := false
	foundFirstCourse := false

	// Program name
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

	// Program code used in course_scraper for other checks
	c.OnHTML("p.subtitle", func(e *colly.HTMLElement) {
		code := strings.TrimSpace(e.Text)
		log.Printf("Raw subtitle bytes: %q\n", code)

		if programCode == "" && isProgramCode(code) {
			programCode = code
			log.Printf("Found program code: %s", programCode)
		}
	})

	// Semester and course links
	c.OnHTML("div.tab-pane#curriculum", func(e *colly.HTMLElement) {
		print("found curriculum tab\n")
		var currentSemester *models.Semester

		e.ForEach("*", func(_ int, el *colly.HTMLElement) {

			// scrape first semester and first course for testing now
			if foundFirstCourse {
				return
			}

			if el.Name == "h3" && !foundFirstSemester { // scrape first semester and first course for testing now

				mutex.Lock()
				program.Semesters = append(program.Semesters, models.Semester{
					Name: strings.TrimSpace(el.Text),
				})
				currentSemester = &program.Semesters[len(program.Semesters)-1]
				mutex.Unlock()

				// scrape first semester and first course for testing now
				foundFirstSemester = true

			}
			if el.Name == "a" && isCourseLink(el.Attr("href")) && currentSemester != nil && !foundFirstCourse { // scrape first semester and first course for testing now
				semesterName := currentSemester.Name // Capture current semester name
				url := e.Request.AbsoluteURL(el.Attr("href"))
				log.Printf("Found course link: %s for semester %s\n", url, semesterName)
				wg.Add(1)

				courseCollector := c.Clone()
				go ScrapeCourse(url, semesterName, courseCollector, &wg, courseChan, programCode)

				// scrape first semester and first course for testing now
				foundFirstCourse = true
			}
		})
	})

	err := c.Visit(url)
	if err != nil {
		return program, err
	}

	c.Wait()

	go func() {
		wg.Wait()
		close(courseChan)
	}()

	for sc := range courseChan {
		mutex.Lock()
		for i := range program.Semesters {
			if program.Semesters[i].Name == sc.SemesterName {
				program.Semesters[i].Courses = append(program.Semesters[i].Courses, sc.Course)
				break
			}
		}
		mutex.Unlock()
	}

	return program, nil
}

func isCourseLink(href string) bool {
	return strings.Contains(href, "/kurs/") || strings.Contains(href, "/course/")
}

func isProgramCode(code string) bool {
	if len(code) != 5 {
		return false
	}

	digits := 0
	letters := 0

	for _, r := range code {
		if r >= '0' && r <= '9' {
			digits++
		} else if r >= 'A' && r <= 'Z' {
			letters++
		} else {
			return false
		}
	}

	return digits == 1 && letters == 4
}
