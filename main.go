package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"regexp"
	"strings"
)

type ParaType int

const (
	PRIMARY_HEADER ParaType = iota
	SECONDARY_HEADER
	TEXT
	CODE
)

var (
	inputFile  = flag.String("in", "", "Input file")
	outputFile = flag.String("out", "", "Output file")
)

type Paragraph struct {
	Type     ParaType `json:"Type"`
	Content  string   `json:"Content"`
	Metadata string   `json:"Metadata"`
}

func main() {
	flag.Parse()

	fileBytes, err := os.ReadFile(*inputFile)
	if err != nil {
		panic(err)
	}

	paragraphs := parseMarkdownToParagraphs(string(fileBytes))

	jsonBytes, err := json.Marshal(paragraphs)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(*outputFile, jsonBytes, os.ModeAppend)
	if err != nil {
		panic(err)
	}
}

func parseMarkdownToParagraphs(input string) []*Paragraph {
	paragraphs := make([]*Paragraph, 0)

	inputLines := strings.Split(input, "\n")

	var currentContent bytes.Buffer
	var currentType ParaType
	var currentMetadata string

	for _, line := range inputLines {
		if isPrimaryHeader(line) {
			currentType = PRIMARY_HEADER
			currentContent.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "#")))
			currentMetadata = ""
			addParagraph(&paragraphs, &currentType, &currentContent, &currentMetadata)
			currentType = TEXT
		} else if isSecondaryHeader(line) {
			currentType = SECONDARY_HEADER
			currentContent.WriteString(strings.TrimSpace(strings.TrimPrefix(line, "##")))
			currentMetadata = ""
			addParagraph(&paragraphs, &currentType, &currentContent, &currentMetadata)
			currentType = TEXT
		} else if isCode(line) {
			if currentType != CODE {
				currentType = CODE
				currentMetadata = getLanguage(line)
			} else {
				addParagraph(&paragraphs, &currentType, &currentContent, &currentMetadata)
				currentType = TEXT
			}
		} else {
			if currentType == CODE {
				currentContent.WriteString(strings.TrimSpace(line))
				currentContent.WriteString("\n")
			} else {
				currentContent.WriteString(strings.TrimSpace(line))
				currentMetadata = ""
				addParagraph(&paragraphs, &currentType, &currentContent, &currentMetadata)
			}
		}
	}

	return paragraphs
}

func isPrimaryHeader(input string) bool {
	match, _ := regexp.MatchString(`^#\s+.*$`, input)
	return match
}

func isSecondaryHeader(input string) bool {
	match, _ := regexp.MatchString(`^##\s+.*$`, input)
	return match
}

func isCode(input string) bool {
	match, _ := regexp.MatchString("^```", input)
	return match
}

func getLanguage(input string) string {
	lang := strings.TrimPrefix(strings.Split(input, " ")[0], "```")
	return strings.TrimSpace(lang)
}

func addParagraph(paragraphs *[]*Paragraph, currentType *ParaType, currentContent *bytes.Buffer, currentMetadata *string) {
	if currentContent.Len() > 0 {
		p := Paragraph{
			Type:     *currentType,
			Content:  strings.TrimSpace(currentContent.String()),
			Metadata: *currentMetadata,
		}

		*paragraphs = append(*paragraphs, &p)

		// reset currentContent
		(*currentContent).Reset()
	}
}
