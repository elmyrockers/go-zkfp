# go-zkfp
[![Go Reference](https://pkg.go.dev/badge/github.com/elmyrockers/go-zkfp.svg)](https://pkg.go.dev/github.com/elmyrockers/go-zkfp)
[![Go Version](https://img.shields.io/badge/go1.27+-00ADD8?logo=go&logoColor=white)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Integration Test](https://github.com/elmyrockers/go-zkfp/actions/workflows/integration-test.yml/badge.svg?branch=main)](https://github.com/elmyrockers/go-zkfp/actions/workflows/integration-test.yml)

Pure Go, low-level binding for ZKTeco fingerprint readers via the ZKFinger driver's `libzkfp.dll` on Windows.

<p align="center"><img src="/img/ZK9500.jpg" width="300px"></p>

## Installation

### Requirements

- Windows XP or later
- Go 1.27.1 or later
- A Go build architecture (`GOARCH=386` or `amd64`) that matches the architecture of `libzkfp.dll`

cgo and a C compiler are **not** required.

### Steps

1. Install the ZKFinger driver, which you can download from [zkteco.com](https://www.zkteco.com).

2. Add the package to your module:

```bash
   go get github.com/elmyrockers/go-zkfp
```

3. Import it in your code:

```go
   import "github.com/elmyrockers/go-zkfp"
```

## API Reference

### SDK

| Go API | Description |
| --- | --- |
| `Init() error` | Initializes the ZKFinger SDK and loads `libzkfp.dll`. Calling it again when the SDK is already initialized is treated as success. |
| `Terminate() error` | Terminates the ZKFinger SDK and releases SDK resources. Close all devices and free all DBs first. |
| `GetDeviceCount() (int, error)` | Returns the number of connected fingerprint devices. |

### Device

| Go API | Parameters | Description |
| --- | --- | --- |
| `OpenDevice(index int) (*Device, error)` | `index` — zero-based device index | Opens a fingerprint device and returns a device handle. |
| `(*Device).Close() error` | None | Closes the device. Safe to call more than once. |
| `(*Device).ImageSize() (int, error)` | None | Returns the required fingerprint image buffer size in bytes. |
| `(*Device).AcquireFingerprint(image []byte) ([]byte, error)` | `image` — destination buffer, sized with `ImageSize` | Captures a fingerprint image into `image` and returns the extracted template. Returns `ErrCaptureFailed` while no finger is on the sensor. |
| `(*Device).AcquireFingerprintImage(image []byte) error` | `image` — destination buffer, sized with `ImageSize` | Captures a fingerprint image into `image` without extracting a template. Returns `ErrCaptureFailed` while no finger is on the sensor. |
| `(*Device).GetParameter(code int, size int) ([]byte, error)` | `code` — SDK parameter code; `size` — value buffer size in bytes | Reads a device parameter and returns its raw value, cut to the size reported by the SDK. |
| `(*Device).SetParameter(code int, value []byte) error` | `code` — SDK parameter code; `value` — raw parameter value | Writes a raw value to a device parameter. Integer parameters are 4-byte little-endian. |

### DB (algorithm cache)

| Go API | Parameters | Description |
| --- | --- | --- |
| `DBInit() (*DB, error)` | None | Creates an algorithm cache that holds registered templates. |
| `(*DB).Free() error` | None | Releases the algorithm cache. Safe to call more than once. |
| `(*DB).Merge(t1, t2, t3 []byte) ([]byte, error)` | `t1`, `t2`, `t3` — three templates captured from the same finger | Merges three templates into one registration template. |
| `(*DB).Add(fid uint, template []byte) error` | `fid` — fingerprint ID; `template` — registration template | Adds a template to the cache under the given ID. |
| `(*DB).Del(fid uint) error` | `fid` — fingerprint ID | Deletes the template with the given ID from the cache. |
| `(*DB).Clear() error` | None | Removes all templates from the cache. |
| `(*DB).Count() (int, error)` | None | Returns the number of templates in the cache. |
| `(*DB).Identify(template []byte) (fid uint, score uint, err error)` | `template` — template to look up | Performs a 1:N comparison against the cache and returns the best matching ID and its score. |
| `(*DB).Match(t1, t2 []byte) (int, error)` | `t1`, `t2` — templates to compare | Performs a 1:1 comparison of two templates and returns the score. |
| `(*DB).ExtractFromImage(path string, dpi uint) ([]byte, error)` | `path` — BMP or JPG file; `dpi` — image resolution (e.g. 500) | Extracts a template from an image file. The path must be ASCII, because the SDK expects an ANSI string. |

### Notes

- Templates are at most 2048 bytes (`maxTemplateSize`).
- Typical order of use: `Init` → `OpenDevice` → `DBInit` → ... → `DB.Free` → `Device.Close` → `Terminate`.
- All methods return `ErrInvalidHandle` on a closed or nil `Device`/`DB`, and `ErrInvalidParam` for empty buffers or paths.

### Parameter Codes

`(*Device).GetParameter` and `(*Device).SetParameter` take an SDK parameter code. Integer parameters are 4-byte little-endian values.

| Constant              |   Code | Access     | Type         | Description                                                                |
| --------------------- | -----: | ---------- | ------------ | -------------------------------------------------------------------------- |
| `ParamImageWidth`     |    `1` | Read       | Int          | Fingerprint image width in pixels.                                         |
| `ParamImageHeight`    |    `2` | Read       | Int          | Fingerprint image height in pixels.                                        |
| `ParamImageDPI`       |    `3` | Read/Write | Int          | Image DPI (`LIVEID20R` only; 750/1000 recommended for children).           |
| `ParamWhiteLED`       |  `101` | Write      | Int          | White LED on the reader (`1` = blink, `0` = off).                          |
| `ParamGreenLED`       |  `102` | Write      | Int          | Green LED on the reader (`1` = blink, `0` = off).                          |
| `ParamRedLED`         |  `103` | Write      | Int          | Red LED on the reader (`1` = blink, `0` = off).                            |
| `ParamBuzzer`         |  `104` | Write      | Int          | Buzzer (`1` = start buzzing, `0` = off).                                   |
| `ParamImageSize`      |  `106` | Read       | Int          | Fingerprint image buffer size in bytes.                                    |
| `ParamVIDPID`         | `1015` | Read       | 4-byte array | USB VID and PID (first two bytes are the VID, last two bytes the PID).     |
| `ParamVendorInfo`     | `1101` | Read       | String       | Vendor information.                                                        |
| `ParamProductName`    | `1102` | Read       | String       | Product name.                                                              |
| `ParamSerialNumber`   | `1103` | Read       | String       | Device serial number (SN).                                                 |
| `ParamAntiFake`       | `2002` | Read/Write | Int          | Anti-fake function (`LIVEID20R` only; `1` = enable, `0` = disable).        |
| `ParamFakeStatus`     | `2004` | Read       | Int          | True if the lower five bits are all 1's (`value&31 == 31`).                |
| `ParamTemplateFormat` | `10001` | Write     | Int          | Template format (ISO/ANSI readers only; `0` = ANSI378, `1` = ISO 19794-2). |

#### Reading a parameter

```go
value, err := dev.GetParameter(zkfp.ParamImageWidth, 4)
if err != nil {
    return err
}

width := int(binary.LittleEndian.Uint32(value))
```

#### Writing a parameter

```go
// Turn the green LED on.
if err := dev.SetParameter(zkfp.ParamGreenLED, []byte{1, 0, 0, 0}); err != nil {
    return err
}
```

`ImageSize()` already wraps `ParamImageSize`, so you rarely need to read it yourself.

### Error Handling

Every function that talks to the SDK returns a `zkfp.Error`, an `int32` that holds the original `libzkfp.dll` result code.

```go
type Error int32
```

#### Quick start

Most calls only need a normal `err != nil` check:

```go
if err := zkfp.Init(); err != nil {
    log.Fatal(err)
}
```

To react to a specific error, compare against its constant with `errors.Is`. This is the usual way to wait for a finger, because `ErrCaptureFailed` only means nothing is on the sensor yet:

```go
for {
    err := dev.AcquireFingerprintImage(buf)
    if err == nil {
        break
    }
    if !errors.Is(err, zkfp.ErrCaptureFailed) {
        return err
    }
    time.Sleep(100 * time.Millisecond)
}
```

To read the raw SDK result code, use `errors.As`:

```go
var zkErr zkfp.Error

if errors.As(err, &zkErr) {
    fmt.Println("ZKFinger error code:", int32(zkErr))
}
```

#### Error messages

Each message contains a description and the SDK code:

```text
zkfp: failed to initialize the algorithm library (-1)
```

A non-zero result code that has no constant is still returned as a `zkfp.Error`:

```text
zkfp: unknown error (<code>)
```

#### Error constants

| Error                 |  Code | Description                                                         |
| --------------------- | ----: | ------------------------------------------------------------------- |
| `nil`                 |   `1` | The SDK is already initialized (`Init` treats this as success).     |
| `nil`                 |   `0` | The operation succeeded.                                            |
| `ErrInitLib`          |  `-1` | Failed to initialize the algorithm library.                         |
| `ErrInitCapture`      |  `-2` | Failed to initialize the capture library.                           |
| `ErrNoDevice`         |  `-3` | No fingerprint device is connected.                                 |
| `ErrNotSupported`     |  `-4` | The requested operation is not supported.                           |
| `ErrInvalidParam`     |  `-5` | An invalid parameter was supplied.                                  |
| `ErrOpenDevice`       |  `-6` | Failed to open the fingerprint device.                              |
| `ErrInvalidHandle`    |  `-7` | The device handle is invalid.                                       |
| `ErrCaptureFailed`    |  `-8` | Fingerprint capture failed.                                         |
| `ErrExtractFailed`    |  `-9` | Fingerprint template extraction failed.                             |
| `ErrAbort`            | `-10` | The operation was suspended.                                        |
| `ErrNoMemory`         | `-11` | Insufficient memory.                                                |
| `ErrBusy`             | `-12` | The fingerprint device is busy.                                     |
| `ErrAddFailed`        | `-13` | Failed to add a fingerprint template to memory.                     |
| `ErrDeleteFailed`     | `-14` | Failed to delete a fingerprint template.                            |
| `ErrOperationFailed`  | `-17` | The operation failed.                                               |
| `ErrCaptureCancelled` | `-18` | Fingerprint capture was cancelled.                                  |
| `ErrMatchFailed`      | `-20` | Fingerprint comparison failed.                                      |
| `ErrMergeFailed`      | `-22` | Failed to combine registered fingerprint templates.                 |
| `ErrOpenFile`         | `-23` | Failed to open a file.                                              |
| `ErrImageProcess`     | `-24` | Fingerprint image processing failed.                                |