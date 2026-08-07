package version

import "strings"

// servasec backend version, injected at build time
// go build -ldflags "-X github.com/servasec/servasec/backend/version.Version=<version>"
var Version = "develop"

func NormalizeVersion() string {
	return strings.TrimPrefix(Version, "v")
}
