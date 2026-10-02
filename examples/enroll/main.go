package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

	image := make([]byte, imageSize)

	dataDir := "../data"

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal(err)
	}

	fid := 1

	for {
		fmt.Printf("\nPlace finger for FID %d...\n", fid)

		// Wait until a fingerprint is captured.
		var template []byte

		for {
			template, err = device.AcquireFingerprint(image)
			if err == nil {
				break
			}

			if !errors.Is(err, zkfp.ErrCaptureFailed) {
				log.Fatal(err)
			}

			time.Sleep(100 * time.Millisecond)
		}

		fmt.Printf("Fingerprint image: %d bytes\n", len(image))
		fmt.Printf("Fingerprint template: %d bytes\n", len(template))

		path := filepath.Join(
			dataDir,
			fmt.Sprintf("fingerprint-%d.dat", fid),
		)

		if err := os.WriteFile(path, template, 0644); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Fingerprint enrolled successfully (FID: %d)\n", fid)
		fmt.Printf("Template saved to %s\n", path)

		fid++

		fmt.Println("Remove your finger...")
		time.Sleep(2 * time.Second)
	}
}