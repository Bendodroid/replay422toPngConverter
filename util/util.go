package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bendodroid/replay422toPngConverter/errors"
)

// FindReplayJson finds a replay.json in a dir's replay_* subdirectory
func FindReplayJson(robotDir string) (string, string) {
	// Match the replay_* directory
	if f := MatchPatternInPath(robotDir, "replay_*"); f != nil {
		// It's a match -> Generate RobotJob and return it
		return filepath.Join(robotDir, f.Name(), "replay.json"), f.Name()
	}
	return "", ""
}

// MatchPatternInPath takes a path and a pattern and returns the first element matching the pattern
// (used to find replay_* directories)
func MatchPatternInPath(path, pattern string) os.FileInfo {
	// Read contents of dir
	files, err := ioutil.ReadDir(path)
	errors.Check(err, "Error reading contents of directory "+path)
	// Match the dir pattern
	for _, f := range files {
		if isMatch, _ := filepath.Match(pattern, f.Name()); isMatch {
			return f
		}
	}
	return nil
}

// ExpandPath expands a given path starting with ~
func ExpandPath(path string) string {
	path = filepath.Clean(path)
	homeDir, _ := os.UserHomeDir()
	if path == "~" {
		return homeDir
	} else if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:])
	} else {
		abs, err := filepath.Abs(path)
		errors.Check(err, "Error getting absolute path for "+path)
		return abs
	}
}
