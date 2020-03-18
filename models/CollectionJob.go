package models

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/Bendodroid/replay422toPngConverter/errors"
	"github.com/Bendodroid/replay422toPngConverter/util"
)

type CollectionJob struct {
	Queue   chan RobotJob // A list of jobs (one per replay.json)
	RootDir string        // The RootDir containing the outputs for each robot
}

func (job *CollectionJob) PrePrepare(rootDir, outputDir string, nJobs, pngCompression int, modifyOriginal bool) {
	log.Println("Building jobs...")
	info, err := ioutil.ReadDir(rootDir)
	errors.Check(err, "Reading contents of "+job.RootDir+" failed!")

	job.Queue = make(chan RobotJob, len(info))

	for _, fileInfo := range info {
		log.Printf("Checking %s ...", fileInfo.Name())
		if fileInfo.IsDir() && strings.HasPrefix(fileInfo.Name(), "10.1.24.") {
			log.Printf("Building job for %s ...", fileInfo.Name())
			var rjc RobotJob
			replayJsonPath, replaySubDir := util.FindReplayJson(filepath.Join(rootDir, fileInfo.Name()))
			outputPath := filepath.Join(outputDir, fileInfo.Name(), replaySubDir)
			rjc.PrePrepare(replayJsonPath, outputPath, nJobs, pngCompression, modifyOriginal)
			log.Printf("Building job for %s ... done", fileInfo.Name())
			job.Queue <- rjc
		}
	}
	close(job.Queue)
	log.Println("Building jobs... done")
}

func (job *CollectionJob) Prepare() {
}

func (job *CollectionJob) Run() {
	for rjc := range job.Queue {
		rjc.Prepare()
		rjc.Run()
	}
}
