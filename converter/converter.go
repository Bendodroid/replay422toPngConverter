package converter

import (
	"image"
	"image/png"
	"io"
	"io/ioutil"
)

const yuv422PxPairSize int = 4

// get4bytes returns 4 bytes individually for a 4-byte slice
func get4bytes(slice *[4]byte) (*byte, *byte, *byte, *byte) {
	return &slice[0], &slice[1], &slice[2], &slice[3]
}

// ConvertFrameToPng converts the frame referenced by the FrameContainer, has to be given source and dest files
func ConvertFrameToPng(r io.Reader, w io.Writer, fc *FrameContainer, encoder *png.Encoder) error {
	var err error
	// byte-array to read the source into
	bytes := make([]byte, fc.ImageSize422[0]*fc.ImageSize422[0]*4)
	// A pair of yuv 422 YCbCr pixels
	var yuv422pixelPair [yuv422PxPairSize]byte
	var px2rgba [yuv422PxPairSize * 2]*byte
	// vars for the channels
	y1, cb, y2, cr, alpha := new(byte), new(byte), new(byte), new(byte), new(byte)
	// Always full alpha
	*alpha = 255
	// Read file into array
	bytes, err = ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	// Create a new image object
	img := image.NewRGBA(image.Rect(0, 0, int(fc.ImageSize444[0]), int(fc.ImageSize444[1])))
	// Iterate over the data in the ugliest way possible
	var i, j int
	for i = 0; i < len(bytes); i += 4 {
		for x := 0; x < yuv422PxPairSize; x++ {
			yuv422pixelPair[x] = bytes[i+x]
		}
		y1, cb, y2, cr = get4bytes(&yuv422pixelPair)

		px2rgba = [yuv422PxPairSize * 2]*byte{y1, cb, cr, alpha, y2, cb, cr, alpha}
		for j = 0; j < yuv422PxPairSize*2; j++ {
			img.Pix[i*2+j] = *px2rgba[j]
		}
	}
	// Encode the image as png and write to file
	err = encoder.Encode(w, img)
	return err
}
