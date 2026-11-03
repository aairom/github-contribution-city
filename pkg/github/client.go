package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ContributionDay represents one day's contributions
type ContributionDay struct {
	ContributionCount int       `json:"contributionCount"`
	ContributionLevel string    `json:"contributionLevel"` // NONE, FIRST_QUARTILE, SECOND_QUARTILE, THIRD_QUARTILE, FOURTH_QUARTILE
	Date              string    `json:"date"`              // YYYY-MM-DD
	Weekday           int       `json:"weekday"`           // 0 = Sunday, 6 = Saturday
	ParsedDate        time.Time `json:"-"`
}

// ContributionWeek represents a week of days (up to 7 days)
type ContributionWeek struct {
	ContributionDays []ContributionDay `json:"contributionDays"`
	FirstDay         string            `json:"firstDay"`
}

// ContributionCalendar is the complete calendar returned by GitHub
type ContributionCalendar struct {
	TotalContributions int                `json:"totalContributions"`
	Weeks              []ContributionWeek `json:"weeks"`
	Colors             []string           `json:"colors"`
	Username           string             `json:"username"`
}

// Client is a GitHub API client that can fetch contribution calendars
type Client struct {
	token      string
	httpClient *http.Client
	graphQLURL string
}

// ClientOption configures the Client
type ClientOption func(*Client)

// WithHTTPClient sets custom http.Client
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithGraphQLURL sets custom GraphQL endpoint (useful for mock testing)
func WithGraphQLURL(url string) ClientOption {
	return func(c *Client) {
		c.graphQLURL = url
	}
}

// NewClient creates a new GitHub client
func NewClient(token string, opts ...ClientOption) *Client {
	c := &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		graphQLURL: "https://api.github.com/graphql",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// FetchCalendar fetches the user's contribution calendar using GraphQL (if token provided) or scraping fallback
func (c *Client) FetchCalendar(ctx context.Context, username string) (*ContributionCalendar, error) {
	if strings.TrimSpace(username) == "" {
		return nil, fmt.Errorf("username cannot be empty")
	}

	if c.token != "" {
		cal, err := c.fetchGraphQL(ctx, username)
		if err == nil && cal != nil {
			return cal, nil
		}
		// If GraphQL fails (e.g., token scope or invalid), fallback to public calendar fetch
	}

	// Fallback to public SVG / HTML contributions endpoint
	return c.fetchPublicCalendar(ctx, username)
}

type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphQLResponse struct {
	Data struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						FirstDay         string `json:"firstDay"`
						ContributionDays []struct {
							ContributionCount int    `json:"contributionCount"`
							ContributionLevel string `json:"contributionLevel"`
							Date              string `json:"date"`
							Weekday           int    `json:"weekday"`
						} `json:"contributionDays"`
					} `json:"weeks"`
					Colors []string `json:"colors"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

const queryContributionCalendar = `
query($username: String!) {
  user(login: $username) {
    contributionsCollection {
      contributionCalendar {
        totalContributions
        colors
        weeks {
          firstDay
          contributionDays {
            contributionCount
            contributionLevel
            date
            weekday
          }
        }
      }
    }
  }
}
`

func (c *Client) fetchGraphQL(ctx context.Context, username string) (*ContributionCalendar, error) {
	reqPayload := graphQLRequest{
		Query: queryContributionCalendar,
		Variables: map[string]interface{}{
			"username": username,
		},
	}

	bodyBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.graphQLURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "github-contribution-city/1.0")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("graphql request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("graphql returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var gqlResp graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode graphql response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", gqlResp.Errors[0].Message)
	}

	rawCal := gqlResp.Data.User.ContributionsCollection.ContributionCalendar
	cal := &ContributionCalendar{
		TotalContributions: rawCal.TotalContributions,
		Colors:             rawCal.Colors,
		Username:           username,
		Weeks:              make([]ContributionWeek, len(rawCal.Weeks)),
	}

	for i, w := range rawCal.Weeks {
		week := ContributionWeek{
			FirstDay:         w.FirstDay,
			ContributionDays: make([]ContributionDay, len(w.ContributionDays)),
		}
		for j, d := range w.ContributionDays {
			parsed, _ := time.Parse("2006-01-02", d.Date)
			week.ContributionDays[j] = ContributionDay{
				ContributionCount: d.ContributionCount,
				ContributionLevel: d.ContributionLevel,
				Date:              d.Date,
				Weekday:           d.Weekday,
				ParsedDate:        parsed,
			}
		}
		cal.Weeks[i] = week
	}

	return cal, nil
}

// fetchPublicCalendar fetches public contribution data from GitHub profile html/svg endpoint
func (c *Client) fetchPublicCalendar(ctx context.Context, username string) (*ContributionCalendar, error) {
	url := fmt.Sprintf("https://github.com/users/%s/contributions", username)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; github-contribution-city/1.0)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch public calendar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("public contributions endpoint returned status %d for user %s", resp.StatusCode, username)
	}

	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return ParseContributionHTML(string(htmlContent), username)
}

// ParseContributionHTML parses GitHub's public contribution table HTML
func ParseContributionHTML(html string, username string) (*ContributionCalendar, error) {
	cal := &ContributionCalendar{
		Username: username,
		Weeks:    []ContributionWeek{},
	}

	// Regex for <td> elements containing data-date, data-level or tooltips
	// Example: <td tabindex="0" data-ix="0" ... data-date="2024-01-01" data-level="1" id="contribution-day-component-0-0" class="ContributionCalendar-day">...</td>
	tdRegex := regexp.MustCompile(`(?s)<td[^>]*?data-date="(\d{4}-\d{2}-\d{2})"[^>]*?>.*?</td>`)
	levelRegex := regexp.MustCompile(`data-level="([0-4])"`)
	countRegex := regexp.MustCompile(`(\d+)\s+contribution`)

	// Total contribution regex e.g. "345 contributions in the last year"
	totalRegex := regexp.MustCompile(`([0-9,]+)\s+contributions?\s+in\s+(the\s+last\s+year|\d{4})`)
	totalMatch := totalRegex.FindStringSubmatch(html)
	if len(totalMatch) > 1 {
		cleaned := strings.ReplaceAll(totalMatch[1], ",", "")
		if tot, err := strconv.Atoi(cleaned); err == nil {
			cal.TotalContributions = tot
		}
	}

	matches := tdRegex.FindAllStringSubmatch(html, -1)
	if len(matches) == 0 {
		// Try fallback td pattern without data-date inside attributes in alternate order
		altTdRegex := regexp.MustCompile(`(?s)<td[^>]*?class="[^"]*ContributionCalendar-day[^"]*"[^>]*?>.*?</td>`)
		matches = altTdRegex.FindAllStringSubmatch(html, -1)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no contribution days found in HTML for %s", username)
	}

	dayMap := make(map[string]ContributionDay)
	var dateList []string

	dateAttrRegex := regexp.MustCompile(`data-date="(\d{4}-\d{2}-\d{2})"`)

	for _, m := range matches {
		tdStr := m[0]
		dateMatch := dateAttrRegex.FindStringSubmatch(tdStr)
		if len(dateMatch) < 2 {
			continue
		}
		dateStr := dateMatch[1]
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		level := "NONE"
		lvlMatch := levelRegex.FindStringSubmatch(tdStr)
		levelNum := 0
		if len(lvlMatch) > 1 {
			levelNum, _ = strconv.Atoi(lvlMatch[1])
			switch levelNum {
			case 1:
				level = "FIRST_QUARTILE"
			case 2:
				level = "SECOND_QUARTILE"
			case 3:
				level = "THIRD_QUARTILE"
			case 4:
				level = "FOURTH_QUARTILE"
			default:
				level = "NONE"
			}
		}

		count := 0
		cntMatch := countRegex.FindStringSubmatch(tdStr)
		if len(cntMatch) > 1 {
			count, _ = strconv.Atoi(cntMatch[1])
		} else if levelNum > 0 {
			count = levelNum * 2 // approximation if exact text isn't inside td tag
		}

		cd := ContributionDay{
			ContributionCount: count,
			ContributionLevel: level,
			Date:              dateStr,
			Weekday:           int(t.Weekday()),
			ParsedDate:        t,
		}

		dayMap[dateStr] = cd
		dateList = append(dateList, dateStr)
	}

	// Group into weeks (weeks start on Sunday = 0)
	var currentWeek ContributionWeek
	for i, dStr := range dateList {
		day := dayMap[dStr]
		if day.Weekday == 0 && len(currentWeek.ContributionDays) > 0 {
			cal.Weeks = append(cal.Weeks, currentWeek)
			currentWeek = ContributionWeek{FirstDay: day.Date, ContributionDays: []ContributionDay{}}
		}
		if len(currentWeek.ContributionDays) == 0 {
			currentWeek.FirstDay = day.Date
		}
		currentWeek.ContributionDays = append(currentWeek.ContributionDays, day)
		if i == len(dateList)-1 && len(currentWeek.ContributionDays) > 0 {
			cal.Weeks = append(cal.Weeks, currentWeek)
		}
	}

	// Calculate total if not set from heading
	if cal.TotalContributions == 0 {
		sum := 0
		for _, w := range cal.Weeks {
			for _, d := range w.ContributionDays {
				sum += d.ContributionCount
			}
		}
		cal.TotalContributions = sum
	}

	return cal, nil
}
