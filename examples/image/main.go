package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/elmyrockers/go-zkfp"
)

func main() {
	if err := zkfp.Init(); err != nil {
		log.Fatal(err)
	}
	defer zkfp.Terminate()

	device, err := zkfp.OpenDevice(0)
	if err != nil {
		log.Fatal(err)
	}
	defer device.Close()

	imageSize, err := device.ImageSize()
	if err != nil {
		log.Fatal(err)
	}

	imageData := make([]byte, imageSize)

	fmt.Println("Place your finger on the reader...")

	for {
		err = device.AcquireFingerprintImage(imageData)
		if err == nil {
			break
		}

		if !errors.Is(err, zkfp.ErrCaptureFailed) {
			log.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Fingerprint image: %d bytes\n", len(imageData))

	// Save the original image data.
	if err := os.WriteFile("fingerprint.raw", imageData, 0644); err != nil {
		log.Fatal(err)
	}

	// The captured image is 300 pixels wide.
	// Calculate the height from the actual image buffer size.
	const width = 300

	if len(imageData)%width != 0 {
		log.Fatalf(
			"unexpected image size: %d bytes",
			len(imageData),
		)
	}

	height := len(imageData) / width

	// Create an 8-bit grayscale image from the captured pixels.
	img := image.NewGray(image.Rect(0, 0, width, height))
	copy(img.Pix, imageData)

	file, err := os.Create("fingerprint.png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Fingerprint image size: %d x %d pixels\n", width, height)
	fmt.Println("Fingerprint image saved to fingerprint.raw")
	fmt.Println("Fingerprint image saved to fingerprint.png")
}