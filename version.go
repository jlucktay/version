// Package version has some global strings that should be set with ldflags at compile time, and will attempt to derive
// some (hopefully) sensible default values as a fallback if the appropriate ldflags are not set.
package version

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Symbols used by goreleaser via ldflags.
var (
	// Executable is the name of the currently executing binary.
	// Defaults to the base path of the string returned by calling 'os.executable()'.
	executable string

	// Version is the semver-compatible git tag that this binary was built from.
	// Defaults to 'v0.0.0'.
	version string

	// Commit is the short hash of the commit that this binary was built from.
	// Defaults to the value returned by running 'git describe --always --dirty'.
	commit string

	// BuildDate is the build timestamp of the currently executing binary.
	// Defaults to the modification time (from calling 'os.Stat') on the path returned by calling 'os.Executable()'.
	buildDate string

	// BuiltBy is the name of the user that built the currently executing binary.
	// Defaults to the username returned by calling 'user.Current()'.
	builtBy string
)

// Details returns a string describing the current binary.
func Details() (string, error) {
	var exePath string

	if executable == "" || buildDate == "" {
		var err error

		exePath, err = os.Executable()
		if err != nil {
			return "", fmt.Errorf("could not look up path of current executable: %w", err)
		}
	}

	if executable == "" {
		executable = filepath.Base(exePath)
	}

	if version == "" {
		version = "v0.0.0-unknown"
	}

	if commit == "" {
		if _, err := exec.LookPath("git"); err != nil {
			commit = "git-not-available"
		} else {
			cmd := exec.Command("git", "describe", "--always", "--dirty")

			output, err := cmd.Output()
			if err != nil {
				return "", fmt.Errorf("could not run '%s': %w", cmd, err)
			}

			commit = strings.TrimSpace(string(output))
		}
	}

	if buildDate == "" {
		osfi, err := os.Stat(exePath)
		if err != nil {
			return "", fmt.Errorf("could not stat current executable: %w", err)
		}

		buildDate = osfi.ModTime().Format(time.RFC3339)
	}

	if builtBy == "" {
		currUser, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("could not get current user: %w", err)
		}

		builtBy = currUser.Username
	}

	return fmt.Sprintf("%s %s built by %s from commit %s with %s at %s.",
		executable, version, builtBy, commit, runtime.Version(), buildDate), nil
}
