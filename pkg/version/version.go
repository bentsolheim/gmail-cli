package version

// These variables are set at build time via ldflags
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// String returns a formatted version string.
func String() string {
	return Version
}

// Full returns the full version info including commit and build date.
func Full() string {
	return Version + " (" + Commit + ") built " + BuildDate
}
