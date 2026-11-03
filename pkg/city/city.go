package city

import (
	"math"
	"time"

	"github.com/alainairom/github-contribution-city/pkg/github"
)

// CellKind represents the structural type of a 2D grid cell
type CellKind string

const (
	CellSky      CellKind = "SKY"
	CellRoof     CellKind = "ROOF"
	CellAntenna  CellKind = "ANTENNA"
	CellWindow   CellKind = "WINDOW"
	CellWall     CellKind = "WALL"
	CellDoor     CellKind = "DOOR"
	CellStreet   CellKind = "STREET"
	CellTree     CellKind = "TREE"
	CellCar      CellKind = "CAR"
	CellCloud    CellKind = "CLOUD"
	CellStar     CellKind = "STAR"
	CellLandmark CellKind = "LANDMARK"
)

// Cell is a single coordinate in the 2D visual grid
type Cell struct {
	Char        rune     `json:"char"`
	Kind        CellKind `json:"kind"`
	Level       int      `json:"level"`      // 0-4 contribution intensity level
	Count       int      `json:"count"`      // actual contribution count
	WeekIndex   int      `json:"week_index"` // -1 if none
	DayIndex    int      `json:"day_index"`  // 0-6 (Sun-Sat), -1 if none
	Date        string   `json:"date"`       // YYYY-MM-DD
	TooltipText string   `json:"tooltip"`
}

// Building3D represents a 3D building block in the isometric / WebGL world
type Building3D struct {
	WeekIndex    int                  `json:"week_index"`
	FirstDay     string               `json:"first_day"`
	WeeklyTotal  int                  `json:"weekly_total"`
	GridX        float64              `json:"grid_x"` // Position along street (e.g. 0 to 52)
	GridZ        float64              `json:"grid_z"` // Depth offset (0 for main street, +/- for district)
	Width        float64              `json:"width"`
	Depth        float64              `json:"depth"`
	Height       float64              `json:"height"`
	Floors       int                  `json:"floors"`
	HasAntenna   bool                 `json:"has_antenna"`
	Days         []BuildingDayDetails `json:"days"`
	District     string               `json:"district"`
	ColorPrimary string               `json:"color_primary"`
	ColorWindows []string             `json:"color_windows"`
}

// BuildingDayDetails contains daily metrics per floor/facet
type BuildingDayDetails struct {
	Date     string `json:"date"`
	Weekday  int    `json:"weekday"`
	Count    int    `json:"count"`
	Level    int    `json:"level"`
	LevelStr string `json:"level_str"`
}

// Options configure city construction
type Options struct {
	Weeks     int    `json:"weeks"`      // e.g. 53
	MaxHeight int    `json:"max_height"` // e.g. 14
	Theme     string `json:"theme"`
}

// CityScene is the complete processed representation of the city
type CityScene struct {
	Username           string       `json:"username"`
	TotalContributions int          `json:"total_contributions"`
	MaxWeeklyTotal     int          `json:"max_weekly_total"`
	BusiestWeekIndex   int          `json:"busiest_week_index"`
	BusiestWeekDate    string       `json:"busiest_week_date"`
	AverageWeekly      float64      `json:"average_weekly"`
	GridWidth          int          `json:"grid_width"`
	GridHeight         int          `json:"grid_height"`
	Grid               [][]Cell     `json:"grid"`
	Buildings3D        []Building3D `json:"buildings_3d"`
	MonthLabels        []MonthLabel `json:"month_labels"`
	GeneratedAt        time.Time    `json:"generated_at"`
}

// MonthLabel marks where a month begins across columns
type MonthLabel struct {
	MonthName string `json:"month_name"`
	Column    int    `json:"column"`
}

// Builder processes GitHub calendar data into City scenes
type Builder struct{}

// NewBuilder creates a new city Builder
func NewBuilder() *Builder {
	return &Builder{}
}

// Build transforms a ContributionCalendar into CityScene (2D Grid + 3D Isometric models)
func (b *Builder) Build(cal *github.ContributionCalendar, opts Options) *CityScene {
	if opts.Weeks < 4 {
		opts.Weeks = 4
	}
	if opts.Weeks > 53 {
		opts.Weeks = 53
	}
	if opts.MaxHeight < 3 {
		opts.MaxHeight = 3
	}
	if opts.MaxHeight > 40 {
		opts.MaxHeight = 40
	}

	weeks := cal.Weeks
	if len(weeks) > opts.Weeks {
		weeks = weeks[len(weeks)-opts.Weeks:]
	}

	numWeeks := len(weeks)
	weeklyTotals := make([]int, numWeeks)
	maxWeekly := 0
	busiestIdx := -1
	totalSum := 0

	for i, w := range weeks {
		wSum := 0
		for _, d := range w.ContributionDays {
			wSum += d.ContributionCount
		}
		weeklyTotals[i] = wSum
		totalSum += wSum
		if wSum >= maxWeekly {
			maxWeekly = wSum
			busiestIdx = i
		}
	}

	// Calculate building heights using logarithmic scale as specified in Quarkus Contribution City:
	// height = weeklyTotal == 0 ? 1 : 1 + round( log1p(weeklyTotal) / log1p(maxWeeklyTotal) * (maxHeight - 1) )
	buildingHeights := make([]int, numWeeks)
	for i, tot := range weeklyTotals {
		if tot == 0 || maxWeekly == 0 {
			buildingHeights[i] = 1
		} else {
			ratio := math.Log1p(float64(tot)) / math.Log1p(float64(maxWeekly))
			h := 1 + int(math.Round(ratio*float64(opts.MaxHeight-1)))
			if h < 1 {
				h = 1
			}
			if h > opts.MaxHeight {
				h = opts.MaxHeight
			}
			buildingHeights[i] = h
		}
	}

	// 2D Grid Setup
	// Each building is 2 character cells wide + 1 cell space or compact (2 cells per week)
	// Additional top rows for Sky, Clouds, Antenna (2 rows), Roof (1 row), MaxHeight floors, Street (1 row), Month label (1 row)
	topPadding := 3                               // for antenna and sky
	gridHeight := topPadding + opts.MaxHeight + 2 // + street + month labels
	gridWidth := numWeeks * 2

	grid := make([][]Cell, gridHeight)
	for r := 0; r < gridHeight; r++ {
		grid[r] = make([]Cell, gridWidth)
		for c := 0; c < gridWidth; c++ {
			grid[r][c] = Cell{
				Char:      ' ',
				Kind:      CellSky,
				Level:     0,
				WeekIndex: -1,
				DayIndex:  -1,
			}
		}
	}

	// Populate Buildings in 2D Grid
	streetRow := topPadding + opts.MaxHeight

	// Fill Street
	for c := 0; c < gridWidth; c++ {
		grid[streetRow][c] = Cell{
			Char:      '=',
			Kind:      CellStreet,
			Level:     0,
			WeekIndex: -1,
			DayIndex:  -1,
		}
	}

	var monthLabels []MonthLabel
	lastMonth := ""

	// Process 3D structures and 2D cells
	buildings3D := make([]Building3D, numWeeks)

	for wIdx, w := range weeks {
		colStart := wIdx * 2
		bHeight := buildingHeights[wIdx]
		wTotal := weeklyTotals[wIdx]
		hasAntenna := (wIdx == busiestIdx && maxWeekly > 0)

		// Days mapping (Sunday=0 to Saturday=6)
		dayMap := make([]github.ContributionDay, 7)
		for _, d := range w.ContributionDays {
			if d.Weekday >= 0 && d.Weekday < 7 {
				dayMap[d.Weekday] = d
			}
		}

		// Check month label
		if len(w.ContributionDays) > 0 {
			firstDate := w.ContributionDays[0].Date
			if t, err := time.Parse("2006-01-02", firstDate); err == nil {
				monthName := t.Format("Jan")
				if monthName != lastMonth {
					monthLabels = append(monthLabels, MonthLabel{
						MonthName: monthName,
						Column:    colStart,
					})
					lastMonth = monthName
				}
			}
		}

		// Draw Antenna if busiest week
		if hasAntenna {
			antennaRowTop := streetRow - bHeight - 2
			antennaRowMid := streetRow - bHeight - 1
			if antennaRowTop >= 0 {
				grid[antennaRowTop][colStart] = Cell{
					Char:        '|',
					Kind:        CellAntenna,
					Level:       4,
					WeekIndex:   wIdx,
					TooltipText: "Busiest Week Antenna",
				}
			}
			if antennaRowMid >= 0 {
				grid[antennaRowMid][colStart] = Cell{
					Char:        '+',
					Kind:        CellAntenna,
					Level:       4,
					WeekIndex:   wIdx,
					TooltipText: "Busiest Week Antenna",
				}
			}
		}

		// Draw Roof
		roofRow := streetRow - bHeight
		if roofRow >= 0 {
			grid[roofRow][colStart] = Cell{
				Char:      '[',
				Kind:      CellRoof,
				Level:     0,
				WeekIndex: wIdx,
			}
			grid[roofRow][colStart+1] = Cell{
				Char:      ']',
				Kind:      CellRoof,
				Level:     0,
				WeekIndex: wIdx,
			}
		}

		// Draw Floors / Windows
		// Cycle through the days of the week (Sunday to Saturday) across rows
		dayCycleIdx := 0
		dayDetails := make([]BuildingDayDetails, 0, len(w.ContributionDays))
		for _, d := range w.ContributionDays {
			lvl := levelToInt(d.ContributionLevel)
			dayDetails = append(dayDetails, BuildingDayDetails{
				Date:     d.Date,
				Weekday:  d.Weekday,
				Count:    d.ContributionCount,
				Level:    lvl,
				LevelStr: d.ContributionLevel,
			})
		}

		for floor := 1; floor < bHeight; floor++ {
			floorRow := streetRow - floor
			if floorRow < 0 {
				continue
			}

			// Left window
			dLeft := dayMap[dayCycleIdx%7]
			lvlLeft := levelToInt(dLeft.ContributionLevel)
			grid[floorRow][colStart] = Cell{
				Char:        '#',
				Kind:        CellWindow,
				Level:       lvlLeft,
				Count:       dLeft.ContributionCount,
				WeekIndex:   wIdx,
				DayIndex:    dLeft.Weekday,
				Date:        dLeft.Date,
				TooltipText: dLeft.Date,
			}
			dayCycleIdx++

			// Right window
			dRight := dayMap[dayCycleIdx%7]
			lvlRight := levelToInt(dRight.ContributionLevel)
			grid[floorRow][colStart+1] = Cell{
				Char:        '#',
				Kind:        CellWindow,
				Level:       lvlRight,
				Count:       dRight.ContributionCount,
				WeekIndex:   wIdx,
				DayIndex:    dRight.Weekday,
				Date:        dRight.Date,
				TooltipText: dRight.Date,
			}
			dayCycleIdx++
		}

		// 3D Building Object
		buildings3D[wIdx] = Building3D{
			WeekIndex:   wIdx,
			FirstDay:    w.FirstDay,
			WeeklyTotal: wTotal,
			GridX:       float64(wIdx) * 1.8,
			GridZ:       float64(wIdx%3-1) * 0.4, // subtle stagger for city depth
			Width:       1.4,
			Depth:       1.4,
			Height:      float64(bHeight) * 0.75,
			Floors:      bHeight,
			HasAntenna:  hasAntenna,
			Days:        dayDetails,
			District:    getDistrict(wIdx, numWeeks),
		}
	}

	busiestDate := ""
	if busiestIdx >= 0 && busiestIdx < len(weeks) {
		busiestDate = weeks[busiestIdx].FirstDay
	}

	avgWeekly := 0.0
	if numWeeks > 0 {
		avgWeekly = float64(totalSum) / float64(numWeeks)
	}

	return &CityScene{
		Username:           cal.Username,
		TotalContributions: cal.TotalContributions,
		MaxWeeklyTotal:     maxWeekly,
		BusiestWeekIndex:   busiestIdx,
		BusiestWeekDate:    busiestDate,
		AverageWeekly:      avgWeekly,
		GridWidth:          gridWidth,
		GridHeight:         gridHeight,
		Grid:               grid,
		Buildings3D:        buildings3D,
		MonthLabels:        monthLabels,
		GeneratedAt:        time.Now().UTC(),
	}
}

func levelToInt(level string) int {
	switch level {
	case "FOURTH_QUARTILE":
		return 4
	case "THIRD_QUARTILE":
		return 3
	case "SECOND_QUARTILE":
		return 2
	case "FIRST_QUARTILE":
		return 1
	default:
		return 0
	}
}

func getDistrict(wIdx, totalWeeks int) string {
	quarter := totalWeeks / 4
	if quarter == 0 {
		quarter = 1
	}
	switch {
	case wIdx < quarter:
		return "Q1 Historical Quarter"
	case wIdx < quarter*2:
		return "Q2 Midtown District"
	case wIdx < quarter*3:
		return "Q3 Tech Promenade"
	default:
		return "Q4 Innovation Skyline"
	}
}
