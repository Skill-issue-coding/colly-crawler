package scraper

import (
	"log"
	"regexp"
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

	c := NewCollector()

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
		if match, _ := regexp.MatchString(`^([A-Za-z]{4}\d{1})$`, code); match && len(code) == 5 {
			programCode := code
		}
	})

	// Semester and course links
	c.OnHTML("div.tab-pane#curriculum", func(e *colly.HTMLElement) {
		print("found curriculum tab\n")
		var currentSemester *models.Semester

		e.ForEach("*", func(_ int, el *colly.HTMLElement) {
			if el.Name == "h3" {

				mutex.Lock()
				program.Semesters = append(program.Semesters, models.Semester{
					Name: strings.TrimSpace(el.Text),
				})
				currentSemester = &program.Semesters[len(program.Semesters)-1]
				mutex.Unlock()

			}
			if el.Name == "a" && isCourseLink(el.Attr("href")) && currentSemester != nil {
				semesterName := currentSemester.Name // Capture current semester name
				url := e.Request.AbsoluteURL(el.Attr("href"))
				log.Printf("Found course link: %s for semester %s\n", url, semesterName)
				wg.Add(1)

				courseCollector := c.Clone()
				go ScrapeCourse(url, semesterName, courseCollector, &wg, courseChan)
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
