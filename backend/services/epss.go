package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/servasec/servasec/backend/config"
)

var cvePattern = regexp.MustCompile(`CVE-\d{4}-\d{4,}`)

func IsCVE(ruleID string) bool {
	return cvePattern.MatchString(strings.ToUpper(ruleID))
}

type epssResponse struct {
	Data []struct {
		CVE  string `json:"cve"`
		EPSS string `json:"epss"`
	} `json:"data"`
}

type EPSSClient struct {
	mu        sync.RWMutex
	cache     map[string]float64
	fetchedAt time.Time
	ttl       time.Duration
	client    *http.Client
}

func NewEPSSClient() *EPSSClient {
	return &EPSSClient{
		cache:  make(map[string]float64),
		ttl:    24 * time.Hour,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *EPSSClient) GetScore(cveID string) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	score, ok := c.cache[cveID]
	return score, ok
}

func (c *EPSSClient) Fetch(cveIDs []string) (map[string]float64, error) {
	if len(cveIDs) == 0 {
		return nil, nil
	}

	unique := make([]string, 0, len(cveIDs))
	seen := make(map[string]bool)
	for _, id := range cveIDs {
		id := strings.ToUpper(strings.TrimSpace(id))
		if id != "" && !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}
	if len(unique) == 0 {
		return nil, nil
	}

	params := url.Values{}
	params.Set("cve", strings.Join(unique, ","))

	u := fmt.Sprintf("https://api.first.org/epss/v2/reproduce?%s", params.Encode())

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("epss request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("epss request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("epss API returned status %d", resp.StatusCode)
	}

	var result epssResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("epss decode: %w", err)
	}

	out := make(map[string]float64, len(result.Data))
	c.mu.Lock()
	for _, d := range result.Data {
		var score float64
		if _, err := fmt.Sscanf(d.EPSS, "%f", &score); err == nil {
			out[d.CVE] = score
			c.cache[d.CVE] = score
		}
	}
	c.fetchedAt = time.Now()
	c.mu.Unlock()

	return out, nil
}

func (c *EPSSClient) NeedsRefresh() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return time.Since(c.fetchedAt) > c.ttl
}

func (c *EPSSClient) CachedCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

func StartEPSSSync(client *EPSSClient) {
	log.Printf("[epss] starting background sync")
	go func() {
		syncEPSSFindings(client)
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			syncEPSSFindings(client)
		}
	}()
}

func syncEPSSFindings(client *EPSSClient) {
	if !client.NeedsRefresh() {
		return
	}
	log.Printf("[epss] fetching scores for open findings")

	type findingRow struct {
		ID     uint
		RuleID string
	}

	var findings []findingRow
	config.DB.Raw(`
		SELECT id, rule_id
		FROM findings
		WHERE status IN ('open', 'confirmed')
		AND rule_id ~ '^CVE-\d{4}-\d{4,}'
		ORDER BY id
	`).Scan(&findings)

	if len(findings) == 0 {
		log.Printf("[epss] no CVE findings to sync")
		return
	}

	var cveIDs []string
	for _, f := range findings {
		cveIDs = append(cveIDs, f.RuleID)
	}

	for i := 0; i < len(cveIDs); i += 100 {
		batch := cveIDs[i:]
		if len(batch) > 100 {
			batch = batch[:100]
		}

		scores, err := client.Fetch(batch)
		if err != nil {
			log.Printf("[epss] fetch error (batch %d): %v", i/100, err)
			continue
		}

		for _, f := range findings {
			score, ok := scores[f.RuleID]
			if !ok {
				continue
			}

			var sev string
			config.DB.Raw(`SELECT severity FROM findings WHERE id = ?`, f.ID).Scan(&sev)
			newScore := CalculateRiskScore(sev, &score, "medium", time.Now())

			config.DB.Exec(
				`UPDATE findings SET epss_score = ?, risk_score = ? WHERE id = ?`,
				score, newScore, f.ID,
			)
		}

		time.Sleep(1 * time.Second)
	}

	log.Printf("[epss] sync complete, %d scores cached", client.CachedCount())
}
