package pro

import "time"

type TopRiskyFinding struct {
	ID        uint    `json:"id"`
	Title     string  `json:"title"`
	RiskScore float64 `json:"riskScore"`
	Severity  string  `json:"severity"`
}

type RiskBucket struct {
	Label string `json:"label"`
	Count int64  `json:"count"`
}

type RiskScorer interface {
	StartEPSSSync()
	CalculateScore(severity string, epssScore *float64, assetCriticality string, createdAt time.Time) *float64
	DashboardStats(accessibleIDs []string) (*float64, []TopRiskyFinding, []RiskBucket)
}

type noopRiskScorer struct{}

func (n *noopRiskScorer) StartEPSSSync() {}

func (n *noopRiskScorer) CalculateScore(severity string, epssScore *float64, assetCriticality string, createdAt time.Time) *float64 {
	return nil
}

func (n *noopRiskScorer) DashboardStats(accessibleIDs []string) (*float64, []TopRiskyFinding, []RiskBucket) {
	return nil, nil, nil
}

var Risk RiskScorer = &noopRiskScorer{}
