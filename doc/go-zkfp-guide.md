# go-zkfp Development Guide

**Go Binding for ZKFinger Reader SDK (`libzkfp.dll`)**

Repository: [github.com/elmyrockers/go-zkfp](https://github.com/elmyrockers/go-zkfp)
Based on ZKTeco ZKFinger Reader SDK C API Version 2.0

## Table of Contents

1. [Overview](#1-overview)
2. [Disclaimer](#2-disclaimer)
3. [System Requirements](#3-system-requirements)
4. [Installation and Deployment](#4-installation-and-deployment)
5. [Description of Go Interfaces](#5-description-of-go-interfaces)
   - [5.1 Design Conventions](#51-design-conventions)
   - [5.2 Function Mapping](#52-function-mapping)
   - [5.3 Type Definitions and Constants](#53-type-definitions-and-constants)
6. [Global Library Functions](#6-global-library-functions)
   - [Init](#init)
   - [Terminate](#terminate)
   - [GetDeviceCount](#getdevicecount)
   - [OpenDevice](#opendevice)
   - [DBInit](#dbinit)
7. [Device Methods](#7-device-methods)
   - [Device.Close](#deviceclose)
   - [Device.ImageSize](#deviceimagesize)
   - [Device.AcquireFingerprint](#deviceacquirefingerprint)
   - [Device.AcquireFingerprintImage](#deviceacquirefingerprintimage)
   - [Device.GetParameter](#devicegetparameter)
   - [Device.SetParameter](#devicesetparameter)
8. [Database/Algorithm Methods](#8-databasealgorithm-methods)
   - [DB.Free](#dbfree)
   - [DB.Merge](#dbmerge)
   - [DB.Add](#dbadd)
   - [DB.Del](#dbdel)
   - [DB.Clear](#dbclear)
   - [DB.Count](#dbcount)
   - [DB.Identify](#dbidentify)
   - [DB.Match](#dbmatch)
   - [DB.ExtractFromImage](#dbextractfromimage)
9. [Usage Example](#9-usage-example)
10. [Appendixes](#10-appendixes)
    - [Appendix 1: List of Common Parameter Codes](#appendix-1-list-of-common-parameter-codes)
    - [Appendix 2: Descriptions of Returned Error Values](#appendix-2-descriptions-of-returned-error-values)

---

## 1. Overview

`go-zkfp` ([github.com/elmyrockers/go-zkfp](https://github.com/elmyrockers/go-zkfp)) is a Go binding for `libzkfp.dll`, the ZKFinger Reader SDK from ZKTeco. It allows Go programs to enumerate ZKTeco fingerprint readers, capture images and templates, and register, identify, and match fingerprints.

This document describes the Go API, mapping one-to-one to the C functions of the ZKFinger Reader SDK (C API Version 2.0).

## 2. Disclaimer

`go-zkfp` is an independent open-source wrapper and does not include the ZKTeco SDK binaries. `libzkfp.dll` must be obtained by installing the official ZKFinger SDK from ZKTeco. Use of the SDK is subject to ZKTeco's licensing terms.

## 3. System Requirements

1. Operating system: Windows XP or a later version
2. Go 1.27.1 or later
3. ZKFinger SDK 5.x / ZKOnline SDK 5.x installed (providing `libzkfp.dll` and reader drivers)
4. The architecture of the Go program (`GOARCH=386` or `amd64`) must match the architecture of `libzkfp.dll`
5. cgo and a C compiler are **not** required: the library is loaded at runtime via `golang.org/x/sys/windows` (`windows.NewLazyDLL`) utilizing standard calling conventions required by the SDK.

## 4. Installation and Deployment

1. Install ZKFinger SDK 5.x / ZKOnline SDK 5.x.
2. Add the package to your module:

   ```bash
   go get github.com/elmyrockers/go-zkfp
   ```

3. Make `libzkfp.dll` loadable by placing it next to your executable or in a directory on your system `PATH`.
4. Import the package in your code:

   ```go
   import "github.com/elmyrockers/go-zkfp"
   ```

## 5. Description of Go Interfaces

### 5.1 Design Conventions

- **Errors.** Functions return a standard Go `error` instead of raw integer return codes. Non-zero SDK results return values of type `zkfp.Error` (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values)), allowing direct comparison with `errors.Is` or `errors.As`.
- **Handles.** C `HANDLE` types are encapsulated into clean Go types: `*Device` (for reader instances) and `*DB` (for algorithm caches/databases).
- **Buffers.** C pointer-and-size parameter pairs are represented as native Go slices (`[]byte`). Output buffers are automatically allocated and trimmed to the actual data size returned by the SDK.
- **Cleanup.** `Close()` and `Free()` methods are implemented safely and idempotently.

### 5.2 Function Mapping

| C function | Go function / method |
|---|---|
| `ZKFPM_Init` | `zkfp.Init` |
| `ZKFPM_Terminate` | `zkfp.Terminate` |
| `ZKFPM_GetDeviceCount` | `zkfp.GetDeviceCount` |
| `ZKFPM_OpenDevice` | `zkfp.OpenDevice` |
| `ZKFPM_CloseDevice` | `(*Device).Close` |
| `ZKFPM_SetParameters` | `(*Device).SetParameter` |
| `ZKFPM_GetParameters` | `(*Device).GetParameter` |
| `ZKFPM_AcquireFingerprint` | `(*Device).AcquireFingerprint` |
| `ZKFPM_AcquireFingerprintImage` | `(*Device).AcquireFingerprintImage` |
| `ZKFPM_DBInit` | `zkfp.DBInit` |
| `ZKFPM_DBFree` | `(*DB).Free` |
| `ZKFPM_DBMerge` | `(*DB).Merge` |
| `ZKFPM_DBAdd` | `(*DB).Add` |
| `ZKFPM_DBDel` | `(*DB).Del` |
| `ZKFPM_DBClear` | `(*DB).Clear` |
| `ZKFPM_DBCount` | `(*DB).Count` |
| `ZKFPM_DBIdentify` | `(*DB).Identify` |
| `ZKFPM_DBMatch` | `(*DB).Match` |
| `ZKFPM_ExtractFromImage` | `(*DB).ExtractFromImage` |

### 5.3 Type Definitions and Constants

```go
package zkfp

type Device struct { /* unexported */ }
type DB struct { /* unexported */ }
type Error int32

const (
    maxTemplateSize = 2048
    paramImageSize  = 106
)
```

## 6. Global Library Functions

### Init

```go
func Init() error
```

**Purpose:** Initializes the ZKFinger SDK resources. Must be called once before invoking other methods.

**Parameters:** None

**Return value:**

- `nil`: Succeeded (including code 1 indicating already initialized)
- `error`: Failed (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### Terminate

```go
func Terminate() error
```

**Purpose:** Releases ZKFinger SDK resources.

**Parameters:** None

**Return value:**

- `nil`: Succeeded
- `error`: Failed (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### GetDeviceCount

```go
func GetDeviceCount() (int, error)
```

**Purpose:** Returns the number of connected fingerprint devices.

**Parameters:** None

**Return value:**

- `int, nil`: Connected device count (>= 0)
- `0, error`: Failed to query device count (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### OpenDevice

```go
func OpenDevice(index int) (*Device, error)
```

**Purpose:** Opens a fingerprint device by its zero-based index.

**Parameters:**

- `index`: Zero-based device index integer

**Return value:**

- `*Device, nil`: Opened device handle instance
- `nil, error`: Failed to open device (returns `ErrOpenDevice`)

### DBInit

```go
func DBInit() (*DB, error)
```

**Purpose:** Creates an algorithm cache/database instance.

**Parameters:** None

**Return value:**

- `*DB, nil`: Initialized algorithm database handle instance
- `nil, error`: Initialization failed

## 7. Device Methods

### Device.Close

```go
func (d *Device) Close() error
```

**Purpose:** Closes the fingerprint device handle.

**Parameters:** None

**Return value:**

- `nil`: Succeeded
- `error`: Failed to close device handle (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### Device.ImageSize

```go
func (d *Device) ImageSize() (int, error)
```

**Purpose:** Queries the required fingerprint image buffer size from the device.

**Parameters:** None

**Return value:**

- `int, nil`: Required image buffer size in bytes
- `0, error`: Failed to retrieve size parameter (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### Device.AcquireFingerprint

```go
func (d *Device) AcquireFingerprint(image []byte) ([]byte, error)
```

**Purpose:** Captures a fingerprint image and extracts its template.

**Parameters:**

- `image`: Pre-allocated byte slice buffer to store the scanned fingerprint image (sized using `ImageSize`)

**Return value:**

- `[]byte, nil`: Extracted fingerprint template slice trimmed to actual size
- `nil, error`: Capture or extraction failed (e.g., `ErrCaptureFailed`)

### Device.AcquireFingerprintImage

```go
func (d *Device) AcquireFingerprintImage(image []byte) error
```

**Purpose:** Captures a raw fingerprint image without template extraction.

**Parameters:**

- `image`: Pre-allocated byte slice buffer to receive the raw image data

**Return value:**

- `nil`: Succeeded
- `error`: Image capture failed (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### Device.GetParameter

```go
func (d *Device) GetParameter(code int, size int) ([]byte, error)
```

**Purpose:** Reads a device parameter code value.

**Parameters:**

- `code`: Integer parameter code (see [Appendix 1](#appendix-1-list-of-common-parameter-codes))
- `size`: Allocation buffer size based on parameter type (e.g., 4 for integer, larger for strings)

**Return value:**

- `[]byte, nil`: Raw parameter bytes trimmed to size reported by SDK
- `nil, error`: Parameter query failed (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### Device.SetParameter

```go
func (d *Device) SetParameter(code int, value []byte) error
```

**Purpose:** Writes or configures a device parameter code value.

**Parameters:**

- `code`: Integer parameter code (see [Appendix 1](#appendix-1-list-of-common-parameter-codes))
- `value`: Raw parameter data bytes (e.g., 4-byte little-endian integer)

**Return value:**

- `nil`: Succeeded
- `error`: Parameter update failed (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

## 8. Database/Algorithm Methods

### DB.Free

```go
func (db *DB) Free() error
```

**Purpose:** Releases the algorithm cache resources.

**Parameters:** None

**Return value:**

- `nil`: Succeeded
- `error`: Failed to release cache (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### DB.Merge

```go
func (db *DB) Merge(template1 []byte, template2 []byte, template3 []byte) ([]byte, error)
```

**Purpose:** Combines three fingerprint templates into a single registered template.

**Parameters:**

- `template1`: First capture template byte slice
- `template2`: Second capture template byte slice
- `template3`: Third capture template byte slice

**Return value:**

- `[]byte, nil`: Merged template slice trimmed to final size
- `nil, error`: Merge operation failed (e.g., `ErrMergeFailed`)

### DB.Add

```go
func (db *DB) Add(fid uint, template []byte) error
```

**Purpose:** Adds a registered fingerprint template to the algorithm cache with an associated ID.

**Parameters:**

- `fid`: Unsigned integer fingerprint ID (> 0)
- `template`: Registered template byte slice

**Return value:**

- `nil`: Succeeded
- `error`: Failed to add template (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### DB.Del

```go
func (db *DB) Del(fid uint) error
```

**Purpose:** Deletes a fingerprint template from the algorithm cache by its ID.

**Parameters:**

- `fid`: Unsigned integer fingerprint ID to delete

**Return value:**

- `nil`: Succeeded
- `error`: Failed to delete template (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### DB.Clear

```go
func (db *DB) Clear() error
```

**Purpose:** Removes all fingerprint templates from the algorithm cache.

**Parameters:** None

**Return value:**

- `nil`: Succeeded
- `error`: Failed to clear cache (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### DB.Count

```go
func (db *DB) Count() (int, error)
```

**Purpose:** Returns the total number of fingerprint templates stored in the algorithm cache.

**Parameters:** None

**Return value:**

- `int, nil`: Template count integer
- `0, error`: Failed to fetch count (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### DB.Identify

```go
func (db *DB) Identify(template []byte) (uint, uint, error)
```

**Purpose:** Performs a 1:N fingerprint comparison against all templates in the cache.

**Parameters:**

- `template`: Query fingerprint template byte slice

**Return value:**

- `fid, score, nil`: Matched fingerprint ID and comparison score (both `uint`)
- `0, 0, error`: Identification failed or no match found (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### DB.Match

```go
func (db *DB) Match(template1, template2 []byte) (int, error)
```

**Purpose:** Compares two fingerprint templates directly (1:1 matching).

**Parameters:**

- `template1`: First fingerprint template byte slice
- `template2`: Second fingerprint template byte slice

**Return value:**

- `int, nil`: Comparison match score integer (>= 0)
- `0, error`: Comparison calculation failed (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

### DB.ExtractFromImage

```go
func (db *DB) ExtractFromImage(path string, dpi uint) ([]byte, error)
```

**Purpose:** Extracts a fingerprint template from an image file (BMP or JPG format).

**Parameters:**

- `path`: Full file system path string to the image file
- `dpi`: Image resolution DPI unsigned integer

**Return value:**

- `[]byte, nil`: Extracted template byte slice trimmed to actual size
- `nil, error`: Extraction failed or unsupported version (see [Appendix 2](#appendix-2-descriptions-of-returned-error-values))

## 9. Usage Example

```go
package main

import (
    "fmt"
    "log"

    "github.com/elmyrockers/go-zkfp"
)

func main() {
    // 1. Initialize the SDK
    if err := zkfp.Init(); err != nil {
        log.Fatalf("Init failed: %v", err)
    }
    defer zkfp.Terminate()

    // 2. Initialize the algorithm database/cache for matching & storage
    db, err := zkfp.DBInit()
    if err != nil {
        log.Fatalf("DBInit failed: %v", err)
    }
    defer db.Free()

    // 3. Check for connected devices
    count, err := zkfp.GetDeviceCount()
    if err != nil || count == 0 {
        log.Fatalf("No device found: %v", err)
    }

    // 4. Open the first device
    dev, err := zkfp.OpenDevice(0)
    if err != nil {
        log.Fatalf("OpenDevice failed: %v", err)
    }
    defer dev.Close()

    fmt.Println("Device opened and Algorithm DB initialized successfully!")

    // 5. Query required image buffer size and capture fingerprint
    imgSizeBytes, err := dev.ImageSize()
    if err != nil {
        log.Fatalf("Failed to get image size: %v", err)
    }

    imgBuf := make([]byte, imgSizeBytes)
    fmt.Println("Please place your finger on the scanner...")

    template, err := dev.AcquireFingerprint(imgBuf)
    if err != nil {
        log.Fatalf("Fingerprint capture failed: %v", err)
    }

    // 6. Add the captured template to the algorithm database with ID 1
    var fingerprintID uint = 1
    if err := db.Add(fingerprintID, template); err != nil {
        log.Fatalf("Failed to add template to DB: %v", err)
    }

    totalCount, _ := db.Count()
    fmt.Printf("Fingerprint successfully registered! Total records in DB: %d\n", totalCount)
}
```

## 10. Appendixes

### Appendix 1: List of Common Parameter Codes

`(*Device).GetParameter` and `(*Device).SetParameter` take an SDK parameter code. Integer parameters are 4-byte little-endian values.

| Constant              |  Code | Access     | Type         | Description                                                                 |
| --------------------- | ----: | ---------- | ------------ | --------------------------------------------------------------------------- |
| `ParamImageWidth`     |   `1` | Read       | Int          | Fingerprint image width in pixels.                                          |
| `ParamImageHeight`    |   `2` | Read       | Int          | Fingerprint image height in pixels.                                         |
| `ParamImageDPI`       |   `3` | Read/Write | Int          | Image DPI (`LIVEID20R` only; 750/1000 recommended for children).            |
| `ParamWhiteLED`       | `101` | Write      | Int          | White LED on the reader (`1` = blink, `0` = off).                           |
| `ParamGreenLED`       | `102` | Write      | Int          | Green LED on the reader (`1` = blink, `0` = off).                           |
| `ParamRedLED`         | `103` | Write      | Int          | Red LED on the reader (`1` = blink, `0` = off).                             |
| `ParamBuzzer`         | `104` | Write      | Int          | Buzzer (`1` = start buzzing, `0` = off).                                    |
| `ParamImageSize`      | `106` | Read       | Int          | Fingerprint image buffer size in bytes.                                     |
| `ParamVIDPID`         | `1015` | Read      | 4-byte array | USB VID and PID (first two bytes are the VID, last two bytes the PID).      |
| `ParamVendorInfo`     | `1101` | Read      | String       | Vendor information.                                                         |
| `ParamProductName`    | `1102` | Read      | String       | Product name.                                                               |
| `ParamSerialNumber`   | `1103` | Read      | String       | Device serial number (SN).                                                  |
| `ParamAntiFake`       | `2002` | Read/Write | Int         | Anti-fake function (`LIVEID20R` only; `1` = enable, `0` = disable).         |
| `ParamFakeStatus`     | `2004` | Read      | Int          | True if the lower five bits are all 1's (`value&31 == 31`).                 |
| `ParamTemplateFormat` | `10001` | Write    | Int          | Template format (ISO/ANSI readers only; `0` = ANSI378, `1` = ISO 19794-2).  |

### Appendix 2: Descriptions of Returned Error Values

| Code | Go Error Value | Description |
|---:|---|---|
| 1 | `nil` | SDK already initialized (treated as success by `Init`) |
| 0 | `nil` | Operation succeeded |
| -1 | `ErrInitLib` | Failed to initialize the algorithm library |
| -2 | `ErrInitCapture` | Failed to initialize the capture library |
| -3 | `ErrNoDevice` | No device connected |
| -4 | `ErrNotSupported` | Not supported by the interface |
| -5 | `ErrInvalidParam` | Invalid parameter |
| -6 | `ErrOpenDevice` | Failed to start the device |
| -7 | `ErrInvalidHandle` | Invalid handle |
| -8 | `ErrCaptureFailed` | Failed to capture the image |
| -9 | `ErrExtractFailed` | Failed to extract the fingerprint template |
| -10 | `ErrAbort` | Suspension operation |
| -11 | `ErrNoMemory` | Insufficient memory |
| -12 | `ErrBusy` | Device is busy capturing |
| -13 | `ErrAddFailed` | Failed to add template to memory |
| -14 | `ErrDeleteFailed` | Failed to delete template |
| -17 | `ErrOperationFailed` | Operation failed (other error) |
| -18 | `ErrCaptureCancelled` | Capture cancelled |
| -20 | `ErrMatchFailed` | Fingerprint comparison failed |
| -22 | `ErrMergeFailed` | Failed to combine registered templates |
| -23 | `ErrOpenFile` | Opening the file failed |
| -24 | `ErrImageProcess` | Image processing failed |