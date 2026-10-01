//go:build windows && integration

package zkfp

import (
	"encoding/binary"
	"errors"
	"testing"
	"time"
)

// Integration tests: they need libzkfp.dll, a connected reader that no other
// program is using, and a finger to place on the sensor when prompted.
//
//	go test -tags integration -run TestIntegration -v -count=1 .

const fingerTimeout = 20 * time.Second

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

func TestIntegration(t *testing.T) {
	// 1. Init
	if !t.Run("Init", func(t *testing.T) {
		if err := Init(); err != nil {
			t.Fatalf("Init: %v", err)
		}
	}) {
		t.FailNow()
	}
	defer func() {
		if err := Terminate(); err != nil {
			t.Errorf("Terminate: %v", err)
		}
	}()

	// Init twice: the C result 1 ("already initialized") must count as success.
	t.Run("InitTwice", func(t *testing.T) {
		if err := Init(); err != nil {
			t.Fatalf("second Init: %v (the C result 1 should be treated as success)", err)
		}
	})

	// 2. GetDeviceCount
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

	// 3. OpenDevice
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
	defer dev.Close() // idempotent; runs before Terminate

	// 4. ImageSize and 5. GetParameter
	var imageSize int
	t.Run("ImageSize", func(t *testing.T) {
		size, err := dev.ImageSize()
		if err != nil {
			t.Fatalf("ImageSize: %v", err)
		}
		if size <= 0 {
			t.Fatalf("ImageSize = %d, want > 0", size)
		}
		imageSize = size
	})

	t.Run("GetParameter", func(t *testing.T) {
		width := readInt(t, dev, 1)
		height := readInt(t, dev, 2)
		if width <= 0 || height <= 0 {
			t.Fatalf("width=%d height=%d, want both > 0", width, height)
		}
		if imageSize != 0 && width*height != imageSize {
			t.Errorf("width*height = %d, want ImageSize %d", width*height, imageSize)
		}
	})

	// 6. SetParameter (buzzer on, then off; you should hear a short beep)
	t.Run("SetParameter", func(t *testing.T) {
		if err := dev.SetParameter(102, []byte{1, 0, 0, 0}); err != nil {
			t.Fatalf("green LED on: %v", err)
		}
		time.Sleep(300 * time.Millisecond)
		if err := dev.SetParameter(102, []byte{0, 0, 0, 0}); err != nil {
			t.Fatalf("green LED off: %v", err)
		}
	})

	// 7. AcquireFingerprintImage
	t.Run("AcquireFingerprintImage", func(t *testing.T) {
		if imageSize <= 0 {
			t.Skip("ImageSize failed earlier")
		}
		buf := make([]byte, imageSize)
		waitForFinger(t, func() error { return dev.AcquireFingerprintImage(buf) })
		if !notBlank(buf) {
			t.Error("captured image is all zero bytes")
		}
	})

	// 8. AcquireFingerprint
	t.Run("AcquireFingerprint", func(t *testing.T) {
		if imageSize <= 0 {
			t.Skip("ImageSize failed earlier")
		}
		buf := make([]byte, imageSize)
		var template []byte
		waitForFinger(t, func() error {
			var err error
			template, err = dev.AcquireFingerprint(buf)
			return err
		})
		if len(template) == 0 || len(template) > maxTemplateSize {
			t.Fatalf("template length = %d, want 1..%d", len(template), maxTemplateSize)
		}
		if !notBlank(template) {
			t.Error("template is all zero bytes")
		}
		if !notBlank(buf) {
			t.Error("image written by AcquireFingerprint is all zero bytes")
		}
		t.Logf("template: %d bytes", len(template))
	})

	// 9. Close (and closed-device behaviour)
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

	// The device must be usable again after Close.
	t.Run("ReopenAfterClose", func(t *testing.T) {
		d, err := OpenDevice(0)
		if err != nil {
			t.Fatalf("OpenDevice after Close: %v", err)
		}
		if err := d.Close(); err != nil {
			t.Errorf("Close after reopen: %v", err)
		}
	})

	// 10. Terminate runs in the deferred function above; a failure is
	// reported there as an error of TestIntegration itself.
}