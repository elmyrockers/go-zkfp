package zkfp

import "fmt"

// Error is a non-zero result code returned by libzkfp.dll.
type Error int32

const (
	ErrInitLib          Error = -1  // failed to initialize the algorithm library
	ErrInitCapture      Error = -2  // failed to initialize the capture library
	ErrNoDevice         Error = -3  // no device connected
	ErrNotSupported     Error = -4  // not supported by the interface
	ErrInvalidParam     Error = -5  // invalid parameter
	ErrOpenDevice       Error = -6  // failed to start the device
	ErrInvalidHandle    Error = -7  // invalid handle
	ErrCaptureFailed    Error = -8  // failed to capture the image (e.g. no finger)
	ErrExtractFailed    Error = -9  // failed to extract the fingerprint template
	ErrAbort            Error = -10 // suspension operation
	ErrNoMemory         Error = -11 // insufficient memory
	ErrBusy             Error = -12 // device is busy capturing
	ErrAddFailed        Error = -13 // failed to add template to memory
	ErrDeleteFailed     Error = -14 // failed to delete template
	ErrOperationFailed  Error = -17 // operation failed (other error)
	ErrCaptureCancelled Error = -18 // capture cancelled
	ErrMatchFailed      Error = -20 // fingerprint comparison failed
	ErrMergeFailed      Error = -22 // failed to combine registered templates
	ErrOpenFile         Error = -23 // opening the file failed
	ErrImageProcess     Error = -24 // image processing failed
)

var errText = map[Error]string{
	ErrInitLib:          "failed to initialize the algorithm library",
	ErrInitCapture:      "failed to initialize the capture library",
	ErrNoDevice:         "no device connected",
	ErrNotSupported:     "not supported by the interface",
	ErrInvalidParam:     "invalid parameter",
	ErrOpenDevice:       "failed to start the device",
	ErrInvalidHandle:    "invalid handle",
	ErrCaptureFailed:    "failed to capture the image",
	ErrExtractFailed:    "failed to extract the fingerprint template",
	ErrAbort:            "suspension operation",
	ErrNoMemory:         "insufficient memory",
	ErrBusy:             "device is busy",
	ErrAddFailed:        "failed to add the fingerprint template to the memory",
	ErrDeleteFailed:     "failed to delete the fingerprint template",
	ErrOperationFailed:  "operation failed",
	ErrCaptureCancelled: "capture cancelled",
	ErrMatchFailed:      "fingerprint comparison failed",
	ErrMergeFailed:      "failed to combine registered fingerprint templates",
	ErrOpenFile:         "opening the file failed",
	ErrImageProcess:     "image processing failed",
}

func (e Error) Error() string {
	if s, ok := errText[e]; ok {
		return fmt.Sprintf("zkfp: %s (%d)", s, int32(e))
	}
	return fmt.Sprintf("zkfp: unknown error (%d)", int32(e))
}

// check converts a C result code into a Go error (0 means success).
func check(code int32) error {
	if code == 0 {
		return nil
	}
	return Error(code)
}