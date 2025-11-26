package utils

import (
	"regexp"
	"strings"

	"github.com/course-creator/core-processor/models"
)

// MarkdownParser handles parsing of markdown course content
type MarkdownParser struct {
	headerPattern *regexp.Regexp
	codePattern   *regexp.Regexp
	imagePattern  *regexp.Regexp
}

// NewMarkdownParser creates a new markdown parser
func NewMarkdownParser() *MarkdownParser {
	return &MarkdownParser{
		headerPattern: regexp.MustCompile(`^#+\s+(.+)$`),
		codePattern:   regexp.MustCompile(`(?s)\x60\x60\x60(\w+)?\n(.*?)\n\x60\x60\x60`),
		imagePattern:  regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`),
	}
}

// Parse parses markdown file into course structure
func (p *MarkdownParser) Parse(content string) (*models.ParsedCourse, error) {
	// Extract title from first header
	title := p.extractTitle(content)

	// Extract sections
	sections := p.extractSections(content)

	// Extract metadata
	metadata := p.extractMetadata(content)

	return &models.ParsedCourse{
		Title:       title,
		Description: p.extractDescription(content),
		Sections:    sections,
		Metadata:    metadata,
	}, nil
}

func (p *MarkdownParser) extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if matches := p.headerPattern.FindStringSubmatch(line); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	return "Untitled Course"
}

func (p *MarkdownParser) extractSections(content string) []models.ParsedSection {
	var sections []models.ParsedSection
	lines := strings.Split(content, "\n")
	var currentSection *models.ParsedSection
	var currentContent []string
	order := 0
	firstHeaderFound := false

	for _, line := range lines {
		if matches := p.headerPattern.FindStringSubmatch(line); len(matches) > 1 {
			// Skip the first header (title)
			if !firstHeaderFound {
				firstHeaderFound = true
				continue
			}

			// Save previous section
			if currentSection != nil {
				currentSection.Content = strings.TrimSpace(strings.Join(currentContent, "\n"))
				sections = append(sections, *currentSection)
				order++
			}

			// Start new section
			currentSection = &models.ParsedSection{
				Title: strings.TrimSpace(matches[1]),
				Order: order,
			}
			currentContent = []string{}
		} else if currentSection != nil {
			currentContent = append(currentContent, line)
		}
	}

	// Add last section
	if currentSection != nil {
		currentSection.Content = strings.TrimSpace(strings.Join(currentContent, "\n"))
		sections = append(sections, *currentSection)
	}

	return sections
}

func (p *MarkdownParser) extractDescription(content string) string {
	lines := strings.Split(content, "\n")
	var descLines []string
	titleFound := false
	for _, line := range lines {
		if p.headerPattern.MatchString(line) {
			if !titleFound {
				titleFound = true
				continue
			} else {
				break
			}
		}
		if strings.TrimSpace(line) != "" {
			descLines = append(descLines, line)
		}
	}
	return strings.TrimSpace(strings.Join(descLines, "\n"))
}

func (p *MarkdownParser) extractMetadata(content string) map[string]interface{} {
	// Placeholder for metadata extraction
	// Could parse YAML frontmatter, author comments, etc.
	return map[string]interface{}{
		"author":   "Unknown",
		"language": "en",
		"tags":     []string{},
	}
}
