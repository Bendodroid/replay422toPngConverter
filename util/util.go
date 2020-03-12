package util

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path/filepath"

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
	// Read contents of dir (something like 10.1.24.xx)
	files, err := ioutil.ReadDir(rjc.RobotPath)
	errors.Check(err, "Error reading contents of directory "+rjc.RobotPath)
	// Match the replay_* directory
	for _, f := range files {
		if isMatch, _ := filepath.Match("replay_*", f.Name()); f.IsDir() && isMatch {
			replayDir = filepath.Join(rjc.RobotPath, f.Name())
			// It's a match -> Generate ReplayJsonContainer and return it
			GetReplayJsonFromDir(rjc, &replayDir)
			return nil
		}
	}
	// Found nothing -> error
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
