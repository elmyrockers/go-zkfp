package main

import (
	"errors"
	"fmt"
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

	image := make([]byte, imageSize)

	var fid int
	fmt.Print("Enter fingerprint number: ")

	if _, err := fmt.Scan(&fid); err != nil {
		log.Fatal(err)
	}

	path := fmt.Sprintf("../data/fingerprint-%d.dat", fid)

	template, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Place your finger on the reader...")

	var capturedTemplate []byte

	for {
		capturedTemplate, err = device.AcquireFingerprint(image)
		if err == nil {
			break
		}

		if !errors.Is(err, zkfp.ErrCaptureFailed) {
			log.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("Fingerprint template: %d bytes\n", len(capturedTemplate))

	db, err := zkfp.DBInit()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Free()

	score, err := db.Match(capturedTemplate, template)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Match score: %d\n", score)
}