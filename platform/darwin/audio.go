//go:build darwin
// +build darwin

package darwin

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreAudio -framework Foundation

#include <CoreAudio/CoreAudio.h>
#include <dispatch/dispatch.h>
#include <stdio.h>
#include <stdlib.h>

// Forward declaration for Go callback
void goDeviceChangeCallback(int eventType, AudioObjectID deviceID);

// Helper function to get string property
static char* getDeviceStringProperty(AudioObjectID deviceID, AudioObjectPropertySelector selector) {
    CFStringRef value = NULL;
    UInt32 size = sizeof(value);

    AudioObjectPropertyAddress address = {
        selector,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    OSStatus status = AudioObjectGetPropertyData(deviceID, &address, 0, NULL, &size, &value);
    if (status != noErr || value == NULL) {
        return NULL;
    }

    CFIndex length = CFStringGetLength(value);
    CFIndex maxSize = CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;
    char *buffer = (char *)malloc(maxSize);

    if (CFStringGetCString(value, buffer, maxSize, kCFStringEncodingUTF8)) {
        CFRelease(value);
        return buffer;
    }

    CFRelease(value);
    free(buffer);
    return NULL;
}

// Check if device has input/output streams
static int hasStreams(AudioObjectID deviceID, AudioObjectPropertyScope scope) {
    AudioObjectPropertyAddress address = {
        kAudioDevicePropertyStreams,
        scope,
        kAudioObjectPropertyElementMain
    };

    UInt32 size = 0;
    OSStatus status = AudioObjectGetPropertyDataSize(deviceID, &address, 0, NULL, &size);
    return (status == noErr && size > 0) ? 1 : 0;
}

// Get device transport type to check if connected
static int isDeviceConnected(AudioObjectID deviceID) {
    UInt32 transportType = 0;
    UInt32 size = sizeof(transportType);

    AudioObjectPropertyAddress address = {
        kAudioDevicePropertyTransportType,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    OSStatus status = AudioObjectGetPropertyData(deviceID, &address, 0, NULL, &size, &transportType);
    if (status != noErr) {
        return 1; // Assume connected if we can't determine
    }

    // Check if device is alive
    UInt32 isAlive = 0;
    size = sizeof(isAlive);
    address.mSelector = kAudioDevicePropertyDeviceIsAlive;

    status = AudioObjectGetPropertyData(deviceID, &address, 0, NULL, &size, &isAlive);
    if (status == noErr && !isAlive) {
        return 0;
    }

    return 1;
}

// Get all audio devices
static AudioObjectID* getAllAudioDevices(int* count) {
    AudioObjectPropertyAddress address = {
        kAudioHardwarePropertyDevices,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    UInt32 size = 0;
    OSStatus status = AudioObjectGetPropertyDataSize(kAudioObjectSystemObject, &address, 0, NULL, &size);
    if (status != noErr) {
        *count = 0;
        return NULL;
    }

    *count = size / sizeof(AudioObjectID);
    AudioObjectID* devices = (AudioObjectID*)malloc(size);

    status = AudioObjectGetPropertyData(kAudioObjectSystemObject, &address, 0, NULL, &size, devices);
    if (status != noErr) {
        free(devices);
        *count = 0;
        return NULL;
    }

    return devices;
}

// Get default device
static AudioObjectID getDefaultDevice(int isInput) {
    AudioObjectPropertyAddress address = {
        isInput ? kAudioHardwarePropertyDefaultInputDevice : kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    AudioObjectID deviceID = kAudioObjectUnknown;
    UInt32 size = sizeof(deviceID);

    AudioObjectGetPropertyData(kAudioObjectSystemObject, &address, 0, NULL, &size, &deviceID);
    return deviceID;
}

// Set default device
static OSStatus setDefaultDevice(AudioObjectID deviceID, int isInput) {
    AudioObjectPropertyAddress address = {
        isInput ? kAudioHardwarePropertyDefaultInputDevice : kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    return AudioObjectSetPropertyData(kAudioObjectSystemObject, &address, 0, NULL, sizeof(deviceID), &deviceID);
}

// Property listener callback
static OSStatus propertyListenerCallback(
    AudioObjectID objectID,
    UInt32 numberAddresses,
    const AudioObjectPropertyAddress* addresses,
    void* clientData) {

    for (UInt32 i = 0; i < numberAddresses; i++) {
        if (addresses[i].mSelector == kAudioHardwarePropertyDevices) {
            // Device list changed - could be added or removed
            goDeviceChangeCallback(0, objectID); // 0 = DeviceListChanged
        } else if (addresses[i].mSelector == kAudioHardwarePropertyDefaultOutputDevice ||
                   addresses[i].mSelector == kAudioHardwarePropertyDefaultInputDevice) {
            // Active device changed
            goDeviceChangeCallback(2, objectID); // 2 = ActiveDeviceChanged
        } else if (addresses[i].mSelector == kAudioDevicePropertyDeviceIsAlive) {
            // Device disconnected
            goDeviceChangeCallback(3, objectID); // 3 = DeviceDisconnected
        }
    }

    return noErr;
}

// Start monitoring
static int startMonitoring() {
    OSStatus status;

    // Monitor device list changes
    AudioObjectPropertyAddress devicesAddress = {
        kAudioHardwarePropertyDevices,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    status = AudioObjectAddPropertyListener(
        kAudioObjectSystemObject,
        &devicesAddress,
        propertyListenerCallback,
        NULL
    );
    if (status != noErr) return -1;

    // Monitor default output device changes
    AudioObjectPropertyAddress outputAddress = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    status = AudioObjectAddPropertyListener(
        kAudioObjectSystemObject,
        &outputAddress,
        propertyListenerCallback,
        NULL
    );
    if (status != noErr) return -1;

    // Monitor default input device changes
    AudioObjectPropertyAddress inputAddress = {
        kAudioHardwarePropertyDefaultInputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    status = AudioObjectAddPropertyListener(
        kAudioObjectSystemObject,
        &inputAddress,
        propertyListenerCallback,
        NULL
    );
    if (status != noErr) return -1;

    return 0;
}

// Stop monitoring
static void stopMonitoring() {
    AudioObjectPropertyAddress devicesAddress = {
        kAudioHardwarePropertyDevices,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    AudioObjectRemovePropertyListener(
        kAudioObjectSystemObject,
        &devicesAddress,
        propertyListenerCallback,
        NULL
    );

    AudioObjectPropertyAddress outputAddress = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    AudioObjectRemovePropertyListener(
        kAudioObjectSystemObject,
        &outputAddress,
        propertyListenerCallback,
        NULL
    );

    AudioObjectPropertyAddress inputAddress = {
        kAudioHardwarePropertyDefaultInputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };
    AudioObjectRemovePropertyListener(
        kAudioObjectSystemObject,
        &inputAddress,
        propertyListenerCallback,
        NULL
    );
}

// Add listener for device alive property
static void addDeviceAliveListener(AudioObjectID deviceID) {
    AudioObjectPropertyAddress aliveAddress = {
        kAudioDevicePropertyDeviceIsAlive,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    AudioObjectAddPropertyListener(
        deviceID,
        &aliveAddress,
        propertyListenerCallback,
        NULL
    );
}

// Remove listener for device alive property
static void removeDeviceAliveListener(AudioObjectID deviceID) {
    AudioObjectPropertyAddress aliveAddress = {
        kAudioDevicePropertyDeviceIsAlive,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    AudioObjectRemovePropertyListener(
        deviceID,
        &aliveAddress,
        propertyListenerCallback,
        NULL
    );
}
*/
import "C"
import (
	"fmt"
	"strconv"
	"sync"
	"unsafe"
)

// AudioDevice represents an audio device
type AudioDevice struct {
	ID          string
	Name        string
	IsInput     bool
	IsOutput    bool
	IsActive    bool
	IsConnected bool
}

// EventType represents the type of audio device event
type EventType int

const (
	DeviceAdded EventType = iota
	DeviceRemoved
	ActiveDeviceChanged
	DeviceDisconnected
)

// Event represents an audio device event
type Event struct {
	Type     EventType
	DeviceID string
	Info     *AudioDevice
}

var (
	callbackMutex sync.Mutex
	userCallback  func(Event)
	deviceStates  = make(map[C.AudioObjectID]bool) // Track known devices
	stateMutex    sync.RWMutex
)

// ListAudioDevices enumerates all audio devices
func ListAudioDevices() ([]AudioDevice, error) {
	var count C.int
	deviceIDs := C.getAllAudioDevices(&count)
	if deviceIDs == nil {
		return nil, fmt.Errorf("failed to enumerate audio devices")
	}
	defer C.free(unsafe.Pointer(deviceIDs))

	defaultInputID := C.getDefaultDevice(1)
	defaultOutputID := C.getDefaultDevice(0)

	devices := make([]AudioDevice, 0, int(count))
	deviceArray := (*[1 << 30]C.AudioObjectID)(unsafe.Pointer(deviceIDs))[:count:count]

	for _, deviceID := range deviceArray {
		// Get device name
		namePtr := C.getDeviceStringProperty(deviceID, C.kAudioObjectPropertyName)
		if namePtr == nil {
			continue
		}
		name := C.GoString(namePtr)
		C.free(unsafe.Pointer(namePtr))

		// Get device UID
		uidPtr := C.getDeviceStringProperty(deviceID, C.kAudioDevicePropertyDeviceUID)
		if uidPtr == nil {
			continue
		}
		uid := C.GoString(uidPtr)
		C.free(unsafe.Pointer(uidPtr))

		// Check if it has input/output streams
		hasInput := C.hasStreams(deviceID, C.kAudioDevicePropertyScopeInput) == 1
		hasOutput := C.hasStreams(deviceID, C.kAudioDevicePropertyScopeOutput) == 1

		// Skip devices with no streams
		if !hasInput && !hasOutput {
			continue
		}

		// Check if it's the active device
		isActiveInput := hasInput && deviceID == defaultInputID
		isActiveOutput := hasOutput && deviceID == defaultOutputID

		// Check if connected
		isConnected := C.isDeviceConnected(deviceID) == 1

		device := AudioDevice{
			ID:          uid,
			Name:        name,
			IsInput:     hasInput,
			IsOutput:    hasOutput,
			IsActive:    isActiveInput || isActiveOutput,
			IsConnected: isConnected,
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// GetActiveOutputDevice returns the currently active output device
func GetActiveOutputDevice() (AudioDevice, error) {
	deviceID := C.getDefaultDevice(0)
	if deviceID == C.kAudioObjectUnknown {
		return AudioDevice{}, fmt.Errorf("no default output device found")
	}

	// Get device name
	namePtr := C.getDeviceStringProperty(deviceID, C.kAudioObjectPropertyName)
	if namePtr == nil {
		return AudioDevice{}, fmt.Errorf("failed to get device name")
	}
	name := C.GoString(namePtr)
	C.free(unsafe.Pointer(namePtr))

	// Get device UID
	uidPtr := C.getDeviceStringProperty(deviceID, C.kAudioDevicePropertyDeviceUID)
	if uidPtr == nil {
		return AudioDevice{}, fmt.Errorf("failed to get device UID")
	}
	uid := C.GoString(uidPtr)
	C.free(unsafe.Pointer(uidPtr))

	// Check if connected
	isConnected := C.isDeviceConnected(deviceID) == 1

	return AudioDevice{
		ID:          uid,
		Name:        name,
		IsInput:     false,
		IsOutput:    true,
		IsActive:    true,
		IsConnected: isConnected,
	}, nil
}

// SetActiveOutputDevice sets the active output device by ID
func SetActiveOutputDevice(deviceUID string) error {
	// First, find the device ID from the UID
	var count C.int
	deviceIDs := C.getAllAudioDevices(&count)
	if deviceIDs == nil {
		return fmt.Errorf("failed to enumerate audio devices")
	}
	defer C.free(unsafe.Pointer(deviceIDs))

	deviceArray := (*[1 << 30]C.AudioObjectID)(unsafe.Pointer(deviceIDs))[:count:count]
	var targetDeviceID C.AudioObjectID = C.kAudioObjectUnknown

	for _, deviceID := range deviceArray {
		uidPtr := C.getDeviceStringProperty(deviceID, C.kAudioDevicePropertyDeviceUID)
		if uidPtr == nil {
			continue
		}
		uid := C.GoString(uidPtr)
		C.free(unsafe.Pointer(uidPtr))

		if uid == deviceUID {
			// Check if it has output streams
			if C.hasStreams(deviceID, C.kAudioDevicePropertyScopeOutput) == 1 {
				targetDeviceID = deviceID
				break
			}
		}
	}

	if targetDeviceID == C.kAudioObjectUnknown {
		return fmt.Errorf("device with UID %s not found or is not an output device", deviceUID)
	}

	// Set as default output device
	status := C.setDefaultDevice(targetDeviceID, 0)
	if status != C.noErr {
		return fmt.Errorf("failed to set default output device: OSStatus %d", status)
	}

	return nil
}

// Helper function to convert AudioObjectID to string
func audioObjectIDToString(id C.AudioObjectID) string {
	return strconv.FormatUint(uint64(id), 10)
}

//export goDeviceChangeCallback
func goDeviceChangeCallback(eventType C.int, deviceID C.AudioObjectID) {
	callbackMutex.Lock()
	callback := userCallback
	callbackMutex.Unlock()

	if callback == nil {
		return
	}

	switch eventType {
	case 0: // Device list changed
		handleDeviceListChange(callback)
	case 2: // Active device changed
		handleActiveDeviceChange(callback)
	case 3: // Device disconnected
		handleDeviceDisconnected(callback, deviceID)
	}
}

func handleDeviceListChange(callback func(Event)) {
	// Get current device list
	var count C.int
	newDeviceIDs := C.getAllAudioDevices(&count)
	if newDeviceIDs == nil {
		return
	}
	defer C.free(unsafe.Pointer(newDeviceIDs))

	newDeviceArray := (*[1 << 30]C.AudioObjectID)(unsafe.Pointer(newDeviceIDs))[:count:count]
	newDeviceSet := make(map[C.AudioObjectID]bool)

	// Check for new devices
	for _, deviceID := range newDeviceArray {
		newDeviceSet[deviceID] = true

		stateMutex.RLock()
		_, exists := deviceStates[deviceID]
		stateMutex.RUnlock()

		if !exists {
			// New device found
			device := getDeviceInfo(deviceID)
			if device != nil {
				callback(Event{
					Type:     DeviceAdded,
					DeviceID: device.ID,
					Info:     device,
				})

				// Add alive listener for this device
				C.addDeviceAliveListener(deviceID)
			}
		}
	}

	// Check for removed devices
	stateMutex.Lock()
	for deviceID := range deviceStates {
		if !newDeviceSet[deviceID] {
			// Device removed
			device := getDeviceInfo(deviceID)
			if device != nil {
				callback(Event{
					Type:     DeviceRemoved,
					DeviceID: device.ID,
					Info:     device,
				})
			}

			// Remove alive listener
			C.removeDeviceAliveListener(deviceID)
			delete(deviceStates, deviceID)
		}
	}

	// Update device states
	deviceStates = newDeviceSet
	stateMutex.Unlock()
}

func handleActiveDeviceChange(callback func(Event)) {
	// Get the new active output device
	device, err := GetActiveOutputDevice()
	if err == nil {
		callback(Event{
			Type:     ActiveDeviceChanged,
			DeviceID: device.ID,
			Info:     &device,
		})
	}
}

func handleDeviceDisconnected(callback func(Event), deviceID C.AudioObjectID) {
	device := getDeviceInfo(deviceID)
	if device != nil {
		callback(Event{
			Type:     DeviceDisconnected,
			DeviceID: device.ID,
			Info:     device,
		})
	}
}

func getDeviceInfo(deviceID C.AudioObjectID) *AudioDevice {
	// Get device name
	namePtr := C.getDeviceStringProperty(deviceID, C.kAudioObjectPropertyName)
	if namePtr == nil {
		return nil
	}
	name := C.GoString(namePtr)
	C.free(unsafe.Pointer(namePtr))

	// Get device UID
	uidPtr := C.getDeviceStringProperty(deviceID, C.kAudioDevicePropertyDeviceUID)
	if uidPtr == nil {
		return nil
	}
	uid := C.GoString(uidPtr)
	C.free(unsafe.Pointer(uidPtr))

	// Check properties
	hasInput := C.hasStreams(deviceID, C.kAudioDevicePropertyScopeInput) == 1
	hasOutput := C.hasStreams(deviceID, C.kAudioDevicePropertyScopeOutput) == 1
	isConnected := C.isDeviceConnected(deviceID) == 1

	// Check if active
	defaultInputID := C.getDefaultDevice(1)
	defaultOutputID := C.getDefaultDevice(0)
	isActive := (hasInput && deviceID == defaultInputID) || (hasOutput && deviceID == defaultOutputID)

	return &AudioDevice{
		ID:          uid,
		Name:        name,
		IsInput:     hasInput,
		IsOutput:    hasOutput,
		IsActive:    isActive,
		IsConnected: isConnected,
	}
}

// OnDeviceChange registers a callback for audio device events
func OnDeviceChange(callback func(Event)) {
	callbackMutex.Lock()
	userCallback = callback
	callbackMutex.Unlock()

	// Initialize device states
	var count C.int
	deviceIDs := C.getAllAudioDevices(&count)
	if deviceIDs != nil {
		deviceArray := (*[1 << 30]C.AudioObjectID)(unsafe.Pointer(deviceIDs))[:count:count]
		stateMutex.Lock()
		for _, deviceID := range deviceArray {
			deviceStates[deviceID] = true
			// Add alive listener for each device
			C.addDeviceAliveListener(deviceID)
		}
		stateMutex.Unlock()
		C.free(unsafe.Pointer(deviceIDs))
	}

	// Start monitoring
	C.startMonitoring()
}
