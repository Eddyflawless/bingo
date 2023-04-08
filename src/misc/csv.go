package main

import (
	"encoding/csv"
	"log"
	"os"
)

type CsvRecord struct {
	filename  string
	delimiter rune
}

func (c *CsvRecord) Read() ([][]string, error) {

	file, err := c.OpenFile("read")
	if err != nil {
		log.Fatalf("Failed to opening CSV file: %v", err)
	}

	reader := csv.NewReader(file)
	reader.Comma = c.delimiter

	records, err := reader.ReadAll()

	return records, err

}

func (c *CsvRecord) OpenFile(mode string) (*os.File, error) {

	if mode == "read" {
		return os.Open(c.filename)
	}
	return os.Create(c.filename)

}

func (c *CsvRecord) Write(records [][]string) {

	csvFile, err := c.OpenFile("write")
	// handle error
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}

	csvWriter := csv.NewWriter(csvFile)

	for _, record := range records {
		_ = csvWriter.Write(record)
	}
	csvWriter.Flush()
}

func NewCsvHandler(filename string, delimiter rune) *CsvRecord {

	return &CsvRecord{filename: filename, delimiter: delimiter}
}
