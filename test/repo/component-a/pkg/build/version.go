package build

// This string is set in the Go runner with
// `-ldflags -X ".../pkg/BuildVersion=..."`.
var buildVersion = "0.0.0" //nolint: gochecknoglobals // intentional

func GetBuildVersion() string {
	return buildVersion
}
