# Cross-Platform Audio Device Management Package in Go

## **Project Name**: `go-audio-control`

---

## **1. Objectives**

Develop a Go package to:

1. **Enumerate** all audio input/output devices.
2. **Get currently active input/output device.**
3. **Set active input/output device.**
4. **Monitor for device changes**, including:
   - Device added/removed
   - Active device change
   - Disconnection (especially Bluetooth or USB audio)
5. **Send these events** to a Go callback handler (suitable for Wails or scripting).

---

## **2. Target Platforms**

| OS       | Core API                          | Note                             |
|----------|-----------------------------------|----------------------------------|
| macOS    | CoreAudio (C-based)               | Stable; requires CGo             |
| Windows  | Core Audio APIs (WASAPI + MMDevice API via COM) | Requires COM and syscall/CGo |

---

## **3. Public API Design**

```go
type AudioDevice struct {
    ID          string
    Name        string
    IsInput     bool
    IsOutput    bool
    IsActive    bool
    IsConnected bool
}

// Enumerate devices
func ListAudioDevices() ([]AudioDevice, error)

// Get active output device
func GetActiveOutputDevice() (AudioDevice, error)

// Set active output device
func SetActiveOutputDevice(deviceID string) error

// Monitor changes
func OnDeviceChange(callback func(Event))

type EventType int
const (
    DeviceAdded EventType = iota
    DeviceRemoved
    ActiveDeviceChanged
    DeviceDisconnected
)

type Event struct {
    Type     EventType
    DeviceID string
    Info     *AudioDevice
}
```

---

## **4. Directory Structure**

```
go-audio-control/
├── audiocontrol.go        # Cross-platform interface
├── platform/
│   ├── darwin/
│   │   ├── devices.go     # macOS CoreAudio logic
│   │   └── listener.go
│   ├── windows/
│   │   ├── devices.go     # Windows MMDevice/COM logic
│   │   └── listener.go
│   └── ...
├── internal/
│   └── util.go            # Helpers, conversions
└── examples/
    ├── list_devices.go    # Usage examples
    ├── monitor_events.go
    └── switch_device.go
```

---

## **5. macOS Implementation Details**

- **List Devices**: `AudioObjectGetPropertyData`
- **Get/Set Active**: `kAudioHardwarePropertyDefaultOutputDevice`
- **Listen Events**:
  - Use `AudioObjectAddPropertyListener`
  - Keys:
    - `kAudioHardwarePropertyDevices`
    - `kAudioHardwarePropertyDefaultOutputDevice`

Wrap with `#cgo` and Objective-C or pure C depending on the API.

---

## **6. Windows Implementation Details**

Use **WASAPI/MMDevice API via COM**:

- **Device Enumeration**:
  - `IMMDeviceEnumerator::EnumAudioEndpoints`
- **Get Active**:
  - `IMMDeviceEnumerator::GetDefaultAudioEndpoint`
- **Set Active**:
  - Requires workaround using `IAudioPolicyConfigFactory` or similar
- **Event Listening**:
  - Implement `IMMNotificationClient` COM interface

Leverage:
- [`go-ole`](https://github.com/go-ole/go-ole)
- `syscall.NewCallback` to receive COM-style events

---

## **7. Wails Integration Example**

```go
go func() {
    audiocontrol.OnDeviceChange(func(e audiocontrol.Event) {
        runtime.EventsEmit(ctx, "audioEvent", e)
    })
}()
```

In JS:
```js
window.runtime.EventsOn("audioEvent", e => {
    console.log("Audio Event", e);
});
```

---

## **8. Error Handling and Compatibility**

- Fallback gracefully if permissions are denied (e.g. macOS mic access).
- Return human-readable errors: `ErrUnsupported`, `ErrPermission`, etc.
- Optional: expose `FeatureSupport` struct for querying platform capabilities.

---

## **9. Development Notes**

- Use build tags (`// +build darwin`, `// +build windows`) to isolate platform code.
- Target minimal external dependencies.
- Maintain consistent naming and device abstraction between platforms.
- Avoid blocking operations in listeners to maintain responsiveness.
