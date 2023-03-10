// Package version has some unexported global strings that should be set with ldflags at compile time, and if any of
// the ldflags are not set, will attempt to derive some (hopefully) sensible default values as a fallback.
//
// The list of symbols in this package that can be set with ldflags is as follows:
//   - executable
//   - version
//   - builtBy
//   - commit
//   - builtWith
//   - buildDate
//
// One simple example of how to set ldflags when calling 'go build':
//
//	go build -ldflags="-X 'go.jlucktay.dev/version.version=v1.2.3'"
package version

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

// A fallback value if errors are returned when attempting to look up sensible defaults.
const unknownValue = "unknown"

// These symbols can be populated with ldflags when building.
//
//nolint:gochecknoglobals // This is the whole point of this package.
var (
	// Executable is the name of the currently executing binary.
	// Defaults to the base path of the string returned by calling 'os.Executable()'.
	executable string

	// Version is the semver-compatible git tag that this binary was built from.
	// Defaults to 'v0.0.0-unknown'.
	version string

	// BuiltBy is the name of the user that built the currently executing binary.
	// Defaults to the username returned by calling 'user.Current()'.
	builtBy string

	// Commit is the short hash of the commit that this binary was built from.
	// Defaults to the value stored against the 'vcs.revision' key in 'debug.BuildSetting' returned by calling
	// 'debug.ReadBuildInfo()'.
	commit string

	// BuiltWith is the version of the Go toolchain that built the binary.
	// Defaults to the 'GoVersion' field returned by calling 'debug.ReadBuildInfo'.
	builtWith string

	// BuildDate is the build timestamp of the currently executing binary.
	// Defaults to the modification time (from calling 'os.Stat') on the path returned by calling 'os.Executable()'.
	buildDate string
)

// Details returns a string describing the caller.
func Details() string {
	// Some variables we might need later.
	var (
		exePath   string
		buildInfo *debug.BuildInfo
	)

	// Pre-populate these if they are needed.
	if executable == "" || buildDate == "" {
		var err error

		exePath, err = os.Executable()
		if err != nil {
			exePath = unknownValue
		}
	}

	if commit == "" || builtWith == "" {
		var biOK bool
		buildInfo, biOK = debug.ReadBuildInfo()

		if !biOK && commit == "" {
			commit = unknownValue
		}

		if !biOK && builtWith == "" {
			commit = unknownValue
		}
	}

	// Check each symbol in turn, and populate if not already set.
	if executable == "" {
		if exePath != unknownValue {
			executable = filepath.Base(exePath)
		} else {
			executable = unknownValue
		}
	}

	if version == "" {
		version = "v0.0.0-" + unknownValue
	}

	if builtBy == "" {
		u, err := user.Current()
		if err != nil {
			builtBy = unknownValue
		} else {
			builtBy = u.Username
		}
	}

	if commit == "" {
		for index := range buildInfo.Settings {
			switch strings.ToLower(buildInfo.Settings[index].Key) {
			case "vcs.revision":
				commit = buildInfo.Settings[index].Value + commit
			case "vcs.modified":
				if strings.EqualFold(buildInfo.Settings[index].Value, "true") {
					commit += "-dirty"
				}
			}
		}
	}

	if builtWith == "" {
		builtWith = buildInfo.GoVersion
	}

	if buildDate == "" {
		buildDate = unknownValue

		if exePath != unknownValue {
			fi, err := os.Stat(exePath)
			if err == nil {
				buildDate = fi.ModTime().Format(time.RFC3339)
			}
		}
	}

	return fmt.Sprintf("%s %s built by %s from commit %s with %s at %s.",
		executable, version, builtBy, commit, builtWith, buildDate)
}
