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
