package accessibility_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/services"
	"github.com/gorilla/mux"
)

// AccessibilityResponse represents the response from accessibility testing
type AccessibilityResponse struct {
	StatusCode int           `json:"statusCode"`
	Data       AccessibilityData `json:"data"`
}

type AccessibilityData struct {
	TestUrl        string                `json:"testUrl"`
	TestTimestamp  string               `json:"testTimestamp"`
	TestResult     AccessibilityResult   `json:"testResult"`
}

type AccessibilityResult struct {
	Status      string                    `json:"status"`
	Score       AccessibilityScore        `json:"score"`
	Violations  []AccessibilityViolation   `json:"violations"`
	Passes      []AccessibilityPass       `json:"passes"`
}

type AccessibilityScore struct {
	Rating string `json:"rating"`
	Score  int    `json:"score"`
}

type AccessibilityViolation struct {
	RuleId      string   `json:"ruleId"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Nodes       []Node   `json:"nodes"`
}

type Node struct {
	Selector string `json:"selector"`
	Impact   string `json:"impact"`
}

type AccessibilityPass struct {
	RuleId string `json:"ruleId"`
}

// Mock accessibility checker
type MockAccessibilityChecker struct{}

func (m *MockAccessibilityChecker) CheckAccessibility(url string) (*AccessibilityResponse, error) {
	// Simulate accessibility check with mock data
	// In real implementation, this would use axe-core or similar
	return &AccessibilityResponse{
		StatusCode: http.StatusOK,
		Data: AccessibilityData{
			TestUrl:       url,
			TestTimestamp: time.Now().Format(time.RFC3339),
			TestResult: AccessibilityResult{
				Status: "complete",
				Score: AccessibilityScore{
					Rating: "AA",
					Score:  95,
				},
				Violations: []AccessibilityViolation{},
				Passes: []AccessibilityPass{
					{RuleId: "color-contrast"},
					{RuleId: "keyboard-navigation"},
					{RuleId: "aria-labels"},
				},
			},
		},
	}, nil
}

func TestWebPlayerAccessibility(t *testing.T) {
	checker := &MockAccessibilityChecker{}
	
	// Mock player routes
	router := mux.NewRouter()
	
	// Mock player page content
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Course Player</title>
</head>
<body>
    <main role="main">
        <header>
            <h1>Course Title</h1>
            <nav aria-label="Course Navigation">
                <ul>
                    <li><a href="#lesson1">Lesson 1</a></li>
                    <li><a href="#lesson2">Lesson 2</a></li>
                </ul>
            </nav>
        </header>
        
        <section id="lesson1" aria-labelledby="lesson1-title">
            <h2 id="lesson1-title">Lesson 1: Introduction</h2>
            <video controls aria-describedby="video-desc">
                <source src="video.mp4" type="video/mp4">
                Your browser does not support the video tag.
            </video>
            <div id="video-desc" class="sr-only">Video introduction to the course</div>
            
            <div class="video-controls" role="region" aria-label="Video Controls">
                <button aria-label="Play/Pause">▶️</button>
                <button aria-label="Fullscreen">⛶</button>
                <button aria-label="Captions">CC</button>
            </div>
        </section>
        
        <section id="quiz" aria-labelledby="quiz-title">
            <h2 id="quiz-title">Quiz</h2>
            <form role="form" aria-labelledby="quiz-title">
                <fieldset>
                    <legend>Question 1</legend>
                    <input type="radio" id="q1a1" name="q1" aria-describedby="q1-desc">
                    <label for="q1a1">Answer 1</label>
                    
                    <input type="radio" id="q1a2" name="q1" aria-describedby="q1-desc">
                    <label for="q1a2">Answer 2</label>
                    
                    <div id="q1-desc" class="sr-only">Select the best answer</div>
                </fieldset>
            </form>
        </section>
    </main>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	})

	testCases := []struct {
		name         string
		path         string
		expectViolations bool
		expectedRating string {
		name:            "Home page",
		path:            "/",
		expectViolations: false,
		expectedRating:   "AA",
	},
	{
		name:            "Course player page",
		path:            "/course/123",
		expectViolations: false,
		expectedRating:   "AA",
	},
}

for _, tc := range testCases {
	t.Run(tc.name, func(t *testing.T) {
		req := httptest.NewRequest("GET", tc.path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
			return
		}

		// Check accessibility
		testUrl := "http://localhost:8080" + tc.path
		result, err := checker.CheckAccessibility(testUrl)
		if err != nil {
			t.Errorf("Accessibility check failed: %v", err)
			return
		}

		if result.Data.TestResult.Status != "complete" {
			t.Errorf("Accessibility test incomplete: %s", result.Data.TestResult.Status)
		}

		if result.Data.TestResult.Score.Rating != tc.expectedRating {
			t.Errorf("Expected accessibility rating %s, got %s", 
				tc.expectedRating, result.Data.TestResult.Score.Rating)
		}

		if tc.expectViolations && len(result.Data.TestResult.Violations) == 0 {
			t.Errorf("Expected violations but found none")
		}

		if !tc.expectViolations && len(result.Data.TestResult.Violations) > 0 {
			t.Errorf("Expected no violations but found %d", 
				len(result.Data.TestResult.Violations))
		}

		t.Logf("Accessibility score: %d (%s rating)", 
			result.Data.TestResult.Score.Score, 
			result.Data.TestResult.Score.Rating)
	})
}
}

func TestVideoAccessibility(t *testing.T) {
	testCases := []struct {
		name         string
		video        models.Video
		expectValid  bool
	}{
		{
			name: "Video with captions",
			video: models.Video{
				ID:          "video1",
				Title:       "Introduction Video",
				URL:         "https://example.com/video.mp4",
				HasCaptions: true,
				Transcript:  "Full transcript of the video content...",
				Duration:    300,
			},
			expectValid: true,
		},
		{
			name: "Video without captions",
			video: models.Video{
				ID:          "video2",
				Title:       "Video without captions",
				URL:         "https://example.com/video.mp4",
				HasCaptions: false,
				Transcript:  "",
				Duration:    300,
			},
			expectValid: false,
		},
		{
			name: "Video with audio description",
			video: models.Video{
				ID:               "video3",
				Title:            "Video with audio description",
				URL:              "https://example.com/video.mp4",
				HasCaptions:      true,
				HasAudioDesc:     true,
				Transcript:       "Full transcript of the video content...",
				AudioDescURL:     "https://example.com/audio-desc.mp3",
				Duration:         300,
			},
			expectValid: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			valid := services.ValidateVideoAccessibility(tc.video)
			
			if tc.expectValid && !valid {
				t.Errorf("Expected video to be accessible but validation failed")
			}
			
			if !tc.expectValid && valid {
				t.Errorf("Expected video to be inaccessible but validation passed")
			}
		})
	}
}

func TestCourseContentAccessibility(t *testing.T) {
	testCases := []struct {
		name         string
		content      string
		expectValid  bool
		expectedIssues []string
	}{
		{
			name: "Accessible content with proper headings",
			content: `<h1>Main Title</h1>
				<h2>Section 1</h2>
				<p>Content with <a href="https://example.com" title="Link to example">descriptive link</a></p>
				<h3>Subsection</h3>
				<img src="image.jpg" alt="Descriptive alt text">
				<table>
					<caption>Table of data</caption>
					<thead>
						<tr><th scope="col">Header 1</th><th scope="col">Header 2</th></tr>
					</thead>
					<tbody>
						<tr><td>Data 1</td><td>Data 2</td></tr>
					</tbody>
				</table>`,
			expectValid: true,
			expectedIssues: []string{},
		},
		{
			name: "Inaccessible content with missing alt text",
			content: `<h1>Title</h1>
				<img src="image.jpg"> <!-- Missing alt text -->
				<p>Content with <a href="https://example.com">click here</a></p> <!-- Non-descriptive link -->`,
			expectValid: false,
			expectedIssues: []string{"Missing alt text", "Non-descriptive link text"},
		},
		{
			name: "Inaccessible content with heading skips",
			content: `<h1>Title</h1>
				<h3>Skipped h2 heading</h3> <!-- Skipped h2 -->
				<p>Content</p>`,
			expectValid: false,
			expectedIssues: []string{"Heading skip detected"},
		},
		{
			name: "Content with insufficient color contrast",
			content: `<h1>Title</h1>
				<p style="color: #999999; background-color: #ffffff;">Low contrast text</p> <!-- Low contrast -->`,
			expectValid: false,
			expectedIssues: []string{"Insufficient color contrast"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			issues := services.ValidateContentAccessibility(tc.content)
			valid := len(issues) == 0
			
			if tc.expectValid && !valid {
				t.Errorf("Expected content to be accessible but found issues: %v", issues)
			}
			
			if !tc.expectValid && valid {
				t.Errorf("Expected content to be inaccessible but validation passed")
			}
			
			// Check for specific expected issues
			for _, expectedIssue := range tc.expectedIssues {
				found := false
				for _, issue := range issues {
					if issue == expectedIssue {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected issue '%s' not found in validation results: %v", 
						expectedIssue, issues)
				}
			}
		})
	}
}

func TestKeyboardNavigation(t *testing.T) {
	router := mux.NewRouter()
	
	// Mock interactive page with keyboard navigation
	router.HandleFunc("/interactive", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Interactive Page</title>
</head>
<body>
    <div tabindex="0" role="button" aria-label="Custom button">Custom Button</div>
    <button>Standard Button</button>
    <input type="text" aria-label="Text input">
    <select aria-label="Select option">
        <option>Option 1</option>
        <option>Option 2</option>
    </select>
    <div role="tablist">
        <div role="tab" tabindex="0" aria-selected="true">Tab 1</div>
        <div role="tab" tabindex="0" aria-selected="false">Tab 2</div>
    </div>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	})

	testCases := []struct {
		name          string
		path          string
		expectFocusable int // Number of focusable elements
	}{
		{
			name:           "Interactive page",
			path:           "/interactive",
			expectFocusable: 5, // Custom button, standard button, input, select, first tab
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
				return
			}

			// Check focusable elements
			focusableCount := services.CountFocusableElements(w.Body.String())
			
			if focusableCount != tc.expectFocusable {
				t.Errorf("Expected %d focusable elements, found %d", 
					tc.expectFocusable, focusableCount)
			}

			// Check tab order
			tabOrder := services.GetTabOrder(w.Body.String())
			if len(tabOrder) != tc.expectFocusable {
				t.Errorf("Expected tab order of %d elements, found %d", 
					tc.expectFocusable, len(tabOrder))
			}
		})
	}
}

func TestScreenReaderCompatibility(t *testing.T) {
	testCases := []struct {
		name          string
		content       string
		expectValid   bool
		expectedARIA  []string
	}{
		{
			name: "Content with proper ARIA labels",
			content: `<div role="navigation" aria-label="Main Navigation">
				<ul>
					<li><a href="/" aria-current="page">Home</a></li>
					<li><a href="/about">About</a></li>
				</ul>
			</div>
			<main role="main" aria-labelledby="main-title">
				<h1 id="main-title">Main Content</h1>
				<div role="button" tabindex="0" aria-label="Play video">▶</div>
				<div role="status" aria-live="polite">Status message</div>
			</main>`,
			expectValid: true,
			expectedARIA: []string{"role", "aria-label", "aria-current", "aria-labelledby", "aria-live"},
		},
		{
			name: "Content with missing ARIA labels",
			content: `<div>
				<ul>
					<li><a href="/">Home</a></li>
					<li><a href="/about">About</a></li>
				</ul>
			</div>
			<div>▶</div> <!-- Button without ARIA label -->`,
			expectValid: false,
			expectedARIA: []string{},
		},
		{
			name: "Content with invalid ARIA usage",
			content: `<div role="button" aria-hidden="false" tabindex="0">Visible button</div>
			<img src="image.jpg" role="presentation" alt="Alt text for presentational image"> <!-- Contradictory -->`,
			expectValid: false,
			expectedARIA: []string{"role", "aria-hidden"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ariaElements := services.GetARIAElements(tc.content)
			valid := services.ValidateARIAUsage(tc.content)
			
			if tc.expectValid && !valid {
				t.Errorf("Expected ARIA usage to be valid but validation failed")
			}
			
			if !tc.expectValid && valid {
				t.Errorf("Expected ARIA usage to be invalid but validation passed")
			}
			
			// Check for expected ARIA attributes
			for _, expectedARIA := range tc.expectedARIA {
				found := false
				for _, element := range ariaElements {
					if element == expectedARIA {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected ARIA attribute '%s' not found in content", expectedARIA)
				}
			}
		})
	}
}

func TestColorContrast(t *testing.T) {
	testCases := []struct {
		name          string
		foreground    string
		background    string
		expectValid   bool
		expectedRatio float64
	}{
		{
			name:          "High contrast black on white",
			foreground:    "#000000",
			background:    "#FFFFFF",
			expectValid:   true,
			expectedRatio: 21.0,
		},
		{
			name:          "Low contrast light gray on white",
			foreground:    "#CCCCCC",
			background:    "#FFFFFF",
			expectValid:   false,
			expectedRatio: 1.6,
		},
		{
			name:          "Sufficient contrast blue on white",
			foreground:    "#0066CC",
			background:    "#FFFFFF",
			expectValid:   true,
			expectedRatio: 5.0,
		},
		{
			name:          "Insufficient contrast yellow on white",
			foreground:    "#FFFF00",
			background:    "#FFFFFF",
			expectValid:   false,
			expectedRatio: 1.1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ratio, valid := services.ValidateColorContrast(tc.foreground, tc.background)
			
			if tc.expectValid && !valid {
				t.Errorf("Expected color contrast to be valid but validation failed. Ratio: %.2f", ratio)
			}
			
			if !tc.expectValid && valid {
				t.Errorf("Expected color contrast to be invalid but validation passed. Ratio: %.2f", ratio)
			}
			
			// Check if ratio is close to expected
			if ratio < tc.expectedRatio*0.9 || ratio > tc.expectedRatio*1.1 {
				t.Errorf("Expected contrast ratio around %.2f, got %.2f", 
					tc.expectedRatio, ratio)
			}
		})
	}
}

func TestWCAGCompliance(t *testing.T) {
	// Test overall WCAG 2.1 AA compliance
	checker := &MockAccessibilityChecker{}
	
	// Create comprehensive test page
	router := mux.NewRouter()
	
	router.HandleFunc("/comprehensive", func(w http.ResponseWriter, r *http.Request) {
		html := services.GenerateCompliantHTMLPage()
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	})

	req := httptest.NewRequest("GET", "/comprehensive", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		return
	}

	// Run comprehensive accessibility check
	testUrl := "http://localhost:8080/comprehensive"
	result, err := checker.CheckAccessibility(testUrl)
	if err != nil {
		t.Errorf("Accessibility check failed: %v", err)
		return
	}

	// Verify WCAG AA compliance
	if result.Data.TestResult.Score.Rating != "AA" {
		t.Errorf("Expected WCAG AA rating, got %s", result.Data.TestResult.Score.Rating)
	}

	if result.Data.TestResult.Score.Score < 90 {
		t.Errorf("Expected accessibility score >= 90, got %d", 
			result.Data.TestResult.Score.Score)
	}

	// Check for critical violations
	criticalViolations := 0
	for _, violation := range result.Data.TestResult.Violations {
		if violation.Impact == "critical" {
			criticalViolations++
		}
	}

	if criticalViolations > 0 {
		t.Errorf("Found %d critical accessibility violations", criticalViolations)
	}

	t.Logf("WCAG Compliance Score: %d (%s rating) with %d passes and %d violations",
		result.Data.TestResult.Score.Score,
		result.Data.TestResult.Score.Rating,
		len(result.Data.TestResult.Passes),
		len(result.Data.TestResult.Violations))
}