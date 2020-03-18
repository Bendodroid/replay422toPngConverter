package converter

import (
	"image/png"
	"os"
)

// FrameWorker is the method that is run in parallel to convert frames
func FrameWorker(jobsCh <-chan *FrameContainer, replyCh chan<- WorkerReply, encoder *png.Encoder) {
	// Get new job from channel
	for job := range jobsCh {
		// Open the source .422 file
		f422, err := os.Open(job.Path422)
		if err != nil {
			sendWorkerReply(replyCh, job, false, err, "While trying to open"+job.Path422+" an error occurred:\n")
			return
		}
		// Open the target file
		fPng, err := os.Create(job.PathPngAbs)
		if err != nil {
			sendWorkerReply(replyCh, job, false, err, "Failed creating new file for "+job.PathPngAbs)
			return
		}
		// Convert the file
		err = ConvertFrameToPng(f422, fPng, job, encoder)
		if err != nil {
			_ = os.Remove(job.PathPngAbs)
			sendWorkerReply(replyCh, job, false, err, "While trying to convert "+job.Path422+" an error occurred:\n")
			return
		}
		// Close the files
		f422.Close()
		fPng.Close()
		// Send a success-reply for this job
		sendWorkerReply(replyCh, job, true, nil, "")
	}
}

// sendWorkerReply send a WorkerReply object to the given channel
func sendWorkerReply(ch chan<- WorkerReply, fc *FrameContainer, success bool, err error, msg string) {
	ch <- WorkerReply{
		Fc:      fc,
		Success: success,
		Err:     err,
		Msg:     msg,
	}
}
