package renderer

// Theme defines the visual color palette for SVG and 3D rendering
type Theme struct {
	Name            string   `json:"name"`
	Background      string   `json:"background"`
	BuildingBody    string   `json:"building_body"`
	BuildingRoof    string   `json:"building_roof"`
	BuildingEdge    string   `json:"building_edge"`
	Street          string   `json:"street"`
	StreetLine      string   `json:"street_line"`
	Antenna         string   `json:"antenna"`
	TextPrimary     string   `json:"text_primary"`
	TextSecondary   string   `json:"text_secondary"`
	TextMuted       string   `json:"text_muted"`
	WindowLevels    []string `json:"window_levels"` // 0: None, 1: Low, 2: Med, 3: High, 4: Max
	BuildingTop3D   string   `json:"building_top_3d"`
	BuildingSide3D  string   `json:"building_side_3d"`
	BuildingFront3D string   `json:"building_front_3d"`
}

var Themes = map[string]Theme{
	"github-dark": {
		Name:            "github-dark",
		Background:      "#0d1117",
		BuildingBody:    "#161b22",
		BuildingRoof:    "#30363d",
		BuildingEdge:    "#21262d",
		Street:          "#21262d",
		StreetLine:      "#30363d",
		Antenna:         "#39d353",
		TextPrimary:     "#f0f6fc",
		TextSecondary:   "#8b949e",
		TextMuted:       "#484f58",
		WindowLevels:    []string{"#161b22", "#0e4429", "#006d32", "#26a641", "#39d353"},
		BuildingTop3D:   "#21262d",
		BuildingSide3D:  "#161b22",
		BuildingFront3D: "#1c2128",
	},
	"github-light": {
		Name:            "github-light",
		Background:      "#ffffff",
		BuildingBody:    "#f6f8fa",
		BuildingRoof:    "#d0d7de",
		BuildingEdge:    "#afb8c1",
		Street:          "#eaeef2",
		StreetLine:      "#afb8c1",
		Antenna:         "#216e39",
		TextPrimary:     "#1f2328",
		TextSecondary:   "#656d76",
		TextMuted:       "#8c959f",
		WindowLevels:    []string{"#ebedf0", "#9be9a8", "#40c463", "#30a14e", "#216e39"},
		BuildingTop3D:   "#e1e4e8",
		BuildingSide3D:  "#d0d7de",
		BuildingFront3D: "#f6f8fa",
	},
	"cyberpunk": {
		Name:            "cyberpunk",
		Background:      "#0a0518",
		BuildingBody:    "#140c2b",
		BuildingRoof:    "#2d124d",
		BuildingEdge:    "#ff007f",
		Street:          "#1f113a",
		StreetLine:      "#00f0ff",
		Antenna:         "#ffe600",
		TextPrimary:     "#00f0ff",
		TextSecondary:   "#ff007f",
		TextMuted:       "#7928ca",
		WindowLevels:    []string{"#1c1038", "#7928ca", "#ff007f", "#00f0ff", "#ffe600"},
		BuildingTop3D:   "#2d124d",
		BuildingSide3D:  "#140c2b",
		BuildingFront3D: "#1c1038",
	},
	"neon": {
		Name:            "neon",
		Background:      "#050510",
		BuildingBody:    "#0f1026",
		BuildingRoof:    "#1e214d",
		BuildingEdge:    "#4d53a6",
		Street:          "#131530",
		StreetLine:      "#00e5ff",
		Antenna:         "#ff2a6d",
		TextPrimary:     "#05d9e8",
		TextSecondary:   "#ff2a6d",
		TextMuted:       "#654062",
		WindowLevels:    []string{"#12132a", "#01012b", "#00567a", "#00b2d6", "#05d9e8"},
		BuildingTop3D:   "#1e214d",
		BuildingSide3D:  "#0f1026",
		BuildingFront3D: "#181b3d",
	},
	"mono": {
		Name:            "mono",
		Background:      "#121212",
		BuildingBody:    "#222222",
		BuildingRoof:    "#444444",
		BuildingEdge:    "#333333",
		Street:          "#2c2c2c",
		StreetLine:      "#555555",
		Antenna:         "#ffffff",
		TextPrimary:     "#ffffff",
		TextSecondary:   "#aaaaaa",
		TextMuted:       "#666666",
		WindowLevels:    []string{"#222222", "#444444", "#777777", "#bbbbbb", "#ffffff"},
		BuildingTop3D:   "#383838",
		BuildingSide3D:  "#1e1e1e",
		BuildingFront3D: "#282828",
	},
	"matrix": {
		Name:            "matrix",
		Background:      "#020b02",
		BuildingBody:    "#041c04",
		BuildingRoof:    "#0a3a0a",
		BuildingEdge:    "#0f5e0f",
		Street:          "#072c07",
		StreetLine:      "#00ff41",
		Antenna:         "#a6ff00",
		TextPrimary:     "#00ff41",
		TextSecondary:   "#008f11",
		TextMuted:       "#003b00",
		WindowLevels:    []string{"#041c04", "#003b00", "#008f11", "#00ff41", "#a6ff00"},
		BuildingTop3D:   "#0a3a0a",
		BuildingSide3D:  "#031503",
		BuildingFront3D: "#062a06",
	},
}

// GetTheme returns the theme by name or defaults to github-dark
func GetTheme(name string) Theme {
	if t, exists := Themes[name]; exists {
		return t
	}
	return Themes["github-dark"]
}
