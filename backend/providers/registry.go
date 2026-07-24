package providers

var registry = map[string]IssueTrackerProvider{
	"github": &GitHubIssueTracker{},
	"gitlab": &GitLabIssueTracker{},
}

func Get(name string) (IssueTrackerProvider, bool) {
	p, ok := registry[name]
	return p, ok
}

func SupportedProviders() []string {
	providers := make([]string, 0, len(registry))
	for name := range registry {
		providers = append(providers, name)
	}
	return providers
}
