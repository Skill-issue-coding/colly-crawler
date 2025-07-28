package scraper

import (
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
			program.Name = parts[0]
			if len(parts) > 1 {
				program.Credits = parts[1]
			}
		}
		mutex.Unlock()
	})

	// Semester and course links
	c.OnHTML("div.tab-pane#curriculum", func(e *colly.HTMLElement) {
		print("found curriculum tab\n")
		var currentSemester *models.Semester

		e.ForEach("*", func(_ int, el *colly.HTMLElement) {
			if el.Name == "h3" {

				semester := models.Semester{Name: strings.TrimSpace(el.Text)}
				currentSemester = &semester

				mutex.Lock()
				program.Semesters = append(program.Semesters, semester)
				mutex.Unlock()

			}
			if el.Name == "a" && isCourseLink(el.Attr("href")) && currentSemester != nil {
				semesterName := currentSemester.Name // Capture current semester name
				url := e.Request.AbsoluteURL(el.Attr("href"))
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

	go func() {
		wg.Wait()
		close(courseChan)
	}()

	c.Wait()

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
