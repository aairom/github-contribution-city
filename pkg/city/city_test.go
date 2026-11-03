package city

import (
	"testing"

	"github.com/alainairom/github-contribution-city/pkg/github"
)

func TestCityBuilder(t *testing.T) {
	cal := &github.ContributionCalendar{
		Username:           "gopher",
		TotalContributions: 150,
		Weeks: []github.ContributionWeek{
			{
				FirstDay: "2024-01-07",
				ContributionDays: []github.ContributionDay{
					{ContributionCount: 1, ContributionLevel: "FIRST_QUARTILE", Date: "2024-01-07", Weekday: 0},
					{ContributionCount: 4, ContributionLevel: "SECOND_QUARTILE", Date: "2024-01-08", Weekday: 1},
				},
			},
			{
				FirstDay: "2024-01-14",
				ContributionDays: []github.ContributionDay{
					{ContributionCount: 10, ContributionLevel: "FOURTH_QUARTILE", Date: "2024-01-14", Weekday: 0},
				},
			},
			{
				FirstDay: "2024-01-21",
				ContributionDays: []github.ContributionDay{
					{ContributionCount: 0, ContributionLevel: "NONE", Date: "2024-01-21", Weekday: 0},
				},
			},
			{
				FirstDay: "2024-01-28",
				ContributionDays: []github.ContributionDay{
					{ContributionCount: 20, ContributionLevel: "FOURTH_QUARTILE", Date: "2024-01-28", Weekday: 0},
				},
			},
		},
	}

	builder := NewBuilder()
	opts := Options{
		Weeks:     4,
		MaxHeight: 14,
		Theme:     "github-dark",
	}

	scene := builder.Build(cal, opts)

	if scene.Username != "gopher" {
		t.Errorf("expected gopher, got %s", scene.Username)
	}
	if scene.TotalContributions != 150 {
		t.Errorf("expected 150, got %d", scene.TotalContributions)
	}
	if scene.MaxWeeklyTotal != 20 {
		t.Errorf("expected max weekly 20, got %d", scene.MaxWeeklyTotal)
	}
	if scene.BusiestWeekIndex != 3 {
		t.Errorf("expected busiest week index 3, got %d", scene.BusiestWeekIndex)
	}
	if len(scene.Buildings3D) != 4 {
		t.Errorf("expected 4 buildings, got %d", len(scene.Buildings3D))
	}

	// Verify busiest week building has antenna
	if !scene.Buildings3D[3].HasAntenna {
		t.Errorf("expected week 3 to have antenna")
	}

	// Verify quiet week has minimal height
	if scene.Buildings3D[2].Floors != 1 {
		t.Errorf("expected quiet week (0 contributions) to have 1 floor, got %d", scene.Buildings3D[2].Floors)
	}
}

func TestLevelToInt(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"NONE", 0},
		{"FIRST_QUARTILE", 1},
		{"SECOND_QUARTILE", 2},
		{"THIRD_QUARTILE", 3},
		{"FOURTH_QUARTILE", 4},
		{"UNKNOWN", 0},
	}

	for _, tt := range tests {
		got := levelToInt(tt.input)
		if got != tt.want {
			t.Errorf("levelToInt(%s) = %d; want %d", tt.input, got, tt.want)
		}
	}
}
