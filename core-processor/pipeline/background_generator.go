package pipeline

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	storage "github.com/course-creator/core-processor/filestorage"
	"github.com/course-creator/core-processor/models"
	"github.com/course-creator/core-processor/utils"
)

// BackgroundGenerator handles dynamic background generation
type BackgroundGenerator struct {
	config        BackgroundConfig
	storage       storage.StorageInterface
	colorPalettes []ColorPalette
	patterns      []PatternGenerator
}

// BackgroundConfig holds background generation configuration
type BackgroundConfig struct {
	Width      int
	Height     int
	Quality    int
	OutputDir  string
	CacheDir   string
	TempDir    string
	Timeout    time.Duration
	MaxRetries int
}

// ColorPalette represents a color palette for backgrounds
type ColorPalette struct {
	Name    string
	Colors  []color.RGBA
	Weights []float64
	Mood    string
}

// PatternGenerator defines interface for background patterns
type PatternGenerator interface {
	Generate(img *image.RGBA, palette ColorPalette, seed int64) error
	GetName() string
}

// SolidPattern generates solid color backgrounds
type SolidPattern struct{}

// GradientPattern generates gradient backgrounds
type GradientPattern struct{}

// GeometricPattern generates geometric pattern backgrounds
type GeometricPattern struct{}

// NoisePattern generates noise/texture backgrounds
type NoisePattern struct{}

// NewBackgroundGenerator creates a new background generator
func NewBackgroundGenerator(storage storage.StorageInterface) *BackgroundGenerator {
	config := BackgroundConfig{
		Width:      1920,
		Height:     1080,
		Quality:    90,
		OutputDir:  "/tmp/backgrounds",
		CacheDir:   "/tmp/bg_cache",
		TempDir:    "/tmp/bg_temp",
		Timeout:    60 * time.Second,
		MaxRetries: 2,
	}

	// Ensure directories exist
	utils.EnsureDir(config.OutputDir)
	utils.EnsureDir(config.CacheDir)
	utils.EnsureDir(config.TempDir)

	bg := &BackgroundGenerator{
		config:  config,
		storage: storage,
	}

	// Initialize color palettes
	bg.initializeColorPalettes()

	// Initialize patterns
	bg.initializePatterns()

	return bg
}

// NewBackgroundGeneratorWithConfig creates a background generator with custom config
func NewBackgroundGeneratorWithConfig(config BackgroundConfig, storage storage.StorageInterface) *BackgroundGenerator {
	utils.EnsureDir(config.OutputDir)
	utils.EnsureDir(config.CacheDir)
	utils.EnsureDir(config.TempDir)

	bg := &BackgroundGenerator{
		config:  config,
		storage: storage,
	}

	bg.initializeColorPalettes()
	bg.initializePatterns()

	return bg
}

// initializeColorPalettes sets up predefined color palettes
func (bg *BackgroundGenerator) initializeColorPalettes() {
	bg.colorPalettes = []ColorPalette{
		{
			Name: "ocean",
			Colors: []color.RGBA{
				{R: 0, G: 119, B: 190, A: 255},   // Deep blue
				{R: 0, G: 180, B: 216, A: 255},   // Sky blue
				{R: 144, G: 224, B: 239, A: 255}, // Light blue
				{R: 255, G: 255, B: 255, A: 255}, // White
			},
			Weights: []float64{0.4, 0.3, 0.2, 0.1},
			Mood:    "calm",
		},
		{
			Name: "forest",
			Colors: []color.RGBA{
				{R: 34, G: 139, B: 34, A: 255},   // Forest green
				{R: 107, G: 142, B: 35, A: 255},  // Olive green
				{R: 144, G: 238, B: 144, A: 255}, // Light green
				{R: 245, G: 245, B: 220, A: 255}, // Beige
			},
			Weights: []float64{0.3, 0.3, 0.25, 0.15},
			Mood:    "natural",
		},
		{
			Name: "sunset",
			Colors: []color.RGBA{
				{R: 255, G: 94, B: 77, A: 255},  // Coral
				{R: 255, G: 154, B: 0, A: 255},  // Orange
				{R: 237, G: 117, B: 57, A: 255}, // Dark orange
				{R: 255, G: 206, B: 84, A: 255}, // Yellow
			},
			Weights: []float64{0.3, 0.25, 0.25, 0.2},
			Mood:    "energetic",
		},
		{
			Name: "lavender",
			Colors: []color.RGBA{
				{R: 230, G: 230, B: 250, A: 255}, // Lavender
				{R: 216, G: 191, B: 216, A: 255}, // Thistle
				{R: 221, G: 160, B: 221, A: 255}, // Plum
				{R: 238, G: 130, B: 238, A: 255}, // Violet
			},
			Weights: []float64{0.3, 0.25, 0.25, 0.2},
			Mood:    "elegant",
		},
		{
			Name: "professional",
			Colors: []color.RGBA{
				{R: 52, G: 73, B: 94, A: 255},    // Dark blue-gray
				{R: 108, G: 117, B: 125, A: 255}, // Gray
				{R: 189, G: 195, B: 199, A: 255}, // Light gray
				{R: 236, G: 240, B: 241, A: 255}, // Very light gray
			},
			Weights: []float64{0.35, 0.25, 0.25, 0.15},
			Mood:    "professional",
		},
	}
}

// initializePatterns sets up available pattern generators
func (bg *BackgroundGenerator) initializePatterns() {
	bg.patterns = []PatternGenerator{
		&SolidPattern{},
		&GradientPattern{},
		&GeometricPattern{},
		&NoisePattern{},
	}
}

// GenerateBackground generates a background based on content and options
func (bg *BackgroundGenerator) GenerateBackground(ctx context.Context, content string, options models.ProcessingOptions) (string, error) {
	// Choose palette based on content analysis
	palette := bg.selectPalette(content, options)

	// Choose pattern based on quality and preferences
	pattern := bg.selectPattern(options)

	// Generate unique seed for reproducibility
	seed := time.Now().UnixNano() + int64(utils.HashString(content))

	// Generate background
	backgroundPath := filepath.Join(bg.config.OutputDir, fmt.Sprintf("bg_%d_%s.png", seed, pattern.GetName()))

	img := image.NewRGBA(image.Rect(0, 0, bg.config.Width, bg.config.Height))

	if err := pattern.Generate(img, palette, seed); err != nil {
		return "", fmt.Errorf("failed to generate %s pattern: %w", pattern.GetName(), err)
	}

	// Save background
	if err := bg.saveBackground(img, backgroundPath); err != nil {
		return "", fmt.Errorf("failed to save background: %w", err)
	}

	fmt.Printf("Generated %s background with %s palette\n", pattern.GetName(), palette.Name)
	return backgroundPath, nil
}

// selectPalette chooses appropriate color palette based on content and options
func (bg *BackgroundGenerator) selectPalette(content string, options models.ProcessingOptions) ColorPalette {
	// Simple content analysis for palette selection
	content = strings.ToLower(content)

	var selectedPalette ColorPalette

	// Check for mood indicators in content
	if strings.Contains(content, "business") || strings.Contains(content, "professional") || strings.Contains(content, "corporate") {
		selectedPalette = bg.colorPalettes[4] // Professional
	} else if strings.Contains(content, "nature") || strings.Contains(content, "environment") || strings.Contains(content, "green") {
		selectedPalette = bg.colorPalettes[1] // Forest
	} else if strings.Contains(content, "relax") || strings.Contains(content, "calm") || strings.Contains(content, "ocean") {
		selectedPalette = bg.colorPalettes[0] // Ocean
	} else if strings.Contains(content, "energy") || strings.Contains(content, "vibrant") || strings.Contains(content, "sunset") {
		selectedPalette = bg.colorPalettes[2] // Sunset
	} else if strings.Contains(content, "creative") || strings.Contains(content, "art") || strings.Contains(content, "design") {
		selectedPalette = bg.colorPalettes[3] // Lavender
	} else {
		// Default based on quality
		switch options.Quality {
		case "high":
			selectedPalette = bg.colorPalettes[0] // Ocean for high quality
		default:
			selectedPalette = bg.colorPalettes[4] // Professional for standard
		}
	}

	return selectedPalette
}

// selectPattern chooses appropriate pattern based on options
func (bg *BackgroundGenerator) selectPattern(options models.ProcessingOptions) PatternGenerator {
	switch options.BackgroundStyle {
	case "solid":
		return bg.patterns[0] // Solid
	case "gradient":
		return bg.patterns[1] // Gradient
	case "geometric":
		return bg.patterns[2] // Geometric
	case "noise":
		return bg.patterns[3] // Noise
	default:
		// Choose based on quality
		switch options.Quality {
		case "high":
			return bg.patterns[2] // Geometric for high quality
		default:
			return bg.patterns[1] // Gradient for standard
		}
	}
}

// saveBackground saves the generated background image
func (bg *BackgroundGenerator) saveBackground(img *image.RGBA, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}

// SolidPattern implementation
func (p *SolidPattern) Generate(img *image.RGBA, palette ColorPalette, seed int64) error {
	bounds := img.Bounds()

	// Choose color from palette based on weights
	rand.Seed(seed)
	selectedColor := palette.WeightedRandomColor()

	// Fill entire image with selected color
	draw.Draw(img, bounds, &image.Uniform{selectedColor}, image.Point{}, draw.Src)

	return nil
}

func (p *SolidPattern) GetName() string {
	return "solid"
}

// GradientPattern implementation
func (p *GradientPattern) Generate(img *image.RGBA, palette ColorPalette, seed int64) error {
	bounds := img.Bounds()
	rand.Seed(seed)

	// Choose two colors from palette
	color1 := palette.RandomColor()
	color2 := palette.RandomColor()

	// Create vertical gradient
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		ratio := float64(y-bounds.Min.Y) / float64(bounds.Max.Y-bounds.Min.Y)

		r := uint8(float64(color1.R)*(1-ratio) + float64(color2.R)*ratio)
		g := uint8(float64(color1.G)*(1-ratio) + float64(color2.G)*ratio)
		b := uint8(float64(color1.B)*(1-ratio) + float64(color2.B)*ratio)
		a := uint8(float64(color1.A)*(1-ratio) + float64(color2.A)*ratio)

		c := color.RGBA{R: r, G: g, B: b, A: a}

		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, c)
		}
	}

	return nil
}

func (p *GradientPattern) GetName() string {
	return "gradient"
}

// GeometricPattern implementation
func (p *GeometricPattern) Generate(img *image.RGBA, palette ColorPalette, seed int64) error {
	bounds := img.Bounds()
	rand.Seed(seed)

	// Fill with base color
	baseColor := palette.WeightedRandomColor()
	draw.Draw(img, bounds, &image.Uniform{baseColor}, image.Point{}, draw.Src)

	// Add geometric shapes
	numShapes := 5 + rand.Intn(10)

	for i := 0; i < numShapes; i++ {
		shapeColor := palette.RandomColor()

		shapeType := rand.Intn(3)
		x := rand.Intn(bounds.Max.X)
		y := rand.Intn(bounds.Max.Y)

		switch shapeType {
		case 0: // Circle
			radius := 20 + rand.Intn(80)
			p.drawCircle(img, x, y, radius, shapeColor)
		case 1: // Rectangle
			width := 40 + rand.Intn(120)
			height := 40 + rand.Intn(120)
			p.drawRectangle(img, x, y, width, height, shapeColor)
		case 2: // Triangle
			size := 30 + rand.Intn(70)
			p.drawTriangle(img, x, y, size, shapeColor)
		}
	}

	return nil
}

func (p *GeometricPattern) drawCircle(img *image.RGBA, x, y, radius int, c color.RGBA) {
	bounds := img.Bounds()
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				px, py := x+dx, y+dy
				if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
					img.Set(px, py, c)
				}
			}
		}
	}
}

func (p *GeometricPattern) drawRectangle(img *image.RGBA, x, y, width, height int, c color.RGBA) {
	bounds := img.Bounds()
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			px, py := x+dx, y+dy
			if px >= bounds.Min.X && px < bounds.Max.X && py >= bounds.Min.Y && py < bounds.Max.Y {
				img.Set(px, py, c)
			}
		}
	}
}

func (p *GeometricPattern) drawTriangle(img *image.RGBA, x, y, size int, c color.RGBA) {
	for i := 0; i < size; i++ {
		for j := 0; j <= i; j++ {
			px := x - i/2 + j
			py := y + i
			if px >= 0 && py >= 0 && px < img.Bounds().Max.X && py < img.Bounds().Max.Y {
				img.Set(px, py, c)
			}
		}
	}
}

func (p *GeometricPattern) GetName() string {
	return "geometric"
}

// NoisePattern implementation
func (p *NoisePattern) Generate(img *image.RGBA, palette ColorPalette, seed int64) error {
	bounds := img.Bounds()
	rand.Seed(seed)

	baseColor := palette.WeightedRandomColor()

	// Add subtle noise texture
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Base color with slight variation
			variation := rand.Intn(30) - 15 // -15 to +15

			r := uint8(clamp(int(baseColor.R)+variation, 0, 255))
			g := uint8(clamp(int(baseColor.G)+variation, 0, 255))
			b := uint8(clamp(int(baseColor.B)+variation, 0, 255))
			a := baseColor.A

			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	return nil
}

func (p *NoisePattern) GetName() string {
	return "noise"
}

// Helper methods for ColorPalette
func (p ColorPalette) RandomColor() color.RGBA {
	return p.Colors[rand.Intn(len(p.Colors))]
}

func (p ColorPalette) WeightedRandomColor() color.RGBA {
	r := rand.Float64()
	cumsum := 0.0

	for i, weight := range p.Weights {
		cumsum += weight
		if r <= cumsum {
			return p.Colors[i]
		}
	}

	return p.Colors[len(p.Colors)-1]
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
