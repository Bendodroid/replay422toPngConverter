package util

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bendodroid/replay422toPngConverter/errors"
	"github.com/Bendodroid/replay422toPngConverter/models"
)

type replayJsonNotFoundError struct {
	dir string
}

func (e replayJsonNotFoundError) Error() string {
	return "No replay.json could be found in " + e.dir
}

// LoadReplayJson loads the json file into target
func LoadReplayJson(filename string, target *models.Info) {
	// Read []byte from file
	dat, err := ioutil.ReadFile(filename)
	errors.Check(err, "Error reading from file")
	// Parse json
	err = json.Unmarshal(dat, target)
	errors.Check(err, "Failed to parse json")
}

// FindReplayJson finds a replay.json in a dir's replay_* subdirectory
func FindReplayJson(rjc *models.ReplayJsonContainer) error {
	var replayDir string
	// Match the replay_* directory
	if f, _ := MatchPatternInPath(rjc.RobotPath, "replay_*"); f != nil {
		replayDir = filepath.Join(rjc.RobotPath, f.Name())
		// It's a match -> Generate ReplayJsonContainer and return it
		GetReplayJsonFromDir(rjc, &replayDir)
		return nil
	}
	return replayJsonNotFoundError{dir: rjc.RobotPath}
}

// GetReplayJsonFromDir returns a converter.ReplayJsonContainer for a dir containing a replay.json
func GetReplayJsonFromDir(rjc *models.ReplayJsonContainer, replayDir *string) {
	var err error
	// Get the abs path to the replay.json
	rjc.ReplayDirAbs, err = filepath.Abs(filepath.Clean(*replayDir))
	errors.Check(err, "Error getting abs path for replay dir")
	// Append replay.json
	rjc.ReplayJsonPath = filepath.Join(rjc.ReplayDirAbs, "replay.json")
	errors.Check(err, "Error getting path for replay.json")
	// Load replay.json
	LoadReplayJson(rjc.ReplayJsonPath, &rjc.ReplayJson)
	log.Printf("Successfully loaded %s !\n", rjc.ReplayJsonPath)
}

func MatchPatternInPath(path, pattern string) (os.FileInfo, error) {
	// Read contents of dir
	files, err := ioutil.ReadDir(path)
	errors.Check(err, "Error reading contents of directory "+path)
	// Match the dir pattern
	for _, f := range files {
		if isMatch, _ := filepath.Match(pattern, f.Name()); isMatch {
			return f, nil
		}
	}
	return nil, err
}

func ExpandHome(path string) string {
	dir, _ := os.UserHomeDir()
	if path == "~" {
		return dir
	} else if strings.HasPrefix(path, "~/") {
		return filepath.Join(dir, path[2:])
	} else {
		return path
	}
}
