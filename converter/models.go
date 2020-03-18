package converter

import "image/png"

// FrameContainer is a container for an individual frame
type FrameContainer struct {
	IsTop        bool                 // Whether the image is from the topCamera
	Path422      string               // The (rel) filename of the .422 file
	PngFileName  string               // The rel path to the .png file
	PathPngAbs   string               // The abs path to the .png file
	ImageSize422 [2]uint16            // Image dimensions of the .422 image
	ImageSize444 [2]uint16            // Image dimensions of the 444 png image
	Compression  png.CompressionLevel // Compression Level for the target image
}

// Worker reply
type WorkerReply struct {
	Fc      *FrameContainer // A reference to the FrameContainer in question
	Success bool            // Whether conversion was a success
	Err     error           // Error value (nil if success)
	Msg     string          // The error message to print (if applicable)
}
