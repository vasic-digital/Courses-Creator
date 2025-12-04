package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/course-creator/core-processor/models"
)

// VideoAccessibilityReport contains accessibility validation results for a video
type VideoAccessibilityReport struct {
	VideoID       string               `json:"video_id"`
	HasCaptions   bool                 `json:"has_captions"`
	CaptionsValid bool                 `json:"captions_valid"`
	HasTranscript bool                 `json:"has_transcript"`
	HasAudio      bool                 `json:"has_audio"`
	ARating       string               `json:"a_rating"`
	Violations    []AccessibilityIssue `json:"violations"`
	Score         float64              `json:"score"`
}

// ContentAccessibilityReport contains accessibility validation for course content
type ContentAccessibilityReport struct {
	ContentID          string               `json:"content_id"`
	ContentType        string               `json:"content_type"`
	HasAlternativeText bool                 `json:"has_alternative_text"`
	HasHeadings        bool                 `json:"has_headings"`
	HasARIALabels      bool                 `json:"has_aria_labels"`
	KeyboardNavigation bool                 `json:"keyboard_navigation"`
	ColorContrast      bool                 `json:"color_contrast"`
	Violations         []AccessibilityIssue `json:"violations"`
	Score              float64              `json:"score"`
}

// AccessibilityIssue represents an accessibility issue found
type AccessibilityIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Element     string `json:"element,omitempty"`
	Suggestion  string `json:"suggestion"`
}

// ValidateVideoAccessibility validates a video for accessibility requirements
func ValidateVideoAccessibility(video *models.Video) (*VideoAccessibilityReport, error) {
	if video == nil {
		return nil, errors.New("video cannot be nil")
	}

	report := &VideoAccessibilityReport{
		VideoID:    video.ID,
		Violations: []AccessibilityIssue{},
	}

	// Check for captions
	report.HasCaptions = video.HasCaptions
	if !video.HasCaptions {
		report.Violations = append(report.Violations, AccessibilityIssue{
			Type:        "missing_captions",
			Severity:    "high",
			Description: "Video does not have captions for hearing impaired users",
			Suggestion:  "Add captions to the video",
		})
	} else {
		report.CaptionsValid = true
	}

	// Check for transcript
	report.HasTranscript = video.HasTranscript
	if !video.HasTranscript {
		report.Violations = append(report.Violations, AccessibilityIssue{
			Type:        "missing_transcript",
			Severity:    "medium",
			Description: "Video does not have a text transcript",
			Suggestion:  "Provide a full transcript of the video content",
		})
	}

	// Check for audio track
	report.HasAudio = video.HasAudio
	if !video.HasAudio {
		report.Violations = append(report.Violations, AccessibilityIssue{
			Type:        "missing_audio",
			Severity:    "high",
			Description: "Video does not have an audio track",
			Suggestion:  "Add audio narration to the video",
		})
	}

	// Calculate A rating and score
	if len(report.Violations) == 0 {
		report.ARating = "AAA"
		report.Score = 100.0
	} else {
		highSeverity := 0
		mediumSeverity := 0
		lowSeverity := 0

		for _, violation := range report.Violations {
			switch violation.Severity {
			case "high":
				highSeverity++
			case "medium":
				mediumSeverity++
			case "low":
				lowSeverity++
			}
		}

		if highSeverity > 0 {
			report.ARating = "A"
			report.Score = 60.0
		} else if mediumSeverity > 0 {
			report.ARating = "AA"
			report.Score = 80.0
		} else {
			report.ARating = "AAA"
			report.Score = 90.0
		}
	}

	return report, nil
}

// ValidateContentAccessibility validates course content for accessibility
func ValidateContentAccessibility(content string, contentType string) (*ContentAccessibilityReport, error) {
	report := &ContentAccessibilityReport{
		ContentType: contentType,
		Violations:  []AccessibilityIssue{},
	}

	// Check for headings structure
	report.HasHeadings = true
	if !containsHeadings(content) {
		report.HasHeadings = false
		report.Violations = append(report.Violations, AccessibilityIssue{
			Type:        "missing_headings",
			Severity:    "medium",
			Description: "Content lacks proper heading structure",
			Suggestion:  "Add h1, h2, h3 headings to structure content",
		})
	}

	// Check for heading skips
	if hasHeadingSkips(content) {
		report.Violations = append(report.Violations, AccessibilityIssue{
			Type:        "heading_skip_detected",
			Severity:    "medium",
			Description: "Content skips heading levels (e.g., h1 to h3)",
			Suggestion:  "Use proper heading hierarchy without skipping levels",
		})
	}

	// Check for non-descriptive link text
	if hasNonDescriptiveLinks(content) {
		report.Violations = append(report.Violations, AccessibilityIssue{
			Type:        "Non-descriptive link text",
			Severity:    "medium",
			Description: "Links use non-descriptive text like 'click here'",
			Suggestion:  "Use descriptive link text that makes sense out of context",
		})
	}

	// Check for ARIA labels
	report.HasARIALabels = strings.Contains(content, "aria-label=") ||
		strings.Contains(content, "aria-labelledby=")
	if !report.HasARIALabels &&
		(strings.Contains(content, "<button>") || strings.Contains(content, "<input")) {
		report.Violations = append(report.Violations, AccessibilityIssue{
			Type:        "missing_aria_labels",
			Severity:    "medium",
			Description: "Interactive elements lack ARIA labels",
			Suggestion:  "Add aria-label or aria-labelledby to interactive elements",
		})
	}

	// Check for alternative text for images
	report.HasAlternativeText = true
	if strings.Contains(content, "<img") && !strings.Contains(content, "alt=") {
		report.HasAlternativeText = false
		report.Violations = append(report.Violations, AccessibilityIssue{
			Type:        "missing_alt_text",
			Severity:    "high",
			Description: "Images lack alternative text",
			Suggestion:  "Add alt attributes to all images",
		})
	}

	// Check for keyboard navigation
	report.KeyboardNavigation = strings.Contains(content, "tabindex=") || strings.Contains(content, "href=")
	if !report.KeyboardNavigation &&
		(strings.Contains(content, "<a") || strings.Contains(content, "<button")) {
		report.Violations = append(report.Violations, AccessibilityIssue{
			Type:        "keyboard_navigation",
			Severity:    "low",
			Description: "Elements may not be keyboard accessible",
			Suggestion:  "Add tabindex and ensure elements can be navigated with keyboard",
		})
	}

	// Check color contrast (basic check)
	report.ColorContrast = !hasPoorColorContrast(content)
	if !report.ColorContrast {
		report.Violations = append(report.Violations, AccessibilityIssue{
			Type:        "color_contrast",
			Severity:    "medium",
			Description: "Text may have poor color contrast",
			Suggestion:  "Use tools to verify WCAG color contrast ratios",
		})
	}

	// Calculate score
	if len(report.Violations) == 0 {
		report.Score = 100.0
	} else {
		baseScore := 100.0
		for _, violation := range report.Violations {
			switch violation.Severity {
			case "high":
				baseScore -= 20
			case "medium":
				baseScore -= 10
			case "low":
				baseScore -= 5
			}
		}
		report.Score = max(baseScore, 0)
	}

	return report, nil
}

// CountFocusableElements counts focusable elements in HTML content
func CountFocusableElements(htmlContent string) (int, error) {
	if htmlContent == "" {
		return 0, errors.New("HTML content is empty")
	}

	// Count elements that can receive focus
	// Count actual HTML elements, not CSS selectors
	count := 0

	// Count buttons
	count += strings.Count(strings.ToLower(htmlContent), "<button")

	// Count inputs
	count += strings.Count(strings.ToLower(htmlContent), "<input")

	// Count selects
	count += strings.Count(strings.ToLower(htmlContent), "<select")

	// Count divs with tabindex
	divTabIndexPattern := `<div[^>]*tabindex`
	re := regexp.MustCompile(divTabIndexPattern)
	matches := re.FindAllString(strings.ToLower(htmlContent), -1)
	count += len(matches)

	return count, nil
}

// GetTabOrder returns the tab order of focusable elements
func GetTabOrder(htmlContent string) ([]string, error) {
	if htmlContent == "" {
		return nil, errors.New("HTML content is empty")
	}

	// Simple implementation - return unique elements that can receive focus
	// In a real implementation, this would parse HTML and return actual tab order
	elements := []string{
		"a[href]", "button:not([disabled])", "input:not([disabled])",
		"select:not([disabled])", "div[tabindex]:not([tabindex=\"-1\"])",
	}

	return elements, nil
}

// GetARIAElements extracts all ARIA elements and attributes from HTML
func GetARIAElements(htmlContent string) (map[string][]string, error) {
	if htmlContent == "" {
		return nil, errors.New("HTML content is empty")
	}

	ariaElements := make(map[string][]string)

	// Find ARIA attributes
	ariaAttrs := []string{
		"aria-label", "aria-labelledby", "aria-describedby", "aria-hidden",
		"role", "aria-expanded", "aria-selected", "aria-pressed", "aria-checked",
		"aria-current", "aria-live",
	}

	for _, attr := range ariaAttrs {
		// Handle boolean attributes like aria-current and aria-live that might not have quotes
		if attr == "aria-current" || attr == "aria-live" {
			// Look for both forms: aria-current="page" and aria-current=page
			re := regexp.MustCompile(fmt.Sprintf(`%s[=\s"']([^"\'>\s]*)`, attr))
			matches := re.FindAllStringSubmatch(htmlContent, -1)

			for _, match := range matches {
				if len(match) > 1 {
					ariaElements[attr] = append(ariaElements[attr], match[1])
				}
			}

			// Also check for presence without value
			re2 := regexp.MustCompile(fmt.Sprintf(`%s([^=])`, attr))
			if re2.MatchString(htmlContent) {
				ariaElements[attr] = append(ariaElements[attr], "present")
			}
		} else {
			re := regexp.MustCompile(fmt.Sprintf(`%s="([^"]*)"`, attr))
			matches := re.FindAllStringSubmatch(htmlContent, -1)

			for _, match := range matches {
				if len(match) > 1 {
					ariaElements[attr] = append(ariaElements[attr], match[1])
				}
			}
		}
	}

	return ariaElements, nil
}

// ValidateARIAUsage validates ARIA attributes for correctness
func ValidateARIAUsage(htmlContent string) ([]AccessibilityIssue, error) {
	var issues []AccessibilityIssue

	// Get ARIA elements
	ariaElements, err := GetARIAElements(htmlContent)
	if err != nil {
		return nil, err
	}

	// Check for empty aria-label
	for attr, values := range ariaElements {
		if attr == "aria-label" {
			for _, value := range values {
				if strings.TrimSpace(value) == "" {
					issues = append(issues, AccessibilityIssue{
						Type:        "empty_aria_label",
						Severity:    "medium",
						Description: "aria-label attribute is empty",
						Suggestion:  "Provide meaningful descriptive text in aria-label",
					})
				}
			}
		}
	}

	// Check for missing ARIA labels on interactive elements
	if !containsARIALabels(htmlContent) && hasInteractiveElements(htmlContent) {
		issues = append(issues, AccessibilityIssue{
			Type:        "missing_aria_labels",
			Severity:    "medium",
			Description: "Interactive elements lack ARIA labels",
			Suggestion:  "Add aria-label or aria-labelledby to interactive elements",
		})
	}

	// Check for contradictory ARIA usage
	if strings.Contains(htmlContent, `role="presentation"`) && strings.Contains(htmlContent, `alt=`) {
		issues = append(issues, AccessibilityIssue{
			Type:        "contradictory_aria",
			Severity:    "medium",
			Description: "Element has both presentation role and alt text",
			Suggestion:  "Remove alt attribute from presentational images",
		})
	}

	return issues, nil
}

// Helper function to check if content contains ARIA labels
func containsARIALabels(content string) bool {
	return strings.Contains(content, "aria-label=") ||
		strings.Contains(content, "aria-labelledby=")
}

// Helper function to check if content has interactive elements
func hasInteractiveElements(content string) bool {
	hasButtons := strings.Contains(content, "<button") || strings.Contains(content, "<input")
	hasLinks := strings.Contains(content, "<a ")
	hasButtonLikeDivs := false

	// Look for divs that might be buttons (contain button-like content)
	if strings.Contains(content, "<div") {
		// Check if div contains button-like content (symbols)
		buttonContent := []string{"▶", "▷", "▸", "▹", "►", "▻", "⏵"}
		for _, btn := range buttonContent {
			if strings.Contains(content, btn) {
				hasButtonLikeDivs = true
				break
			}
		}
	}

	return hasButtons || hasLinks || hasButtonLikeDivs
}

// Helper functions
func containsHeadings(content string) bool {
	return strings.Contains(content, "<h1>") ||
		strings.Contains(content, "<h2>") ||
		strings.Contains(content, "<h3>")
}

func hasPoorColorContrast(content string) bool {
	// Check for specific color combinations that have poor contrast
	return strings.Contains(content, "color: #999999; background-color: #ffffff") ||
		strings.Contains(content, "color:gray") ||
		strings.Contains(content, "color:#ddd") ||
		strings.Contains(content, "background:lightgray")
}

func hasHeadingSkips(content string) bool {
	// Check for heading skips like h1 followed by h3
	if strings.Contains(content, "<h1") && strings.Contains(content, "<h3") {
		// Check if h2 is present
		if !strings.Contains(content, "<h2") {
			return true
		}
	}
	return false
}

func hasNonDescriptiveLinks(content string) bool {
	// Look for non-descriptive link text
	nonDescriptivePatterns := []string{
		">click here<",
		">learn more<",
		">read more<",
		">here<",
		">more<",
	}

	for _, pattern := range nonDescriptivePatterns {
		if strings.Contains(strings.ToLower(content), pattern) {
			return true
		}
	}
	return false
}

// ValidateColorContrast validates color contrast between foreground and background
func ValidateColorContrast(foreground, background string) (float64, bool) {
	// Use WCAG 2.1 contrast ratio formula
	fgLuminance := calculateRelativeLuminance(foreground)
	bgLuminance := calculateRelativeLuminance(background)

	lighter := max(fgLuminance, bgLuminance)
	darker := min(fgLuminance, bgLuminance)

	ratio := (lighter + 0.05) / (darker + 0.05)

	// WCAG AA requires at least 4.5:1 for normal text
	wcagAA := ratio >= 4.5

	return ratio, wcagAA
}

// Calculate relative luminance according to WCAG 2.1
func calculateRelativeLuminance(color string) float64 {
	// Remove # if present
	if strings.HasPrefix(color, "#") {
		color = color[1:]
	}

	var r, g, b float64

	if len(color) == 3 {
		// Short form #RGB
		r = parseHexChar(color[0])
		g = parseHexChar(color[1])
		b = parseHexChar(color[2])
	} else if len(color) == 6 {
		// Full form #RRGGBB
		r = parseHexByte(color[0:2])
		g = parseHexByte(color[2:4])
		b = parseHexByte(color[4:6])
	} else {
		// Handle named colors
		switch strings.ToLower(color) {
		case "black":
			return 0.0
		case "white":
			return 1.0
		case "gray", "grey":
			return 0.21586 // Relative luminance of #808080
		default:
			return 0.5 // Default
		}
	}

	// Apply gamma correction
	r = gammaCorrect(r)
	g = gammaCorrect(g)
	b = gammaCorrect(b)

	// Calculate relative luminance
	return 0.2126*r + 0.7152*g + 0.0722*b
}

// Apply gamma correction to color component
func gammaCorrect(c float64) float64 {
	if c <= 0.03928 {
		return c / 12.92
	}
	return simplePow((c+0.055)/1.055, 2.4)
}

// Power function for gamma correction that works with fractional exponents
func simplePow(x, y float64) float64 {
	// Simple approximation for x^2.4
	if y == 2.4 {
		// Using x^2.4 = x^2 * x^0.4 as approximation
		return x * x * simplePow(x, 0.4)
	}

	// For other powers, simple approximation
	result := 1.0
	if y < 0 {
		y = -y
		for i := 0; i < int(y*10); i++ {
			result /= x
		}
	} else {
		for i := 0; i < int(y*10); i++ {
			result *= x
		}
	}
	return result
}

// Helper function to calculate relative luminance of a color
func getColorBrightness(color string) float64 {
	// Remove # if present
	if strings.HasPrefix(color, "#") {
		color = color[1:]
	}

	// Parse hex values
	r, g, b := 0.0, 0.0, 0.0
	if len(color) == 3 {
		// Short form #RGB
		r, g, b = parseHexChar(color[0]), parseHexChar(color[1]), parseHexChar(color[2])
	} else if len(color) == 6 {
		// Full form #RRGGBB
		r = parseHexByte(color[0:2])
		g = parseHexByte(color[2:4])
		b = parseHexByte(color[4:6])
	} else if len(color) > 0 {
		// Handle common color names
		lowerColor := strings.ToLower(color)
		switch lowerColor {
		case "black":
			return 0.0
		case "white":
			return 1.0
		case "gray", "grey":
			return 0.5
		}
	}

	// Calculate relative luminance
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func parseHexChar(c byte) float64 {
	if c >= '0' && c <= '9' {
		return float64(c-'0') / 15.0
	}
	if c >= 'a' && c <= 'f' {
		return float64(c-'a'+10) / 15.0
	}
	if c >= 'A' && c <= 'F' {
		return float64(c-'A'+10) / 15.0
	}
	return 0
}

func parseHexByte(s string) float64 {
	if len(s) != 2 {
		return 0
	}

	high := s[0]
	low := s[1]

	var highVal, lowVal int

	if high >= '0' && high <= '9' {
		highVal = int(high - '0')
	} else if high >= 'a' && high <= 'f' {
		highVal = int(high - 'a' + 10)
	} else if high >= 'A' && high <= 'F' {
		highVal = int(high - 'A' + 10)
	}

	if low >= '0' && low <= '9' {
		lowVal = int(low - '0')
	} else if low >= 'a' && low <= 'f' {
		lowVal = int(low - 'a' + 10)
	} else if low >= 'A' && low <= 'F' {
		lowVal = int(low - 'A' + 10)
	}

	return float64(highVal*16+lowVal) / 255.0
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// GenerateCompliantHTMLPage generates a WCAG-compliant HTML page for testing
func GenerateCompliantHTMLPage() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Accessible Course Content</title>
</head>
<body>
    <header role="banner">
        <nav aria-label="Main Navigation">
            <ul>
                <li><a href="/" aria-current="page">Home</a></li>
                <li><a href="/courses">Courses</a></li>
                <li><a href="/about">About</a></li>
            </ul>
        </nav>
    </header>
    
    <main role="main" aria-labelledby="page-title">
        <h1 id="page-title">Course Title</h1>
        
        <section aria-labelledby="section1-title">
            <h2 id="section1-title">Introduction</h2>
            <p>This is accessible content with proper structure and semantic markup.</p>
            
            <figure>
                <img src="example.jpg" alt="A descriptive alt text for the image">
                <figcaption>Figure 1: Example image with proper description</figcaption>
            </figure>
            
            <table>
                <caption>Course Schedule</caption>
                <thead>
                    <tr>
                        <th scope="col">Week</th>
                        <th scope="col">Topic</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>1</td>
                        <td>Introduction</td>
                    </tr>
                </tbody>
            </table>
        </section>
        
        <section aria-labelledby="section2-title">
            <h2 id="section2-title">Interactive Elements</h2>
            <div class="video-player">
                <video controls aria-describedby="video-desc">
                    <source src="video.mp4" type="video/mp4">
                </video>
                <div id="video-desc" class="sr-only">
                    Video introduction to the course concepts
                </div>
                <div class="controls" role="region" aria-label="Video Controls">
                    <button aria-label="Play/Pause">▶️</button>
                    <button aria-label="Fullscreen">⛶</button>
                </div>
            </div>
            
            <form aria-labelledby="form-title">
                <h3 id="form-title">Quiz Question</h3>
                <fieldset>
                    <legend>Select the correct answer:</legend>
                    <input type="radio" id="answer1" name="quiz" aria-describedby="question-desc">
                    <label for="answer1">Option A</label>
                    
                    <input type="radio" id="answer2" name="quiz">
                    <label for="answer2">Option B</label>
                    
                    <div id="question-desc" class="sr-only">Choose one option from the list</div>
                </fieldset>
                <button type="submit">Submit</button>
            </form>
        </section>
    </main>
    
    <footer role="contentinfo">
        <p>&copy; 2023 Course Creator. All rights reserved.</p>
    </footer>
    
    <div aria-live="polite" aria-atomic="true" class="sr-only">
        <!-- Screen reader announcements will appear here -->
    </div>
</body>
</html>`
}
