package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Bendodroid/replay422toPngConverter/errors"
	"github.com/Bendodroid/replay422toPngConverter/replay"
	"github.com/Bendodroid/replay422toPngConverter/util"
)

// A container for an individual frame
type frameContainer struct {
	isTop          bool     // Whether the image is from the topCamera
	filename       string   // The (rel) filename of the .422 file
	filenamePngRel string   // The rel path to the .png file
	filenamePngAbs string   // The abs path to the .png file
	imageSize422   []uint16 // Image dimensions of the .422 image
	imageSize444   []uint16 // Image dimensions of the 444 png image
}

// Worker reply
type workerReply struct {
	fc      *frameContainer // A reference to the frameContainer in question
	success bool            // Whether conversion was a success
	err     error           // Error value (nil if success)
	msg     string          // The error message to print (if applicable)
}

// sendWorkerReply send a workerReply object to the given channel
func sendWorkerReply(ch chan<- workerReply, fc *frameContainer, success bool, err error, msg string) {
	ch <- workerReply{
		fc:      fc,
		success: success,
		err:     err,
		msg:     msg,
	}
}

// convertToPng converts the frame referenced by the frameContainer, has to be given source and dest files
func convertToPng(r io.Reader, w io.Writer, fc *frameContainer) error {
	var err error
	// byte-array to read the source into
	var bytes []uint8
	// vars for the pixels
	var y1, cb, y2, cr uint8
	// Read file into array
	bytes, err = ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	// Create a new image object
	img := image.NewRGBA(image.Rect(0, 0, int(fc.imageSize444[0]), int(fc.imageSize444[1])))
	// Iterate over the data in the ugliest way possible
	for i := 0; i < len(bytes); i += 4 {
		y1 = bytes[i+0]
		cb = bytes[i+1]
		y2 = bytes[i+2]
		cr = bytes[i+3]
		i2 := i * 2
		img.Pix[i2+0] = y1
		img.Pix[i2+1] = cb
		img.Pix[i2+2] = cr
		img.Pix[i2+3] = 255
		img.Pix[i2+4] = y2
		img.Pix[i2+5] = cb
		img.Pix[i2+6] = cr
		img.Pix[i2+7] = 255
	}
	// Encode the image as png and write to file
	err = png.Encode(w, img)

	return err
}

// worker is the method that runs in parallel to convert frames
func worker(jobsCh <-chan *frameContainer, replyCh chan<- workerReply) {
	// Get new job from channel
	for job := range jobsCh {
		// Open the source .422 file
		f422, err := os.Open(job.filename)
		if err != nil {
			sendWorkerReply(replyCh, job, false, err, "While trying to open"+job.filename+" an error occurred:\n")
			return
		}
		// Open the target file
		fPng, err := os.Create(job.filenamePngAbs)
		if err != nil {
			sendWorkerReply(replyCh, job, false, err, "Failed creating new file for "+job.filenamePngAbs)
			return
		}
		// Convert the file
		err = convertToPng(f422, fPng, job)
		if err != nil {
			_ = os.Remove(job.filenamePngAbs)
			sendWorkerReply(replyCh, job, false, err, "While trying to convert "+job.filename+" an error occurred:\n")
			return
		}
		// Close the files
		f422.Close()
		fPng.Close()
		// Send a success-reply for this job
		sendWorkerReply(replyCh, job, true, nil, "")
	}
}

func main() {
	var err error
	startTime := time.Now()

	// Input arguments
	replayDir := flag.String("replayDir", ".", "A dir containing a replay.json and images")
	outputDir := flag.String("outputDir", ".", "Where to put the results")
	nJobs := flag.Int("j", runtime.NumCPU()+2, "Number of jobs to use for converting")
	help := flag.Bool("h", false, "Display Help text")
	flag.Parse()
	if *help {
		flag.PrintDefaults()
	}

	var replayJson = new(replay.Info)
	// Get the abs path to the replay.json
	replayDirAbs, err := filepath.Abs(filepath.Clean(*replayDir))
	errors.Check(err, "Error getting abs path for replay dir")
	replayJsonPath := filepath.Join(replayDirAbs, "replay.json")
	errors.Check(err, "Error getting path for replay.json")
	// Load replay.json
	util.LoadJSON(replayJsonPath, replayJson)
	log.Println("Successfully loaded replay.json!")

	// Create the output dir if necessary
	outputDirAbs, err := filepath.Abs(filepath.Clean(*outputDir))
	errors.Check(err, "Error getting abs path for output")
	err = os.MkdirAll(outputDirAbs, os.FileMode(700))
	errors.Check(err, "Error creating output dir at "+outputDirAbs)

	// Populate frame list
	var frames = make([]frameContainer, len(replayJson.Frames))
	for i, frame := range replayJson.Frames {
		// Calculate image Size for the 444 image
		imageSize444 := []uint16{frame.ImageSize422[0] * 2, frame.ImageSize422[1]}
		// Handle TopImage and BottomImage separately
		if frame.TopImage != "" {
			fPath := filepath.Join(replayDirAbs, frame.TopImage)
			frames[i] = frameContainer{
				isTop:          true,
				filename:       fPath,
				filenamePngRel: strings.Replace(frame.TopImage, ".422", ".png", 1),
				filenamePngAbs: strings.Replace(fPath, ".422", ".png", 1),
				imageSize422:   frame.ImageSize422,
				imageSize444:   imageSize444,
			}
		} else if frame.BottomImage != "" {
			fPath := filepath.Join(replayDirAbs, frame.BottomImage)
			frames[i] = frameContainer{
				isTop:          false,
				filename:       fPath,
				filenamePngRel: strings.Replace(frame.BottomImage, ".422", ".png", 1),
				filenamePngAbs: strings.Replace(fPath, ".422", ".png", 1),
				imageSize422:   frame.ImageSize422,
				imageSize444:   imageSize444,
			}
		} else {
			log.Fatalln("This frame has neither topImage nor bottomImage: \n", frame)
		}
	}

	// Put all the frames into a channel for the workers
	jobsChan := make(chan *frameContainer, len(frames))
	replyChan := make(chan workerReply, len(frames))
	for i, _ := range frames {
		jobsChan <- &frames[i]
	}
	close(jobsChan)

	// Start some goroutines that convert the images
	for i := 0; i < *nJobs; i++ {
		go worker(jobsChan, replyChan)
	}
	// Get all the replies from the channel and produce logging output
	for i := 0; i < len(frames); i++ {
		reply := <-replyChan
		if !reply.success {
			log.Println("[ERROR] Conversion unsuccessful for file ", reply.fc.filename, reply.msg, reply.err)
		}
		if reply.success {
			log.Println("Success for image ", reply.fc.filename)
		}
	}

	// Write the new filenames into the replayJson object
	for i, frame := range frames {
		if frame.isTop {
			replayJson.Frames[i].TopImage = frame.filenamePngRel
		}
	}
	// Generate a new json object with the new filenames
	newJson, err := json.Marshal(replayJson)
	errors.Check(err, "Error creating new replay.json")
	// Overwrite the replay.json
	err = ioutil.WriteFile(replayJsonPath, newJson, os.FileMode(600))
	errors.Check(err, "Error writing new replay.json to file")

	// Finish and print time it took
	fmt.Println("Finished converting after ", time.Since(startTime).String(), "!")
}
