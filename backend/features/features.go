package features

const (
	// Free features (toujours activées)
	FeatureFindings  = "findings"
	FeatureWebhooks  = "webhooks"
	FeatureTeams     = "teams"
	FeatureDashboard = "dashboard"

	// Pro features (nécessitent license)
	FeatureAuditLog         = "audit_log"
	FeatureRiskScoring      = "risk_scoring"
	FeatureAdvancedReporting = "advanced_reporting"
	FeatureMCPServer         = "mcp_server"
	FeatureSSO               = "sso"
)

func FreeFeatures() []string {
	return []string{
		FeatureFindings,
		FeatureWebhooks,
		FeatureTeams,
		FeatureDashboard,
	}
}
