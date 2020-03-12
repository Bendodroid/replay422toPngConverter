package models

import "image/png"

// Info is the struct representing the data in a replay.json file
type Info struct {
	Config []interface{}
	Frames []struct {
		BallDetectionData struct {
			Candidates []struct {
				Center     []int64 `json:"center"`
				Confidence float64 `json:"confidence"`
				Radius     int64   `json:"radius"`
			} `json:"candidates"`
			LastCandidates []struct {
				Center     []int64 `json:"center"`
				Confidence float64 `json:"confidence"`
				Radius     int64   `json:"radius"`
			} `json:"lastCandidates"`
		} `json:"ballDetectionData"`
		BottomImage      string    `json:"bottomImage"`
		FsrLeft          []float64 `json:"fsrLeft"`
		FsrRight         []float64 `json:"fsrRight"`
		HeadMatrixBuffer struct {
			Buffer []struct {
				Head2torso   [][]float64 `json:"head2torso"`
				Timestamp    int64       `json:"timestamp"`
				Torso2ground [][]float64 `json:"torso2ground"`
			} `json:"buffer"`
			Valid bool `json:"valid"`
		} `json:"headMatrixBuffer"`
		ImageSize422 []uint16  `json:"imageSize422"`
		Imu          []float64 `json:"imu"`
		JointAngles  []float64 `json:"jointAngles"`
		SonarDist    []float64 `json:"sonarDist"`
		SonarValid   []bool    `json:"sonarValid"`
		Switches     []int64   `json:"switches"`
		Timestamp    int64     `json:"timestamp"`
		TopImage     string    `json:"topImage"`
	} `json:"frames"`
}

// FrameContainer is a container for an individual frame
type FrameContainer struct {
	IsTop          bool                 // Whether the image is from the topCamera
	Filename       string               // The (rel) filename of the .422 file
	FilenamePngRel string               // The rel path to the .png file
	FilenamePngAbs string               // The abs path to the .png file
	ImageSize422   []uint16             // Image dimensions of the .422 image
	ImageSize444   []uint16             // Image dimensions of the 444 png image
	Compression    png.CompressionLevel // Compression Level for the target image
}

// Worker reply
type WorkerReply struct {
	Fc      *FrameContainer // A reference to the FrameContainer in question
	Success bool            // Whether conversion was a success
	Err     error           // Error value (nil if success)
	Msg     string          // The error message to print (if applicable)
}

type ReplayJsonContainer struct {
	RobotName        string
	RobotPath        string
	ReplayJson       Info
	ReplayDirAbs     string
	ReplayJsonPath   string
	OutputDir        string
	CompressionLevel png.CompressionLevel
	NJobs            int
	ModifyOriginal   bool
}

type SuperJob struct {
	Queue chan ReplayJsonContainer // A list of jobs (one per replay.json)
	Dir   string                   // The Dir containing the outputs for each robot
}
