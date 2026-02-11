package version

import "fmt"

// Build-time variables injected via ldflags
var (
	// Version is the semantic version of the binary
	Version = "dev"
	// CommitHash is the git commit hash
	CommitHash = "unknown"
	// BuildTime is the build timestamp
	BuildTime = "unknown"
	// BinaryName is the name of the binary
	BinaryName = "localrecall-mcp-server"
)

// GetVersionInfo returns a formatted version string
func GetVersionInfo() string {
	return fmt.Sprintf("%s version %s (commit: %s, built: %s)",
		BinaryName, Version, CommitHash, BuildTime)
}
