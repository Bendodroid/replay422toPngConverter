package logic

import (
	"os"

	"github.com/Bendodroid/replay422toPngConverter/converter"
	"github.com/Bendodroid/replay422toPngConverter/models"
)

// FrameWorker is the method that is run in parallel to convert frames
func FrameWorker(jobsCh <-chan *models.FrameContainer, replyCh chan<- models.WorkerReply) {
	// Get new job from channel
	for job := range jobsCh {
		// Open the source .422 file
		f422, err := os.Open(job.Path422)
		if err != nil {
			SendWorkerReply(replyCh, job, false, err, "While trying to open"+job.Path422+" an error occurred:\n")
			return
		}
		// Open the target file
		fPng, err := os.Create(job.PathPngAbs)
		if err != nil {
			SendWorkerReply(replyCh, job, false, err, "Failed creating new file for "+job.PathPngAbs)
			return
		}
		// Convert the file
		err = converter.ConvertFrameToPng(f422, fPng, job)
		if err != nil {
			_ = os.Remove(job.PathPngAbs)
			SendWorkerReply(replyCh, job, false, err, "While trying to convert "+job.Path422+" an error occurred:\n")
			return
		}
		// Close the files
		f422.Close()
		fPng.Close()
		// Send a success-reply for this job
		SendWorkerReply(replyCh, job, true, nil, "")
	}
}
