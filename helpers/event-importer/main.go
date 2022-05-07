package main

import (
	"encoding/csv"
	"fmt"
	"github.com/gocarina/gocsv"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"
)

type Input struct {
	Datum        string  `csv:"Datum"`
	Uhrzeit      *string `csv:"Uhrzeit,omitempty"`
	Ort          *string `csv:"Ort,omitempty"`
	Name         string  `csv:"Name"`
	Veranstalter string  `csv:"Veranstalter"`
}

type Output struct {
	Date   string
	Title  string
	Verein string
	Ort    string
}

func main() {
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(in)
		r.Comma = ';'
		return r
	})

	inputFile, err := os.OpenFile("files/2022.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer inputFile.Close()

	events := []*Input{}

	if err := gocsv.UnmarshalFile(inputFile, &events); err != nil {
		panic(err)
	}
	for _, input := range events {
		inputLayout := "02.01.2006 15:04"
		defaultTime := "00:00"
		timeOfEvent := input.Uhrzeit
		if timeOfEvent == nil {
			timeOfEvent = &defaultTime
		}

		formattedDate := fmt.Sprintf("%s %s", strings.TrimSpace(input.Datum), strings.TrimSpace(*timeOfEvent))

		dateOfEvent, err := time.Parse(inputLayout, formattedDate)

		if err != nil {
			fmt.Println(err)
		}

		defaultLocation := ""
		location := defaultLocation
		if input.Ort != nil {
			location = fmt.Sprintf("%s, Niederjosbach", strings.TrimSpace(*input.Ort))
		}

		output := Output{
			Date:   dateOfEvent.Format(time.RFC3339),
			Title:  input.Name,
			Verein: input.Veranstalter,
			Ort:    location,
		}
		fileName := getFileName(*input)
		fileLocation := fmt.Sprintf("../../content/termine/%d/%s.md", dateOfEvent.Year(), fileName)

		directoryToCreate, _ := filepath.Split(fileLocation)

		err = os.MkdirAll(directoryToCreate, os.ModePerm)
		if err != nil {
			panic(err)
		}
		file, err := os.Create(fileLocation)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		t := template.Must(template.New("event.tmpl").ParseFiles("event.tmpl"))
		err = t.Execute(file, output)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("generated %d events\n", len(events))
}
func getFileName(input Input) string {
	normalizedString := strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(
						strings.ReplaceAll(
							strings.ReplaceAll(
								strings.ReplaceAll(
									strings.ReplaceAll(
										strings.ReplaceAll(
											strings.ReplaceAll(
												strings.ReplaceAll(
													strings.ReplaceAll(strings.ToLower(input.Name), " ", "_"),
													"ü", "ue"),
												"ö", "oe"),
											"ä", "ae"),
										"ß", "ss"),
									".", ""),
								"\"", ""),
							"-", "_"),
						"/", ""),
					"(", ""),
				")", ""),
			"+", "und"),
		"__", "_")
	transformer := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	output, _, err := transform.String(transformer, normalizedString)
	if err != nil {
		panic(err)
	}
	return output
}
