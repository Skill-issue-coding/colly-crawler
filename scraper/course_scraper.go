package scraper

import (
	"sync"

	"github.com/Skill-issue-coding/colly-crawler/models"
	"github.com/gocolly/colly/v2"
)

func ScrapeCourse(url string, semesterName string, c *colly.Collector, wg *sync.WaitGroup, courseChan chan<- models.Course) {

}
