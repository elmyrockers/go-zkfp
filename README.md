# go-zkfp
Pure Go, low-level binding for ZKTeco fingerprint readers via the ZKFinger driver's `libzkfp.dll` on Windows.

<p align="center"><img src="/img/ZK9500.jpg" width="300px"></p>


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

| Constant              | Code  | Access     | Description                                  |
| --------------------- | ----: | ---------- | -------------------------------------------- |
| `ParamImageWidth`     |   `1` | Read       | Fingerprint image width in pixels.           |
| `ParamImageHeight`    |   `2` | Read       | Fingerprint image height in pixels.          |
| `ParamGreenLED`       | `102` | Write      | Green LED on the reader (`1` = on, `0` = off). |
| `ParamImageSize`      | `106` | Read       | Fingerprint image buffer size in bytes.      |

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