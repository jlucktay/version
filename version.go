// Package version has some global strings that should be set with ldflags at compile time, and will attempt to derive
// some (hopefully) sensible default values as a fallback if the appropriate ldflags are not set.
package version

import (
	_ "embed"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"time"
)

//go:generate bash version.sh

const unknownValue = "unknown"

// Symbols used by goreleaser via ldflags.
var (
	// Executable is the name of the currently executing binary.
	// Defaults to the base path of the string returned by calling 'os.Executable()'.
	executable string

	// Version is the semver-compatible git tag that this binary was built from.
	// Defaults to 'v0.0.0-unknown'.
	//go:embed version.txt
	version string

	// BuiltBy is the name of the user that built the currently executing binary.
	// Defaults to the username returned by calling 'user.Current()'.
	builtBy string

	// Commit is the short hash of the commit that this binary was built from.
	commit string

	// BuildDate is the build timestamp of the currently executing binary.
	// Defaults to the modification time (from calling 'os.Stat') on the path returned by calling 'os.Executable()'.
	buildDate string
)

// Details returns a string describing the current binary.
func Details() string {
	var exePath string

	if executable == "" || buildDate == "" {
		var err error

		exePath, err = os.Executable()
		if err != nil {
			exePath = unknownValue
		}
	}

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
		commit = unknownValue
	}

	if buildDate == "" {
		if exePath != unknownValue {
			fi, err := os.Stat(exePath)
			if err != nil {
				buildDate = unknownValue
			} else {
				buildDate = fi.ModTime().Format(time.RFC3339)
			}
		} else {
			buildDate = unknownValue
		}
	}

	return fmt.Sprintf("%s %s built by %s from commit %s with %s at %s.",
		executable, version, builtBy, commit, runtime.Version(), buildDate)
}
