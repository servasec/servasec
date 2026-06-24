package services

import (
	"math"
	"strings"
	"time"
)

var severityScore = map[string]float64{
	"critical": 1.0,
	"high":     0.7,
	"medium":   0.4,
	"low":      0.1,
	"info":     0.0,
}

var criticalityScore = map[string]float64{
	"critical": 1.0,
	"high":     0.7,
	"medium":   0.4,
	"low":      0.1,
}

const (
	weightSeverity  = 0.40
	weightEPSS      = 0.35
	weightAge       = 0.15
	weightCritAsset = 0.10
)

func SeverityToScore(severity string) float64 {
	s, ok := severityScore[strings.ToLower(severity)]
	if !ok {
		return 0.0
	}
	return s
}

func CriticalityToScore(criticality string) float64 {
	s, ok := criticalityScore[strings.ToLower(criticality)]
	if !ok {
		return 0.4
	}
	return s
}

func CalculateRiskScore(severity string, epssScore *float64, assetCriticality string, createdAt time.Time) float64 {
	sevScore := SeverityToScore(severity)

	var epss float64
	if epssScore != nil {
		epss = *epssScore
		if epss < 0 {
			epss = 0
		}
		if epss > 1 {
			epss = 1
		}
	}

	ageDays := time.Since(createdAt).Hours() / 24
	ageFactor := math.Min(ageDays/365.0, 1.0)

	critScore := CriticalityToScore(assetCriticality)

	score := (sevScore * weightSeverity) +
		(epss * weightEPSS) +
		(ageFactor * weightAge) +
		(critScore * weightCritAsset)

	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return math.Round(score*1000) / 1000
}
