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

	db, err := zkfp.DBInit()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Free()

	dataDir := "../data"

	files, err := filepath.Glob(filepath.Join(dataDir, "fingerprint-*.dat"))
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		log.Fatal("no enrolled fingerprints found")
	}

	for _, path := range files {
		var fid uint
		if _, err := fmt.Sscanf(filepath.Base(path), "fingerprint-%d.dat", &fid); err != nil {
			continue
		}

		template, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		if err := db.Add(fid, template); err != nil {
			log.Fatal(err)
		}
	}

	count, err := db.Count()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Enrolled fingerprints: %d\n", count)
	fmt.Println("Place your finger on the reader...")

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

	fmt.Printf("Fingerprint template: %d bytes\n", len(template))

	fid, score, err := db.Identify(template)
	if err != nil {
		if errors.Is(err, zkfp.ErrMatchFailed) {
			fmt.Println("No matching fingerprint found.")
			return
		}

		log.Fatal(err)
	}

	fmt.Printf("Matched fingerprint: FID %d\n", fid)
	fmt.Printf("Match score: %d\n", score)
}