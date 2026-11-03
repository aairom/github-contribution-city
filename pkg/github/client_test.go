package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const sampleGraphQLResponse = `{
  "data": {
    "user": {
      "contributionsCollection": {
        "contributionCalendar": {
          "totalContributions": 42,
          "colors": ["#ebedf0", "#9be9a8", "#40c463", "#30a14e", "#216e39"],
          "weeks": [
            {
              "firstDay": "2024-01-01",
              "contributionDays": [
                {
                  "contributionCount": 5,
                  "contributionLevel": "FIRST_QUARTILE",
                  "date": "2024-01-01",
                  "weekday": 1
                },
                {
                  "contributionCount": 10,
                  "contributionLevel": "SECOND_QUARTILE",
                  "date": "2024-01-02",
                  "weekday": 2
                }
              ]
            }
          ]
        }
      }
    }
  }
}`

const sampleHTML = `
<div>
  <h2>42 contributions in the last year</h2>
  <table>
    <tbody>
      <tr>
        <td data-date="2024-01-01" data-level="1">1 contribution on January 1, 2024</td>
        <td data-date="2024-01-02" data-level="2">2 contributions on January 2, 2024</td>
        <td data-date="2024-01-07" data-level="4">8 contributions on January 7, 2024</td>
      </tr>
    </tbody>
  </table>
</div>
`

func TestFetchGraphQL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(sampleGraphQLResponse))
	}))
	defer ts.Close()

	client := NewClient("dummy-token", WithGraphQLURL(ts.URL))
	cal, err := client.FetchCalendar(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cal.TotalContributions != 42 {
		t.Errorf("expected 42 total contributions, got %d", cal.TotalContributions)
	}
	if len(cal.Weeks) != 1 {
		t.Errorf("expected 1 week, got %d", len(cal.Weeks))
	}
	if len(cal.Weeks[0].ContributionDays) != 2 {
		t.Errorf("expected 2 days, got %d", len(cal.Weeks[0].ContributionDays))
	}
	if cal.Weeks[0].ContributionDays[0].ContributionCount != 5 {
		t.Errorf("expected 5 count, got %d", cal.Weeks[0].ContributionDays[0].ContributionCount)
	}
}

func TestParseContributionHTML(t *testing.T) {
	cal, err := ParseContributionHTML(sampleHTML, "testuser")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if cal.TotalContributions != 42 {
		t.Errorf("expected 42, got %d", cal.TotalContributions)
	}
	if len(cal.Weeks) == 0 {
		t.Errorf("expected non-empty weeks")
	}
}
