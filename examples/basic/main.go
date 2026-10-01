package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/elmyrockers/go-zkfp"
)

func main() {
	if err := zkfp.Init(); err != nil {
		log.Fatal(err)
	}
	defer zkfp.Terminate()

	count, err := zkfp.GetDeviceCount()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Device count: %d\n", count)

	if count == 0 {
		fmt.Println("No fingerprint reader found.")
		return
	}

	device, err := zkfp.OpenDevice(0)
	if err != nil {
		log.Fatal(err)
	}
	defer device.Close()

	for _, code := range []int{101, 102, 103, 104} {
		for _, v := range [][]byte{{1, 0, 0, 0}, {1}} {
			err := device.SetParameter(code, v)
			fmt.Println("code", code, "len", len(v), "->", err)
			time.Sleep(300 * time.Millisecond)
			device.SetParameter(code, make([]byte, len(v))) // turn off
		}
	}

	imageSize, err := device.ImageSize()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Image size: %d bytes\n", imageSize)

	image := make([]byte, imageSize)

	//-----------------------------------------------------------------------------
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

	fmt.Printf("Fingerprint image: %d bytes\n", len(image))
	fmt.Printf("Fingerprint template: %d bytes\n", len(template))
}