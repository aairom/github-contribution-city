package renderer

import (
	"strings"
	"testing"

	"github.com/alainairom/github-contribution-city/pkg/city"
)

func createTestScene() *city.CityScene {
	grid := make([][]city.Cell, 10)
	for i := range grid {
		grid[i] = make([]city.Cell, 8)
		for j := range grid[i] {
			grid[i][j] = city.Cell{
				Char:      '#',
				Kind:      city.CellWindow,
				Level:     2,
				Count:     5,
				WeekIndex: j / 2,
				DayIndex:  j % 2,
				Date:      "2024-01-01",
			}
		}
	}

	return &city.CityScene{
		Username:           "testdev",
		TotalContributions: 250,
		MaxWeeklyTotal:     40,
		BusiestWeekIndex:   1,
		GridWidth:          8,
		GridHeight:         10,
		Grid:               grid,
		Buildings3D: []city.Building3D{
			{
				WeekIndex:   0,
				FirstDay:    "2024-01-01",
				WeeklyTotal: 20,
				Floors:      5,
				HasAntenna:  false,
				District:    "Q1 Historical Quarter",
			},
			{
				WeekIndex:   1,
				FirstDay:    "2024-01-08",
				WeeklyTotal: 40,
				Floors:      10,
				HasAntenna:  true,
				District:    "Q1 Historical Quarter",
			},
		},
		MonthLabels: []city.MonthLabel{
			{MonthName: "Jan", Column: 0},
		},
	}
}

func TestRenderSkylineSVG(t *testing.T) {
	scene := createTestScene()
	renderer := NewSVGRenderer()

	svg := renderer.RenderSkyline(scene, "github-dark")
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "</svg>") {
		t.Errorf("expected valid SVG document")
	}
	if !strings.Contains(svg, "testdev's Contribution City") {
		t.Errorf("expected username in SVG title")
	}
	if !strings.Contains(svg, "250 total contributions") {
		t.Errorf("expected total contributions in subtitle")
	}
}

func TestRenderIsometric3DSVG(t *testing.T) {
	scene := createTestScene()
	renderer := NewSVGRenderer()

	svg := renderer.RenderIsometric3D(scene, "cyberpunk")
	if !strings.Contains(svg, "<svg") || !strings.Contains(svg, "</svg>") {
		t.Errorf("expected valid SVG document")
	}
	if !strings.Contains(svg, "3D Contribution City") {
		t.Errorf("expected 3D title")
	}
	if !strings.Contains(svg, "CYBERPUNK") {
		t.Errorf("expected theme name in subtitle")
	}
	if !strings.Contains(svg, "polygon") {
		t.Errorf("expected isometric polygons")
	}
}

func TestRenderInteractiveHTML(t *testing.T) {
	scene := createTestScene()
	renderer := NewHTMLRenderer()

	html := renderer.RenderInteractive3D(scene, "github-dark")
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Errorf("expected valid HTML doctype")
	}
	if !strings.Contains(html, "Three.js") {
		t.Errorf("expected Three.js inclusion")
	}
	if !strings.Contains(html, "testdev") {
		t.Errorf("expected testdev in HTML")
	}
}

func TestThemes(t *testing.T) {
	tDark := GetTheme("github-dark")
	if tDark.Name != "github-dark" {
		t.Errorf("expected github-dark, got %s", tDark.Name)
	}

	tUnknown := GetTheme("non-existent-theme")
	if tUnknown.Name != "github-dark" {
		t.Errorf("expected fallback to github-dark, got %s", tUnknown.Name)
	}
}
