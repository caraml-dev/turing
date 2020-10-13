package internal

import "runtime"

// Info is struct that holds information about the current build of the app
type Info struct {
	Version   string `json:"version"`
	Branch    string `json:"branch"`
	BuildUser string `json:"build_user"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
}

// Build information. Populated at build-time.
var (
	Version   string
	Branch    string
	BuildUser string
	BuildDate string

	VersionInfo = &Info{
		Version:   Version,
		Branch:    Branch,
		BuildUser: BuildUser,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
	}
)
