package models

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
		ImageSize422 [2]uint16 `json:"imageSize422"`
		Imu          []float64 `json:"imu"`
		JointAngles  []float64 `json:"jointAngles"`
		SonarDist    []float64 `json:"sonarDist"`
		SonarValid   []bool    `json:"sonarValid"`
		Switches     []int64   `json:"switches"`
		Timestamp    int64     `json:"timestamp"`
		TopImage     string    `json:"topImage"`
	} `json:"frames"`
}

type Job interface {
	PrePrepare(replayJsonPath, outputDir string, nJobs, pngCompression int, modifyOriginal bool)
	Prepare()
	Run()
}
