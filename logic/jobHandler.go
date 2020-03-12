package logic

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

func HandleReplayJSON(rjc *models.ReplayJsonContainer) {
	// Create the output dir if necessary
	outputDirAbs, err := filepath.Abs(filepath.Clean(rjc.OutputDir))
	errors.Check(err, "Error getting abs path for output")
	err = os.MkdirAll(outputDirAbs, os.FileMode(700))
	errors.Check(err, "Error creating output dir at "+outputDirAbs)

	log.Println("Populating Frames list...")
	// Populate frame list
	var frames = make([]models.FrameContainer, len(rjc.ReplayJson.Frames))
	for i, frame := range rjc.ReplayJson.Frames {
		// Calculate image Size for the 444 image
		imageSize444 := []uint16{frame.ImageSize422[0] * 2, frame.ImageSize422[1]}
		// Handle TopImage and BottomImage separately
		if frame.TopImage != "" {
			fPath := filepath.Join(rjc.ReplayDirAbs, frame.TopImage)
			frames[i] = models.FrameContainer{
				IsTop:          true,
				Filename:       fPath,
				FilenamePngRel: strings.Replace(frame.TopImage, ".422", ".png", 1),
				FilenamePngAbs: strings.Replace(fPath, ".422", ".png", 1),
			}
		} else if frame.BottomImage != "" {
			fPath := filepath.Join(rjc.ReplayDirAbs, frame.BottomImage)
			frames[i] = models.FrameContainer{
				IsTop:          false,
				Filename:       fPath,
				FilenamePngRel: strings.Replace(frame.BottomImage, ".422", ".png", 1),
				FilenamePngAbs: strings.Replace(fPath, ".422", ".png", 1),
			}
		} else {
			log.Fatalln("This frame has neither topImage nor bottomImage: \n", frame)
		}
		frames[i].ImageSize422 = frame.ImageSize422
		frames[i].ImageSize444 = imageSize444
		frames[i].Compression = rjc.CompressionLevel
	}

	log.Println("Populating Job Queue...")
	// Put all the frames into a channel for the workers
	jobsChan := make(chan *models.FrameContainer, len(frames))
	replyChan := make(chan models.WorkerReply, len(frames))
	for i := range frames {
		jobsChan <- &frames[i]
	}
	close(jobsChan)

	log.Println("Starting some goroutines...")
	// Start some goroutines that convert the images
	for i := 0; i < rjc.NJobs; i++ {
		go FrameWorker(jobsChan, replyChan)
	}
	// Get all the replies from the channel and produce logging output
	for i := 0; i < len(frames); i++ {
		reply := <-replyChan
		if !reply.Success {
			log.Println("[ERROR] Conversion unsuccessful for file ", reply.Fc.Filename, reply.Msg, reply.Err)
		}
		if reply.Success {
			log.Println("Success for image ", reply.Fc.Filename)
		}
	}

	// We only want to modify the original replay.json if the user says so
	if rjc.ModifyOriginal {
		// Write the new filenames into the replayJson object
		for i, frame := range frames {
			if frame.IsTop {
				rjc.ReplayJson.Frames[i].TopImage = frame.FilenamePngRel
			}
		}
		// Generate a new json object with the new filenames
		newJson, err := json.Marshal(rjc.ReplayJson)
		errors.Check(err, "Error creating new replay.json")
		// Overwrite the replay.json
		err = ioutil.WriteFile(rjc.ReplayJsonPath, newJson, os.FileMode(600))
		errors.Check(err, "Error writing new replay.json to file")
	}
}
