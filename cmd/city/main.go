package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alainairom/github-contribution-city/pkg/city"
	"github.com/alainairom/github-contribution-city/pkg/config"
	"github.com/alainairom/github-contribution-city/pkg/github"
	"github.com/alainairom/github-contribution-city/pkg/renderer"
)

func main() {
	var (
		serverMode = flag.Bool("server", false, "Start the web visualization server")
		username   = flag.String("user", "", "GitHub username to visualize (overrides config)")
		token      = flag.String("token", "", "GitHub Personal Access Token (PAT)")
		weeks      = flag.Int("weeks", 0, "Number of weeks to render (4 - 53)")
		maxHeight  = flag.Int("height", 0, "Maximum building height (3 - 30)")
		theme      = flag.String("theme", "", "Visual theme (github-dark, github-light, cyberpunk, mono, neon, matrix)")
		format     = flag.String("format", "svg", "Export format: svg, iso, html, json, all")
		outFile    = flag.String("output", "", "Output filepath or directory")
		batchInput = flag.String("batch", "", "Process all user files recursively from input directory")
		port       = flag.Int("port", 0, "HTTP server port (default 8088)")
		envFile    = flag.String("env", ".env", "Path to .env configuration file")
	)
	flag.Parse()

	cfg := config.LoadConfig(*envFile)

	// Apply CLI overrides
	if *username != "" {
		cfg.GitHubUsername = *username
	}
	if *token != "" {
		cfg.GitHubToken = *token
	}
	if *weeks > 0 {
		cfg.Weeks = *weeks
	}
	if *maxHeight > 0 {
		cfg.MaxHeight = *maxHeight
	}
	if *theme != "" {
		cfg.Theme = *theme
	}
	if *port > 0 {
		cfg.Port = *port
	}

	// Ensure output directory exists
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		log.Fatalf("failed to create output dir: %v", err)
	}

	client := github.NewClient(cfg.GitHubToken)
	builder := city.NewBuilder()
	svgRenderer := renderer.NewSVGRenderer()
	htmlRenderer := renderer.NewHTMLRenderer()

	// If batch input flag provided or requested
	if *batchInput != "" {
		runBatchProcessing(cfg, client, builder, svgRenderer, htmlRenderer, *batchInput)
		return
	}

	// If Server Mode
	if *serverMode {
		runServer(cfg, client, builder, svgRenderer, htmlRenderer)
		return
	}

	// CLI Mode: Generate for user
	if cfg.GitHubUsername == "" {
		fmt.Println("Usage: github-contribution-city -user <username> [options]")
		fmt.Println("       github-contribution-city -server")
		flag.PrintDefaults()
		os.Exit(1)
	}

	fmt.Printf("🏙️  Generating Contribution City for '%s' (Weeks: %d, MaxHeight: %d, Theme: %s)...\n",
		cfg.GitHubUsername, cfg.Weeks, cfg.MaxHeight, cfg.Theme)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	cal, err := client.FetchCalendar(ctx, cfg.GitHubUsername)
	if err != nil {
		log.Fatalf("❌ Error fetching contribution data: %v", err)
	}

	opts := city.Options{
		Weeks:     cfg.Weeks,
		MaxHeight: cfg.MaxHeight,
		Theme:     cfg.Theme,
	}

	scene := builder.Build(cal, opts)

	timestamp := time.Now().Format("20060102_150405")
	baseOutName := fmt.Sprintf("%s_city_%s", cfg.GitHubUsername, timestamp)

	if *outFile != "" {
		baseOutName = strings.TrimSuffix(*outFile, filepath.Ext(*outFile))
	} else {
		baseOutName = filepath.Join(cfg.OutputDir, baseOutName)
	}

	fmt.Printf("✅ Fetched %d contributions. Busiest week: %d contributions (Week %d).\n",
		scene.TotalContributions, scene.MaxWeeklyTotal, scene.BusiestWeekIndex+1)

	switch strings.ToLower(*format) {
	case "svg":
		outPath := baseOutName + ".svg"
		svgContent := svgRenderer.RenderSkyline(scene, cfg.Theme)
		if err := os.WriteFile(outPath, []byte(svgContent), 0644); err != nil {
			log.Fatalf("failed to write svg: %v", err)
		}
		fmt.Printf("✨ SVG Skyline written to: %s\n", outPath)

	case "iso":
		outPath := baseOutName + "_3d.svg"
		isoContent := svgRenderer.RenderIsometric3D(scene, cfg.Theme)
		if err := os.WriteFile(outPath, []byte(isoContent), 0644); err != nil {
			log.Fatalf("failed to write iso svg: %v", err)
		}
		fmt.Printf("✨ 3D Isometric SVG written to: %s\n", outPath)

	case "html":
		outPath := baseOutName + ".html"
		htmlContent := htmlRenderer.RenderInteractive3D(scene, cfg.Theme)
		if err := os.WriteFile(outPath, []byte(htmlContent), 0644); err != nil {
			log.Fatalf("failed to write html: %v", err)
		}
		fmt.Printf("✨ 3D Interactive HTML written to: %s\n", outPath)

	case "json":
		outPath := baseOutName + ".json"
		jsonData, _ := json.MarshalIndent(scene, "", "  ")
		if err := os.WriteFile(outPath, jsonData, 0644); err != nil {
			log.Fatalf("failed to write json: %v", err)
		}
		fmt.Printf("✨ Scene JSON data written to: %s\n", outPath)

	case "all":
		svgPath := baseOutName + ".svg"
		isoPath := baseOutName + "_3d.svg"
		htmlPath := baseOutName + ".html"
		jsonPath := baseOutName + ".json"

		_ = os.WriteFile(svgPath, []byte(svgRenderer.RenderSkyline(scene, cfg.Theme)), 0644)
		_ = os.WriteFile(isoPath, []byte(svgRenderer.RenderIsometric3D(scene, cfg.Theme)), 0644)
		_ = os.WriteFile(htmlPath, []byte(htmlRenderer.RenderInteractive3D(scene, cfg.Theme)), 0644)
		jsonData, _ := json.MarshalIndent(scene, "", "  ")
		_ = os.WriteFile(jsonPath, jsonData, 0644)

		fmt.Println("✨ Generated all formats:")
		fmt.Printf("   • SVG:  %s\n", svgPath)
		fmt.Printf("   • ISO:  %s\n", isoPath)
		fmt.Printf("   • HTML: %s\n", htmlPath)
		fmt.Printf("   • JSON: %s\n", jsonPath)

	default:
		log.Fatalf("unknown format '%s'", *format)
	}
}

func runBatchProcessing(
	cfg *config.Config,
	client *github.Client,
	builder *city.Builder,
	svgR *renderer.SVGRenderer,
	htmlR *renderer.HTMLRenderer,
	inputDir string,
) {
	fmt.Printf("📁 Running batch processing on input directory: %s\n", inputDir)
	_ = os.MkdirAll(cfg.OutputDir, 0755)

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// If it's a JSON or text file containing usernames or cached calendars
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".txt" || ext == ".json" {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}

			// Try parsing as lines of usernames
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				user := strings.TrimSpace(line)
				if user == "" || strings.HasPrefix(user, "#") {
					continue
				}

				fmt.Printf("🔄 Batch processing user: %s...\n", user)
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				cal, fErr := client.FetchCalendar(ctx, user)
				cancel()
				if fErr != nil {
					fmt.Printf("⚠️  Skipping %s (error: %v)\n", user, fErr)
					continue
				}

				scene := builder.Build(cal, city.Options{
					Weeks:     cfg.Weeks,
					MaxHeight: cfg.MaxHeight,
					Theme:     cfg.Theme,
				})

				timestamp := time.Now().Format("20060102_150405")
				svgFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_skyline_%s.svg", user, timestamp))
				isoFile := filepath.Join(cfg.OutputDir, fmt.Sprintf("%s_iso3d_%s.svg", user, timestamp))
				_ = os.WriteFile(svgFile, []byte(svgR.RenderSkyline(scene, cfg.Theme)), 0644)
				_ = os.WriteFile(isoFile, []byte(svgR.RenderIsometric3D(scene, cfg.Theme)), 0644)
				fmt.Printf("   -> Generated %s and %s\n", svgFile, isoFile)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("batch processing error: %v", err)
	} else {
		fmt.Println("✅ Batch processing finished.")
	}
}

func runServer(
	cfg *config.Config,
	client *github.Client,
	builder *city.Builder,
	svgR *renderer.SVGRenderer,
	htmlR *renderer.HTMLRenderer,
) {
	mux := http.NewServeMux()

	// Landing Page & Interactive 3D Viewer
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		user := r.URL.Query().Get("user")
		if user == "" {
			user = cfg.GitHubUsername
		}
		theme := r.URL.Query().Get("theme")
		if theme == "" {
			theme = cfg.Theme
		}
		weeks := cfg.Weeks
		if wStr := r.URL.Query().Get("weeks"); wStr != "" {
			if wVal, err := strconv.Atoi(wStr); err == nil && wVal >= 4 && wVal <= 53 {
				weeks = wVal
			}
		}
		height := cfg.MaxHeight
		if hStr := r.URL.Query().Get("height"); hStr != "" {
			if hVal, err := strconv.Atoi(hStr); err == nil && hVal >= 3 && hVal <= 30 {
				height = hVal
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		cal, err := client.FetchCalendar(ctx, user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching contribution calendar for '%s': %v", user, err), http.StatusBadGateway)
			return
		}

		scene := builder.Build(cal, city.Options{
			Weeks:     weeks,
			MaxHeight: height,
			Theme:     theme,
		})

		htmlDoc := htmlR.RenderInteractive3D(scene, theme)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(htmlDoc))
	})

	// SVG Skyline Endpoint
	mux.HandleFunc("/api/svg", func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		if user == "" {
			user = cfg.GitHubUsername
		}
		theme := r.URL.Query().Get("theme")
		if theme == "" {
			theme = cfg.Theme
		}
		weeks := cfg.Weeks
		if wStr := r.URL.Query().Get("weeks"); wStr != "" {
			if wVal, err := strconv.Atoi(wStr); err == nil {
				weeks = wVal
			}
		}
		height := cfg.MaxHeight
		if hStr := r.URL.Query().Get("height"); hStr != "" {
			if hVal, err := strconv.Atoi(hStr); err == nil {
				height = hVal
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		cal, err := client.FetchCalendar(ctx, user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching calendar: %v", err), http.StatusInternalServerError)
			return
		}

		scene := builder.Build(cal, city.Options{
			Weeks:     weeks,
			MaxHeight: height,
			Theme:     theme,
		})

		svgContent := svgR.RenderSkyline(scene, theme)
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(svgContent))
	})

	// 3D Isometric SVG Endpoint
	mux.HandleFunc("/api/iso", func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		if user == "" {
			user = cfg.GitHubUsername
		}
		theme := r.URL.Query().Get("theme")
		if theme == "" {
			theme = cfg.Theme
		}
		weeks := cfg.Weeks
		if wStr := r.URL.Query().Get("weeks"); wStr != "" {
			if wVal, err := strconv.Atoi(wStr); err == nil {
				weeks = wVal
			}
		}
		height := cfg.MaxHeight
		if hStr := r.URL.Query().Get("height"); hStr != "" {
			if hVal, err := strconv.Atoi(hStr); err == nil {
				height = hVal
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		cal, err := client.FetchCalendar(ctx, user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching calendar: %v", err), http.StatusInternalServerError)
			return
		}

		scene := builder.Build(cal, city.Options{
			Weeks:     weeks,
			MaxHeight: height,
			Theme:     theme,
		})

		isoContent := svgR.RenderIsometric3D(scene, theme)
		w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(isoContent))
	})

	// JSON City Model Endpoint
	mux.HandleFunc("/api/city", func(w http.ResponseWriter, r *http.Request) {
		user := r.URL.Query().Get("user")
		if user == "" {
			user = cfg.GitHubUsername
		}
		theme := r.URL.Query().Get("theme")
		if theme == "" {
			theme = cfg.Theme
		}

		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		cal, err := client.FetchCalendar(ctx, user)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error fetching calendar: %v", err), http.StatusInternalServerError)
			return
		}

		scene := builder.Build(cal, city.Options{
			Weeks:     cfg.Weeks,
			MaxHeight: cfg.MaxHeight,
			Theme:     theme,
		})

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(scene)
	})

	// Health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
	})

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	serverURL := fmt.Sprintf("http://localhost:%d", cfg.Port)

	fmt.Printf("\n🚀 GitHub Contribution City Web Server is starting...\n")
	fmt.Printf("🌐 Application URL: %s\n", serverURL)
	fmt.Printf("   • Interactive 3D View: %s/?user=%s&theme=%s\n", serverURL, cfg.GitHubUsername, cfg.Theme)
	fmt.Printf("   • SVG Skyline:        %s/api/svg?user=%s&theme=%s\n", serverURL, cfg.GitHubUsername, cfg.Theme)
	fmt.Printf("   • 3D Isometric SVG:   %s/api/iso?user=%s&theme=%s\n", serverURL, cfg.GitHubUsername, cfg.Theme)
	fmt.Printf("   • JSON API Model:     %s/api/city?user=%s\n\n", serverURL, cfg.GitHubUsername)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
