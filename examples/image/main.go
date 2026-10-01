package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/elmyrockers/go-zkfp"
)

// Parameter codes from the SDK documentation (Appendix 1).
const (
	paramImageWidth  = 1
	paramImageHeight = 2
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// readInt reads a 4-byte integer device parameter.
func readInt(device *zkfp.Device, code int) (int, error) {
	value, err := device.GetParameter(code, 4)
	if err != nil {
		return 0, err
	}
	if len(value) != 4 {
		return 0, fmt.Errorf("parameter %d: unexpected size %d", code, len(value))
	}
	return int(binary.LittleEndian.Uint32(value)), nil
}

func run() error {
	if err := zkfp.Init(); err != nil {
		return err
	}
	defer zkfp.Terminate()

	count, err := zkfp.GetDeviceCount()
	if err != nil {
		return err
	}
	if count == 0 {
		fmt.Println("No fingerprint reader found.")
		return nil
	}

	device, err := zkfp.OpenDevice(0)
	if err != nil {
		return err
	}
	defer device.Close()

	width, err := readInt(device, paramImageWidth)
	if err != nil {
		return err
	}
	height, err := readInt(device, paramImageHeight)
	if err != nil {
		return err
	}
	size, err := device.ImageSize()
	if err != nil {
		return err
	}
	fmt.Printf("Image: %dx%d, %d bytes\n", width, height, size)

	// The image is 8-bit grayscale, one byte per pixel.
	if size != width*height {
		return fmt.Errorf("image size %d does not match %dx%d", size, width, height)
	}
	buf := make([]byte, size)

	fmt.Println("Place your finger on the reader...")

	// ErrCaptureFailed just means "no finger yet", so keep polling.
	for {
		err = device.AcquireFingerprintImage(buf)
		if err == nil {
			break
		}
		if !errors.Is(err, zkfp.ErrCaptureFailed) {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}

	img := &image.Gray{
		Pix:    buf,
		Stride: width,
		Rect:   image.Rect(0, 0, width, height),
	}

	file, err := os.Create("fingerprint.png")
	if err != nil {
		return err
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		return err
	}

	fmt.Println("Saved fingerprint.png")
	return nil
}