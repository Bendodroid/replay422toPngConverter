package logic

import (
	"image/png"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/Bendodroid/replay422toPngConverter/errors"
	"github.com/Bendodroid/replay422toPngConverter/models"
	"github.com/Bendodroid/replay422toPngConverter/util"
)

func HandleSuperJob(job *models.SuperJob) {
	for rjc := range job.Queue {
		log.Println("Loading replay.json for", rjc.RobotName)
		err := util.FindReplayJson(&rjc)
		errors.Check(err, "No replay.json could be found in dir "+rjc.ReplayDirAbs+". Are your files intact?")
		HandleReplayJSON(&rjc)
	}
}

func BuildSuperJob(job *models.SuperJob, nJobs int, pngCompression int, modifyOriginal bool) {
	info, err := ioutil.ReadDir(job.Dir)
	errors.Check(err, "Reading contents of "+job.Dir+" failed!")
	job.Queue = make(chan models.ReplayJsonContainer, len(info))

	for _, fileInfo := range info {
		log.Println("Starting to build job for", fileInfo.Name())
		if fileInfo.IsDir() && strings.HasPrefix(fileInfo.Name(), "10.1.24.") {
			rjc := models.ReplayJsonContainer{}
			rjc.RobotName = fileInfo.Name()
			rjc.RobotPath = filepath.Join(job.Dir, fileInfo.Name())
			rjc.OutputDir = rjc.ReplayDirAbs
			rjc.NJobs = nJobs
			rjc.CompressionLevel = png.CompressionLevel(pngCompression)
			rjc.ModifyOriginal = modifyOriginal
			log.Println("Finished building job for", rjc.ReplayJsonPath)
			job.Queue <- rjc
		}
	}
	close(job.Queue)
}
