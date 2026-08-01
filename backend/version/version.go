package version

// servasec backend version, injected at build time
// go build -ldflags "-X github.com/servasec/servasec/backend/version.Version=<version>"
var Version = "develop"
