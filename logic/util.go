package logic

import "github.com/Bendodroid/replay422toPngConverter/models"

// SendWorkerReply send a WorkerReply object to the given channel
func SendWorkerReply(ch chan<- models.WorkerReply, fc *models.FrameContainer, success bool, err error, msg string) {
	ch <- models.WorkerReply{
		Fc:      fc,
		Success: success,
		Err:     err,
		Msg:     msg,
	}
}
