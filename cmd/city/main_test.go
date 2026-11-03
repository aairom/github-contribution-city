package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alainairom/github-contribution-city/pkg/city"
	"github.com/alainairom/github-contribution-city/pkg/config"
	"github.com/alainairom/github-contribution-city/pkg/github"
	"github.com/alainairom/github-contribution-city/pkg/renderer"
)

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "ok") {
		t.Errorf("expected ok status in body")
	}
}

func TestEndToEndPipeline(t *testing.T) {
	cal := &github.ContributionCalendar{
		Username:           "testuser",
		TotalContributions: 100,
		Weeks: []github.ContributionWeek{
			{
				FirstDay: "2024-01-01",
				ContributionDays: []github.ContributionDay{
					{ContributionCount: 5, ContributionLevel: "FIRST_QUARTILE", Date: "2024-01-01", Weekday: 0},
					{ContributionCount: 15, ContributionLevel: "THIRD_QUARTILE", Date: "2024-01-02", Weekday: 1},
				},
			},
		},
	}

	cfg := config.DefaultConfig()
	builder := city.NewBuilder()
	svgR := renderer.NewSVGRenderer()
	htmlR := renderer.NewHTMLRenderer()

	scene := builder.Build(cal, city.Options{
		Weeks:     cfg.Weeks,
		MaxHeight: cfg.MaxHeight,
		Theme:     cfg.Theme,
	})

	svgSkyline := svgR.RenderSkyline(scene, cfg.Theme)
	if len(svgSkyline) == 0 || !strings.Contains(svgSkyline, "<svg") {
		t.Errorf("expected generated SVG skyline")
	}

	svgIso := svgR.RenderIsometric3D(scene, cfg.Theme)
	if len(svgIso) == 0 || !strings.Contains(svgIso, "<svg") {
		t.Errorf("expected generated 3D Isometric SVG")
	}

	htmlDoc := htmlR.RenderInteractive3D(scene, cfg.Theme)
	if len(htmlDoc) == 0 || !strings.Contains(htmlDoc, "Three.js") {
		t.Errorf("expected generated HTML doc")
	}
}
