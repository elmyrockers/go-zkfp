package zkfp

import (
	"fmt"
	"errors"
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
	procAcquireFingerprint    = dll.NewProc("ZKFPM_AcquireFingerprint")
	procAcquireFingerprintImg = dll.NewProc("ZKFPM_AcquireFingerprintImage")
	procGetParameters         = dll.NewProc("ZKFPM_GetParameters")
	procSetParameters         = dll.NewProc("ZKFPM_SetParameters")


	procDBInit           = dll.NewProc("ZKFPM_DBInit")
	procDBFree           = dll.NewProc("ZKFPM_DBFree")
	procDBMerge          = dll.NewProc("ZKFPM_DBMerge")
	procDBAdd            = dll.NewProc("ZKFPM_DBAdd")
	procDBDel            = dll.NewProc("ZKFPM_DBDel")
	procDBClear          = dll.NewProc("ZKFPM_DBClear")
	procDBCount          = dll.NewProc("ZKFPM_DBCount")
	procDBIdentify       = dll.NewProc("ZKFPM_DBIdentify")
	procDBMatch          = dll.NewProc("ZKFPM_DBMatch")
	procExtractFromImage = dll.NewProc("ZKFPM_ExtractFromImage")
)

// Device represents an opened fingerprint device.
type Device struct {
	handle windows.Handle
}

// DB represents an initialized ZKFinger algorithm database/cache.
type DB struct {
	handle windows.Handle
}

// Init initializes the ZKFinger SDK.
func Init() error {
	if err := dll.Load(); err != nil {
		var dllErr *windows.DLLError
		if errors.As(err, &dllErr) {
			return fmt.Errorf("%w: %v (%d)", ErrLoadLibrary, dllErr.Err, dllErr.Err)
		}
		return fmt.Errorf("%w: %v", ErrLoadLibrary, err)
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
	n := uint32(unsafe.Sizeof(size)) // [in] buffer size, [out] returned size

	r1, _, _ := procGetParameters.Call(
		uintptr(d.handle),
		uintptr(paramImageSize),
		uintptr(unsafe.Pointer(&size)),
		uintptr(unsafe.Pointer(&n)), // pointer to the length, not the length itself
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


// GetParameter reads a device parameter (ZKFPM_GetParameters).
//
// size is the number of bytes to allocate for the value, based on the
// parameter code (4 for an Int, 4 for the VID/PID array, larger for strings).
// The returned slice is cut to the size reported by the SDK.
func (d *Device) GetParameter(code int, size int) ([]byte, error) {
	if d == nil || d.handle == 0 {
		return nil, ErrInvalidHandle
	}
	if size <= 0 {
		return nil, ErrInvalidParam
	}
 
	value := make([]byte, size)
	n := uint32(size) // [in] buffer size, [out] returned size
 
	r1, _, _ := procGetParameters.Call(
		uintptr(d.handle),
		uintptr(code),
		uintptr(unsafe.Pointer(&value[0])),
		uintptr(unsafe.Pointer(&n)),
	)
 
	if err := check(int32(r1)); err != nil {
		return nil, err
	}
	if int(n) > len(value) {
		return nil, ErrOperationFailed
	}
 
	return value[:n], nil
}
 
// SetParameter writes a device parameter (ZKFPM_SetParameters).
//
// value holds the raw parameter bytes (a 4-byte little-endian integer for
// Int parameters). Its length is passed as the parameter data length.
func (d *Device) SetParameter(code int, value []byte) error {
	if d == nil || d.handle == 0 {
		return ErrInvalidHandle
	}
	if len(value) == 0 {
		return ErrInvalidParam
	}
 
	r1, _, _ := procSetParameters.Call(
		uintptr(d.handle),
		uintptr(code),
		uintptr(unsafe.Pointer(&value[0])),
		uintptr(len(value)),
	)
 
	return check(int32(r1))
}

//-----------------------------------------------------------------------------------------------------------------------------------
// DBInit creates an algorithm cache.
func DBInit() (*DB, error) {
	r1, _, _ := procDBInit.Call()

	if r1 == 0 {
		return nil, ErrOperationFailed
	}

	return &DB{
		handle: windows.Handle(r1),
	}, nil
}

// Free releases the algorithm cache.
func (db *DB) Free() error {
	if db == nil || db.handle == 0 {
		return nil
	}

	r1, _, _ := procDBFree.Call(
		uintptr(db.handle),
	)

	if err := check(int32(r1)); err != nil {
		return err
	}

	db.handle = 0

	return nil
}

// Merge combines three fingerprint templates into one registered template.
func (db *DB) Merge(
	template1 []byte,
	template2 []byte,
	template3 []byte,
) ([]byte, error) {
	if db == nil || db.handle == 0 {
		return nil, ErrInvalidHandle
	}

	if len(template1) == 0 ||
		len(template2) == 0 ||
		len(template3) == 0 {
		return nil, ErrInvalidParam
	}

	template := make([]byte, maxTemplateSize)
	templateSize := uint32(len(template))

	r1, _, _ := procDBMerge.Call(
		uintptr(db.handle),
		uintptr(unsafe.Pointer(&template1[0])),
		uintptr(unsafe.Pointer(&template2[0])),
		uintptr(unsafe.Pointer(&template3[0])),
		uintptr(unsafe.Pointer(&template[0])),
		uintptr(unsafe.Pointer(&templateSize)),
	)

	if err := check(int32(r1)); err != nil {
		return nil, err
	}

	if templateSize > uint32(len(template)) {
		return nil, ErrOperationFailed
	}

	return template[:templateSize], nil
}

// Add adds a fingerprint template to the algorithm cache.
func (db *DB) Add(fid uint, template []byte) error {
	if db == nil || db.handle == 0 {
		return ErrInvalidHandle
	}

	if len(template) == 0 {
		return ErrInvalidParam
	}

	r1, _, _ := procDBAdd.Call(
		uintptr(db.handle),
		uintptr(fid),
		uintptr(unsafe.Pointer(&template[0])),
		uintptr(len(template)),
	)

	return check(int32(r1))
}

// Del deletes a fingerprint template from the algorithm cache.
func (db *DB) Del(fid uint) error {
	if db == nil || db.handle == 0 {
		return ErrInvalidHandle
	}

	r1, _, _ := procDBDel.Call(
		uintptr(db.handle),
		uintptr(fid),
	)

	return check(int32(r1))
}

// Clear removes all fingerprint templates from the algorithm cache.
func (db *DB) Clear() error {
	if db == nil || db.handle == 0 {
		return ErrInvalidHandle
	}

	r1, _, _ := procDBClear.Call(
		uintptr(db.handle),
	)

	return check(int32(r1))
}

// Count returns the number of fingerprint templates in the algorithm cache.
func (db *DB) Count() (int, error) {
	if db == nil || db.handle == 0 {
		return 0, ErrInvalidHandle
	}

	var count uint32

	r1, _, _ := procDBCount.Call(
		uintptr(db.handle),
		uintptr(unsafe.Pointer(&count)),
	)

	if err := check(int32(r1)); err != nil {
		return 0, err
	}

	return int(count), nil
}

// Identify performs a 1:N fingerprint comparison.
func (db *DB) Identify(template []byte) (uint, uint, error) {
	if db == nil || db.handle == 0 {
		return 0, 0, ErrInvalidHandle
	}

	if len(template) == 0 {
		return 0, 0, ErrInvalidParam
	}

	var fid uint32
	var score uint32

	r1, _, _ := procDBIdentify.Call(
		uintptr(db.handle),
		uintptr(unsafe.Pointer(&template[0])),
		uintptr(len(template)),
		uintptr(unsafe.Pointer(&fid)),
		uintptr(unsafe.Pointer(&score)),
	)

	if err := check(int32(r1)); err != nil {
		return 0, 0, err
	}

	return uint(fid), uint(score), nil
}

// Match compares two fingerprint templates and returns the comparison score.
func (db *DB) Match(template1, template2 []byte) (int, error) {
	if db == nil || db.handle == 0 {
		return 0, ErrInvalidHandle
	}

	if len(template1) == 0 || len(template2) == 0 {
		return 0, ErrInvalidParam
	}

	r1, _, _ := procDBMatch.Call(
		uintptr(db.handle),
		uintptr(unsafe.Pointer(&template1[0])),
		uintptr(len(template1)),
		uintptr(unsafe.Pointer(&template2[0])),
		uintptr(len(template2)),
	)

	result := int32(r1)

	if result < 0 {
		return 0, Error(result)
	}

	return int(result), nil
}

// ExtractFromImage extracts a fingerprint template from a BMP or JPG file.
func (db *DB) ExtractFromImage(path string, dpi uint) ([]byte, error) {
	if db == nil || db.handle == 0 {
		return nil, ErrInvalidHandle
	}

	if path == "" {
		return nil, ErrInvalidParam
	}

	filePath, err := windows.BytePtrFromString(path)
	if err != nil {
		return nil, err
	}

	template := make([]byte, maxTemplateSize)
	templateSize := uint32(len(template))

	r1, _, _ := procExtractFromImage.Call(
		uintptr(db.handle),
		uintptr(unsafe.Pointer(filePath)),
		uintptr(dpi),
		uintptr(unsafe.Pointer(&template[0])),
		uintptr(unsafe.Pointer(&templateSize)),
	)

	if err := check(int32(r1)); err != nil {
		return nil, err
	}

	if templateSize > uint32(len(template)) {
		return nil, ErrOperationFailed
	}

	return template[:templateSize], nil
}