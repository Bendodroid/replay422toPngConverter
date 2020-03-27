package models

import (
	"encoding/json"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bendodroid/replay422toPngConverter/converter"
	"github.com/Bendodroid/replay422toPngConverter/errors"
)

// RobotJob is the data Structure holding information to convert data from a single robot
type RobotJob struct {
	// Public
	ReplayJson       Info
	ReplayJsonPath   string
	ReplayJsonDir    string
	OutputDir        string
	RobotName        string
	CompressionLevel png.CompressionLevel
	NJobs            int
	ModifyOriginal   bool
	// Private
	frames   []converter.FrameContainer
	encoders []*png.Encoder
}

// PrePrepare fills out the general info for one RobotJob
func (job *RobotJob) PrePrepare(replayJsonPath, outputDir string, nJobs, pngCompression int, modifyOriginal bool) {
	// Fill out the details...
	job.ReplayJsonPath = replayJsonPath
	job.ReplayJsonDir = filepath.Dir(replayJsonPath)
	job.OutputDir = outputDir
	job.RobotName = filepath.Clean(filepath.Join(replayJsonPath, "..", ".."))
	job.NJobs = nJobs
	job.CompressionLevel = png.CompressionLevel(pngCompression)
	job.ModifyOriginal = modifyOriginal
	// Create the output dir if necessary
	err := os.MkdirAll(job.OutputDir, 0755)
	errors.Check(err, "Error creating output dir at "+job.OutputDir)
	log.Printf("Using %s as output dir for %s", job.OutputDir, job.RobotName)
}

// Prepare reads the replay.json into the RobotJob and populates the frame list
func (job *RobotJob) Prepare() {
	// Load replay.json
	log.Println("Loading replay.json for", job.RobotName)
	job.loadReplayJson()
	// Populate frame list
	log.Println("Populating Frames list...")
	job.frames = make([]converter.FrameContainer, len(job.ReplayJson.Frames))
	var fPath422Abs string
	for i, frame := range job.ReplayJson.Frames {
		// Calculate image Size for the 444 image
		imageSize444 := [2]uint16{frame.ImageSize422[0] * 2, frame.ImageSize422[1]}
		// Handle TopImage and BottomImage separately
		if frame.TopImage != "" {
			// It's an image from the top camera
			fPath422Abs = filepath.Join(job.ReplayJsonDir, frame.TopImage)
			pathPngRel := strings.Replace(frame.TopImage, ".422", ".png", 1)
			job.frames[i] = converter.FrameContainer{
				IsTop:       true,
				Path422:     fPath422Abs,
				PngFileName: pathPngRel,
				PathPngAbs:  filepath.Join(job.OutputDir, pathPngRel),
			}
		} else if frame.BottomImage != "" {
			// It's an image from the bottom camera
			fPath422Abs = filepath.Join(job.ReplayJsonDir, frame.BottomImage)
			pathPngRel := strings.Replace(frame.BottomImage, ".422", ".png", 1)
			job.frames[i] = converter.FrameContainer{
				IsTop:       false,
				Path422:     fPath422Abs,
				PngFileName: pathPngRel,
				PathPngAbs:  filepath.Join(job.OutputDir, pathPngRel),
			}
		} else {
			log.Fatalln("This frame has neither topImage nor bottomImage: \n", frame)
		}
		job.frames[i].ImageSize422 = frame.ImageSize422
		job.frames[i].ImageSize444 = imageSize444
		job.frames[i].Compression = job.CompressionLevel
	}
}

// Run creates some encoders, starts some goroutines and synchronizes at the end
func (job *RobotJob) Run() {
	// Populate list of encoders
	job.encoders = make([]*png.Encoder, job.NJobs)
	for i := range job.encoders {
		job.encoders[i] = &png.Encoder{CompressionLevel: job.CompressionLevel}
	}

	// Put all the frames into a channel for the workers
	log.Println("Populating Job Queue...")
	jobsChan := make(chan *converter.FrameContainer, len(job.frames))
	replyChan := make(chan converter.WorkerReply, len(job.frames))
	for i := range job.frames {
		jobsChan <- &job.frames[i]
	}
	close(jobsChan)

	// Start some goroutines that convert the images
	log.Println("Starting some goroutines...")
	for i := 0; i < job.NJobs; i++ {
		go converter.FrameWorker(jobsChan, replyChan, job.encoders[i])
	}

	// Get all the replies from the channel and produce logging output
	for i := 0; i < len(job.frames); i++ {
		reply := <-replyChan
		if !reply.Success {
			log.Println("[ERROR] Conversion unsuccessful for file ", reply.Fc.Path422, reply.Msg, reply.Err)
		}
		if reply.Success {
			log.Printf("[Success] %s -> %s", filepath.Base(reply.Fc.Path422), reply.Fc.PathPngAbs)
		}
	}
	close(replyChan)

	// We only want to modify the original replay.json if the user says so
	if job.ModifyOriginal {
		// Write the new filenames into the replayJson object
		for i, frame := range job.frames {
			if frame.IsTop {
				job.ReplayJson.Frames[i].TopImage = frame.PngFileName
			}
		}
		// Generate a new json object with the new filenames
		newJson, err := json.Marshal(job.ReplayJson)
		errors.Check(err, "Error creating new replay.json")
		// Overwrite the replay.json
		err = ioutil.WriteFile(job.ReplayJsonPath, newJson, os.FileMode(600))
		errors.Check(err, "Error writing new replay.json to file")
	}
}

// loadReplayJson loads the json file into target
func (job *RobotJob) loadReplayJson() {
	// Read []byte from file
	dat, err := ioutil.ReadFile(job.ReplayJsonPath)
	errors.Check(err, "Error reading from file "+job.ReplayJsonPath)
	// Parse json
	err = json.Unmarshal(dat, &job.ReplayJson)
	errors.Check(err, "Failed to parse json "+job.ReplayJsonPath)
}
