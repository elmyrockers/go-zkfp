# go-zkfp
Pure Go binding for integrating ZKTeco fingerprint reader via `libzkfp.dll` on Windows.

<p align="center"><img src="/img/ZK9500.jpg" width="300px"></p>


## API Reference

| Go API                                                       | Parameters                                                                      | Return Value     | Description                                                                          |
| ------------------------------------------------------------ | ------------------------------------------------------------------------------- | ---------------- | ------------------------------------------------------------------------------------ |
| `Init() error`                                               | None                                                                            | `error`          | Initializes the ZKFinger SDK.                                                        |
| `Terminate() error`                                          | None                                                                            | `error`          | Terminates the ZKFinger SDK and releases SDK resources.                              |
| `GetDeviceCount() (int, error)`                              | None                                                                            | `int, error`     | Returns the number of connected fingerprint devices.                                 |
| `OpenDevice(index int) (*Device, error)`                     | `index int` — zero-based device index                                           | `*Device, error` | Opens a fingerprint device and returns a device handle.                              |
| `(*Device).Close() error`                                    | None                                                                            | `error`          | Closes the fingerprint device.                                                       |
| `(*Device).ImageSize() (int, error)`                         | None                                                                            | `int, error`     | Returns the required fingerprint image buffer size.                                  |
| `(*Device).AcquireFingerprint(image []byte) ([]byte, error)` | `image []byte` — destination buffer for the fingerprint image                   | `[]byte, error`  | Captures a fingerprint image and extracts its fingerprint template.                  |
| `(*Device).AcquireFingerprintImage(image []byte) error`      | `image []byte` — destination buffer for the fingerprint image                   | `error`          | Captures a fingerprint image into the supplied buffer without extracting a template. |
| `(*Device).GetParameter(code int, size int) ([]byte, error)` | `code int` — SDK parameter code; `size int` — size of the value buffer in bytes | `[]byte, error`  | Reads a device parameter and returns its raw value.                                  |
| `(*Device).SetParameter(code int, value []byte) error`       | `code int` — SDK parameter code; `value []byte` — raw parameter value           | `error`          | Writes a raw value to a device parameter.                                            |

## Error Handling

The SDK returns errors as the exported `Error` type. Each error contains the original `libzkfp.dll` result code.

```go
type Error int32
```

For example:

```go
if err := zkfp.Init(); err != nil {
    log.Fatal(err)
}
```

The following error constants are provided:

| Error                 |  Code | Description                                         |
| --------------------- | ----: | --------------------------------------------------- |
| `ErrInitLib`          |  `-1` | Failed to initialize the algorithm library.         |
| `ErrInitCapture`      |  `-2` | Failed to initialize the capture library.           |
| `ErrNoDevice`         |  `-3` | No fingerprint device is connected.                 |
| `ErrNotSupported`     |  `-4` | The requested operation is not supported.           |
| `ErrInvalidParam`     |  `-5` | An invalid parameter was supplied.                  |
| `ErrOpenDevice`       |  `-6` | Failed to open the fingerprint device.              |
| `ErrInvalidHandle`    |  `-7` | The device handle is invalid.                       |
| `ErrCaptureFailed`    |  `-8` | Fingerprint capture failed.                         |
| `ErrExtractFailed`    |  `-9` | Fingerprint template extraction failed.             |
| `ErrAbort`            | `-10` | The operation was suspended.                        |
| `ErrNoMemory`         | `-11` | Insufficient memory.                                |
| `ErrBusy`             | `-12` | The fingerprint device is busy.                     |
| `ErrAddFailed`        | `-13` | Failed to add a fingerprint template to memory.     |
| `ErrDeleteFailed`     | `-14` | Failed to delete a fingerprint template.            |
| `ErrOperationFailed`  | `-17` | The operation failed.                               |
| `ErrCaptureCancelled` | `-18` | Fingerprint capture was cancelled.                  |
| `ErrMatchFailed`      | `-20` | Fingerprint comparison failed.                      |
| `ErrMergeFailed`      | `-22` | Failed to combine registered fingerprint templates. |
| `ErrOpenFile`         | `-23` | Failed to open a file.                              |
| `ErrImageProcess`     | `-24` | Fingerprint image processing failed.                |

### Checking Specific Errors

Because `Error` implements Go's `error` interface, errors can be checked using `errors.As`:

```go
var zkErr zkfp.Error

if errors.As(err, &zkErr) {
    fmt.Println("ZKFinger error code:", int32(zkErr))
}
```

The error's string representation includes both the description and the original SDK error code:

```text
zkfp: failed to initialize the algorithm library (-1)
```

Unknown non-zero SDK result codes are also represented as `zkfp.Error` and reported as:

```text
zkfp: unknown error (<code>)
```
