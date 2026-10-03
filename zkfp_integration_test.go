//go:build windows && integration

package zkfp

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Integration tests: they need libzkfp.dll, a connected reader that no other
// program is using, and a finger to place on the sensor when prompted.
//
//	go test -tags integration -run TestIntegration -v -count=1 .

const (
	fingerTimeout = 20 * time.Second

	// Device parameter codes used by the tests.
	testParamImageWidth  = 1
	testParamImageHeight = 2
	testParamGreenLED    = 102

	// 500 DPI expressed in pixels per meter (500 / 0.0254).
	pixelsPerMeter = 19685
)

// waitForFinger keeps calling capture until it succeeds. ErrCaptureFailed
// only means "no finger yet".
func waitForFinger(t *testing.T, capture func() error) {
	t.Helper()
	t.Log(">>> place a finger on the reader")

	deadline := time.Now().Add(fingerTimeout)
	for {
		err := capture()
		if err == nil {
			return
		}
		if !errors.Is(err, ErrCaptureFailed) {
			t.Fatalf("capture failed: %v", err)
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out after %s waiting for a finger", fingerTimeout)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func readInt(t *testing.T, d *Device, code int) int {
	t.Helper()

	value, err := d.GetParameter(code, 4)
	if err != nil {
		t.Fatalf("GetParameter(%d): %v", code, err)
	}
	if len(value) != 4 {
		t.Fatalf("GetParameter(%d) returned %d bytes, want 4", code, len(value))
	}

	return int(binary.LittleEndian.Uint32(value))
}

func notBlank(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return true
		}
	}
	return false
}

// checkTemplate fails the test if tmpl is not a plausible template.
func checkTemplate(t *testing.T, name string, tmpl []byte) {
	t.Helper()

	if len(tmpl) == 0 || len(tmpl) > maxTemplateSize {
		t.Fatalf("%s length = %d, want 1..%d", name, len(tmpl), maxTemplateSize)
	}
	if !notBlank(tmpl) {
		t.Fatalf("%s is all zero bytes", name)
	}
}

// writeGrayBMP writes an 8-bit grayscale BMP file.
//
// The fingerprint image returned by the SDK is an 8-bit grayscale image.
// BMP rows are padded to a 4-byte boundary and stored bottom-up.
func writeGrayBMP(path string, pixels []byte, width, height int) error {
	if width <= 0 || height <= 0 {
		return errors.New("invalid image dimensions")
	}
	if len(pixels) != width*height {
		return errors.New("image size does not match dimensions")
	}

	rowSize := (width + 3) &^ 3
	pixelSize := rowSize * height

	const (
		fileHeaderSize = 14
		infoHeaderSize = 40
		paletteSize    = 256 * 4
	)

	pixelOffset := fileHeaderSize + infoHeaderSize + paletteSize
	fileSize := pixelOffset + pixelSize

	data := make([]byte, fileSize)

	// BITMAPFILEHEADER
	data[0] = 'B'
	data[1] = 'M'
	binary.LittleEndian.PutUint32(data[2:6], uint32(fileSize))
	binary.LittleEndian.PutUint32(data[10:14], uint32(pixelOffset))

	// BITMAPINFOHEADER
	binary.LittleEndian.PutUint32(data[14:18], infoHeaderSize)
	binary.LittleEndian.PutUint32(data[18:22], uint32(width))
	binary.LittleEndian.PutUint32(data[22:26], uint32(height))
	binary.LittleEndian.PutUint16(data[26:28], 1) // planes
	binary.LittleEndian.PutUint16(data[28:30], 8) // bits per pixel
	binary.LittleEndian.PutUint32(data[30:34], 0) // BI_RGB
	binary.LittleEndian.PutUint32(data[34:38], uint32(pixelSize))
	binary.LittleEndian.PutUint32(data[38:42], pixelsPerMeter)
	binary.LittleEndian.PutUint32(data[42:46], pixelsPerMeter)
	binary.LittleEndian.PutUint32(data[46:50], 256) // colors used

	// Grayscale palette.
	paletteOffset := fileHeaderSize + infoHeaderSize
	for i := 0; i < 256; i++ {
		offset := paletteOffset + i*4
		data[offset] = byte(i)
		data[offset+1] = byte(i)
		data[offset+2] = byte(i)
		data[offset+3] = 0
	}

	// Pixel data is stored bottom-up.
	for y := 0; y < height; y++ {
		srcStart := y * width
		dstStart := pixelOffset + (height-1-y)*rowSize
		copy(data[dstStart:dstStart+width], pixels[srcStart:srcStart+width])
	}

	return os.WriteFile(path, data, 0600)
}

func TestIntegration(t *testing.T) {
	// -------------------------------------------------------------------------
	// SDK init
	// -------------------------------------------------------------------------
	if !t.Run("Init", func(t *testing.T) {
		err := Init()
		if errors.Is(err, ErrLoadLibrary) {
			t.Fatalf("Init: %v (is the ZKTeco driver installed, and does GOARCH match libzkfp.dll?)", err)
		}
		if err != nil {
			t.Fatalf("Init: %v", err)
		}
	}) {
		t.FailNow()
	}

	// Deferred calls run last-in-first-out. Terminate is registered first so
	// it runs after DB.Free and Device.Close.
	defer func() {
		if err := Terminate(); err != nil {
			t.Errorf("Terminate: %v", err)
		}
	}()

	// Init twice: only 0 is success, so the C result 1 ("already initialized")
	// must come back as ErrAlreadyInit.
	t.Run("InitTwice", func(t *testing.T) {
		if err := Init(); !errors.Is(err, ErrAlreadyInit) {
			t.Fatalf("second Init: got %v, want ErrAlreadyInit", err)
		}
	})
	
	// -------------------------------------------------------------------------
	// Device
	// -------------------------------------------------------------------------

	if !t.Run("GetDeviceCount", func(t *testing.T) {
		n, err := GetDeviceCount()
		if err != nil {
			t.Fatalf("GetDeviceCount: %v", err)
		}
		if n < 1 {
			t.Fatalf("GetDeviceCount = %d, want at least 1 (is the reader plugged in?)", n)
		}
	}) {
		t.FailNow()
	}

	t.Run("OpenDeviceBadIndex", func(t *testing.T) {
		d, err := OpenDevice(99)
		if err == nil {
			d.Close()
			t.Fatal("OpenDevice(99) succeeded, want an error")
		}
	})

	var dev *Device

	if !t.Run("OpenDevice", func(t *testing.T) {
		d, err := OpenDevice(0)
		if err != nil {
			t.Fatalf("OpenDevice(0): %v (close the demo app and any other program using the reader)", err)
		}
		dev = d
	}) {
		t.FailNow()
	}

	// Safety net in case a later step aborts the test before the explicit
	// Close subtest. Close is idempotent, so this is harmless afterwards.
	defer dev.Close()

	var imageSize int

	if !t.Run("ImageSize", func(t *testing.T) {
		size, err := dev.ImageSize()
		if err != nil {
			t.Fatalf("ImageSize: %v", err)
		}
		if size <= 0 {
			t.Fatalf("ImageSize = %d, want > 0", size)
		}
		imageSize = size
	}) {
		t.FailNow()
	}

	var imageWidth, imageHeight int

	t.Run("GetParameter", func(t *testing.T) {
		imageWidth = readInt(t, dev, testParamImageWidth)
		imageHeight = readInt(t, dev, testParamImageHeight)

		if imageWidth <= 0 || imageHeight <= 0 {
			t.Fatalf("width=%d height=%d, want both > 0", imageWidth, imageHeight)
		}
		if imageWidth*imageHeight != imageSize {
			t.Errorf("width*height = %d, want ImageSize %d", imageWidth*imageHeight, imageSize)
		}
	})

	t.Run("SetParameter", func(t *testing.T) {
		if err := dev.SetParameter(testParamGreenLED, []byte{1, 0, 0, 0}); err != nil {
			t.Fatalf("green LED on: %v", err)
		}

		time.Sleep(300 * time.Millisecond)

		if err := dev.SetParameter(testParamGreenLED, []byte{0, 0, 0, 0}); err != nil {
			t.Fatalf("green LED off: %v", err)
		}
	})

	// The BMP is written in the parent test so it outlives the subtest.
	bmpPath := filepath.Join(t.TempDir(), "capture.bmp")
	bmpWritten := false

	t.Run("AcquireFingerprintImage", func(t *testing.T) {
		buf := make([]byte, imageSize)

		waitForFinger(t, func() error {
			return dev.AcquireFingerprintImage(buf)
		})

		if !notBlank(buf) {
			t.Fatal("captured image is all zero bytes")
		}

		// Save the capture as a BMP so ExtractFromImage can use it later.
		if imageWidth > 0 && imageHeight > 0 && imageWidth*imageHeight == len(buf) {
			if err := writeGrayBMP(bmpPath, buf, imageWidth, imageHeight); err != nil {
				t.Errorf("writeGrayBMP: %v", err)
				return
			}
			bmpWritten = true
		}
	})

	t.Run("AcquireFingerprint", func(t *testing.T) {
		buf := make([]byte, imageSize)

		var template []byte

		waitForFinger(t, func() error {
			var err error
			template, err = dev.AcquireFingerprint(buf)
			return err
		})

		checkTemplate(t, "template", template)

		if !notBlank(buf) {
			t.Error("image written by AcquireFingerprint is all zero bytes")
		}

		t.Logf("template: %d bytes", len(template))
	})

	// -------------------------------------------------------------------------
	// DB
	// -------------------------------------------------------------------------

	db, err := DBInit()
	if err != nil {
		t.Fatalf("DBInit: %v", err)
	}

	// Free is deliberately called after all DB operations below.
	defer func() {
		if err := db.Free(); err != nil {
			t.Errorf("DB.Free: %v", err)
		}
	}()

	// Capture three templates from the same finger for DB.Merge.
	var templates [3][]byte

	for i := range templates {
		t.Logf(">>> place the same finger for template %d/3", i+1)

		buf := make([]byte, imageSize)

		waitForFinger(t, func() error {
			var err error
			templates[i], err = dev.AcquireFingerprint(buf)
			return err
		})

		checkTemplate(t, "template "+string(rune('1'+i)), templates[i])
		t.Logf("template %d: %d bytes", i+1, len(templates[i]))

		if i < len(templates)-1 {
			t.Log(">>> remove your finger")
			time.Sleep(2 * time.Second)
		}
	}

	mergedTemplate, err := db.Merge(templates[0], templates[1], templates[2])
	if err != nil {
		t.Fatalf("DB.Merge: %v", err)
	}
	checkTemplate(t, "merged template", mergedTemplate)
	t.Logf("merged template: %d bytes", len(mergedTemplate))

	const fid uint = 1

	if err := db.Add(fid, mergedTemplate); err != nil {
		t.Fatalf("DB.Add: %v", err)
	}

	count, err := db.Count()
	if err != nil {
		t.Fatalf("DB.Count: %v", err)
	}
	if count != 1 {
		t.Fatalf("DB.Count = %d, want 1", count)
	}

	gotFID, identifyScore, err := db.Identify(templates[0])
	if err != nil {
		t.Fatalf("DB.Identify: %v", err)
	}
	if gotFID != fid {
		t.Fatalf("DB.Identify FID = %d, want %d", gotFID, fid)
	}
	if identifyScore == 0 {
		t.Fatal("DB.Identify score = 0, want > 0")
	}
	t.Logf("identified FID: %d, score: %d", gotFID, identifyScore)

	matchScore, err := db.Match(templates[0], mergedTemplate)
	if err != nil {
		t.Fatalf("DB.Match: %v", err)
	}
	t.Logf("match score: %d", matchScore)
	if matchScore <= 0 {
		t.Fatalf("DB.Match score = %d, want > 0", matchScore)
	}

	if err := db.Del(fid); err != nil {
		t.Fatalf("DB.Del: %v", err)
	}

	count, err = db.Count()
	if err != nil {
		t.Fatalf("DB.Count after Del: %v", err)
	}
	if count != 0 {
		t.Fatalf("DB.Count after Del = %d, want 0", count)
	}

	// Add again so DB.Clear can be tested.
	if err := db.Add(fid, mergedTemplate); err != nil {
		t.Fatalf("DB.Add for Clear: %v", err)
	}

	if err := db.Clear(); err != nil {
		t.Fatalf("DB.Clear: %v", err)
	}

	count, err = db.Count()
	if err != nil {
		t.Fatalf("DB.Count after Clear: %v", err)
	}
	if count != 0 {
		t.Fatalf("DB.Count after Clear = %d, want 0", count)
	}

	// Extract a template from the BMP captured earlier. The SDK accepts BMP
	// or JPG (not PNG), and expects an ANSI path, so a temp dir under a
	// non-ASCII user name may fail here.
	t.Run("ExtractFromImage", func(t *testing.T) {
		if !bmpWritten {
			t.Skip("no BMP was captured earlier")
		}

		tmpl, err := db.ExtractFromImage(bmpPath, 500)
		if err != nil {
			t.Fatalf("ExtractFromImage: %v", err)
		}

		checkTemplate(t, "extracted template", tmpl)
		t.Logf("extracted template: %d bytes", len(tmpl))
	})

	// -------------------------------------------------------------------------
	// Cleanup
	// -------------------------------------------------------------------------

	t.Run("Close", func(t *testing.T) {
		if err := dev.Close(); err != nil {
			t.Fatalf("Close: %v", err)
		}

		if err := dev.Close(); err != nil {
			t.Errorf("second Close: %v, want nil", err)
		}

		if _, err := dev.ImageSize(); !errors.Is(err, ErrInvalidHandle) {
			t.Errorf("ImageSize after Close: %v, want ErrInvalidHandle", err)
		}
	})
}