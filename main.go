package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Skill-issue-coding/colly-crawler/scraper"
)

func main() {
	// Start the scraping process
	programUrl := "https://studieinfo.liu.se/program/6CMEN/5712#overview"
	program, err := scraper.ScrapeProgram(programUrl)
	if err != nil {
		log.Fatal(err)
	}

	outputDir := "data"
	outputFile := "test.json" // TODO: Change name later after testing

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal("Failed to create output directory:", err)
	}

	// Convert the program data to JSON
	jsonData, _ := json.MarshalIndent(program, "", "  ")

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
