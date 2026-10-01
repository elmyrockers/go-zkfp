package zkfp

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// ZKFPM_GetParameters parameter for fingerprint image size.
	paramImageSize = 106

	// Maximum fingerprint template size defined by the ZKFinger SDK.
	maxTemplateSize = 2048
)

var (
	dll = windows.NewLazyDLL("libzkfp.dll")

	procInit                  = dll.NewProc("ZKFPM_Init")
	procTerminate             = dll.NewProc("ZKFPM_Terminate")
	procGetDeviceCount        = dll.NewProc("ZKFPM_GetDeviceCount")
	procOpenDevice            = dll.NewProc("ZKFPM_OpenDevice")
	procCloseDevice           = dll.NewProc("ZKFPM_CloseDevice")
	procGetParameters         = dll.NewProc("ZKFPM_GetParameters")
	procAcquireFingerprint    = dll.NewProc("ZKFPM_AcquireFingerprint")
	procAcquireFingerprintImg = dll.NewProc("ZKFPM_AcquireFingerprintImage")
)

// Device represents an opened fingerprint device.
type Device struct {
	handle windows.Handle
}

// Init initializes the ZKFinger SDK.
func Init() error {
	if err := dll.Load(); err != nil {
		return err
	}

	r1, _, _ := procInit.Call()
	return check(int32(r1))
}

// Terminate terminates the ZKFinger SDK.
func Terminate() error {
	r1, _, _ := procTerminate.Call()

	return check(int32(r1))
}

// GetDeviceCount returns the number of connected fingerprint devices.
func GetDeviceCount() (int, error) {
	r1, _, _ := procGetDeviceCount.Call()

	result := int32(r1)

	if result < 0 {
		return 0, Error(result)
	}

	return int(result), nil
}

// OpenDevice opens a fingerprint device by its zero-based index.
func OpenDevice(index int) (*Device, error) {
	r1, _, _ := procOpenDevice.Call(
		uintptr(index),
	)

	if r1 == 0 {
		return nil, ErrOpenDevice
	}

	return &Device{
		handle: windows.Handle(r1),
	}, nil
}

// Close closes the fingerprint device.
func (d *Device) Close() error {
	if d == nil || d.handle == 0 {
		return nil
	}

	r1, _, _ := procCloseDevice.Call(
		uintptr(d.handle),
	)

	if err := check(int32(r1)); err != nil {
		return err
	}

	d.handle = 0

	return nil
}

// ImageSize returns the required fingerprint image buffer size.
func (d *Device) ImageSize() (int, error) {
	if d == nil || d.handle == 0 {
		return 0, ErrInvalidHandle
	}

	var size uint32

	r1, _, _ := procGetParameters.Call(
		uintptr(d.handle),
		uintptr(paramImageSize),
		uintptr(unsafe.Pointer(&size)),
		uintptr(unsafe.Sizeof(size)),
	)

	if err := check(int32(r1)); err != nil {
		return 0, err
	}

	return int(size), nil
}

// AcquireFingerprint captures a fingerprint image and extracts its template.
//
// The fingerprint image is written into the supplied image buffer.
// Use ImageSize to determine the required image buffer size.
func (d *Device) AcquireFingerprint(image []byte) ([]byte, error) {
	if d == nil || d.handle == 0 {
		return nil, ErrInvalidHandle
	}

	if len(image) == 0 {
		return nil, ErrInvalidParam
	}

	template := make([]byte, maxTemplateSize)
	templateSize := uint32(len(template))

	r1, _, _ := procAcquireFingerprint.Call(
		uintptr(d.handle),
		uintptr(unsafe.Pointer(&image[0])),
		uintptr(len(image)),
		uintptr(unsafe.Pointer(&template[0])),
		uintptr(unsafe.Pointer(&templateSize)),
	)

	if err := check(int32(r1)); err != nil {
		return nil, err
	}

	return template[:templateSize], nil
}

// AcquireFingerprintImage captures a fingerprint image.
//
// The fingerprint image is written into the supplied image buffer.
// Use ImageSize to determine the required image buffer size.
func (d *Device) AcquireFingerprintImage(image []byte) error {
	if d == nil || d.handle == 0 {
		return ErrInvalidHandle
	}

	if len(image) == 0 {
		return ErrInvalidParam
	}

	r1, _, _ := procAcquireFingerprintImg.Call(
		uintptr(d.handle),
		uintptr(unsafe.Pointer(&image[0])),
		uintptr(len(image)),
	)

	return check(int32(r1))
}