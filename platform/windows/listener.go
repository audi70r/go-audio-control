//go:build windows
// +build windows

package windows

import (
	"sync"
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

// DeviceEvent represents an audio device event
type DeviceEvent struct {
	Type     EventType
	DeviceID string
	Device   *AudioDevice
}

// EventType represents the type of device event
type EventType int

const (
	DeviceAdded EventType = iota
	DeviceRemoved
	ActiveDeviceChanged
	DeviceDisconnected
)

// DeviceListener manages device change notifications
type DeviceListener struct {
	enumerator       *IMMDeviceEnumerator
	notificationClient *NotificationClient
	callback         func(DeviceEvent)
	mu               sync.Mutex
	running          bool
}

// NotificationClient implements IMMNotificationClient
type NotificationClient struct {
	vtbl     *IMMNotificationClientVtbl
	ref      int32
	listener *DeviceListener
}

// IMMNotificationClientVtbl is the virtual method table for IMMNotificationClient
type IMMNotificationClientVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	OnDeviceStateChanged     uintptr
	OnDeviceAdded            uintptr
	OnDeviceRemoved          uintptr
	OnDefaultDeviceChanged   uintptr
	OnPropertyValueChanged   uintptr
}

// GetDevice retrieves a device by ID (method for IMMDeviceEnumerator)
func (e *IMMDeviceEnumerator) GetDevice(deviceID string) (*IMMDevice, error) {
	deviceIDPtr, err := syscall.UTF16PtrFromString(deviceID)
	if err != nil {
		return nil, err
	}
	
	var device *IMMDevice
	hr, _, _ := syscall.Syscall(
		e.vtbl.GetDevice,
		3,
		uintptr(unsafe.Pointer(e)),
		uintptr(unsafe.Pointer(deviceIDPtr)),
		uintptr(unsafe.Pointer(&device)),
	)
	if hr != 0 {
		return nil, ole.NewError(hr)
	}
	return device, nil
}

// RegisterEndpointNotificationCallback registers a notification callback
func (e *IMMDeviceEnumerator) RegisterEndpointNotificationCallback(client *NotificationClient) error {
	hr, _, _ := syscall.Syscall(
		e.vtbl.RegisterEndpointNotificationCallback,
		2,
		uintptr(unsafe.Pointer(e)),
		uintptr(unsafe.Pointer(client)),
		0,
	)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}

// UnregisterEndpointNotificationCallback unregisters a notification callback
func (e *IMMDeviceEnumerator) UnregisterEndpointNotificationCallback(client *NotificationClient) error {
	hr, _, _ := syscall.Syscall(
		e.vtbl.UnregisterEndpointNotificationCallback,
		2,
		uintptr(unsafe.Pointer(e)),
		uintptr(unsafe.Pointer(client)),
		0,
	)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}

// NewDeviceListener creates a new device listener
func NewDeviceListener() (*DeviceListener, error) {
	enumerator, err := CreateDeviceEnumerator()
	if err != nil {
		return nil, err
	}
	
	listener := &DeviceListener{
		enumerator: enumerator,
	}
	
	// Create notification client
	client := &NotificationClient{
		ref:      1,
		listener: listener,
	}
	
	// Set up vtable
	vtbl := &IMMNotificationClientVtbl{
		QueryInterface:         syscall.NewCallback(notificationClientQueryInterface),
		AddRef:                 syscall.NewCallback(notificationClientAddRef),
		Release:                syscall.NewCallback(notificationClientRelease),
		OnDeviceStateChanged:   syscall.NewCallback(notificationClientOnDeviceStateChanged),
		OnDeviceAdded:          syscall.NewCallback(notificationClientOnDeviceAdded),
		OnDeviceRemoved:        syscall.NewCallback(notificationClientOnDeviceRemoved),
		OnDefaultDeviceChanged: syscall.NewCallback(notificationClientOnDefaultDeviceChanged),
		OnPropertyValueChanged: syscall.NewCallback(notificationClientOnPropertyValueChanged),
	}
	client.vtbl = vtbl
	
	listener.notificationClient = client
	
	return listener, nil
}

// Start begins listening for device changes
func (l *DeviceListener) Start(callback func(DeviceEvent)) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.running {
		return nil
	}
	
	l.callback = callback
	
	// Register notification callback
	err := l.enumerator.RegisterEndpointNotificationCallback(l.notificationClient)
	if err != nil {
		return err
	}
	
	l.running = true
	return nil
}

// Stop stops listening for device changes
func (l *DeviceListener) Stop() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if !l.running {
		return nil
	}
	
	// Unregister notification callback
	err := l.enumerator.UnregisterEndpointNotificationCallback(l.notificationClient)
	if err != nil {
		return err
	}
	
	l.running = false
	l.callback = nil
	return nil
}

// Close releases resources
func (l *DeviceListener) Close() {
	l.Stop()
	if l.enumerator != nil {
		l.enumerator.Release()
		l.enumerator = nil
	}
}

// IUnknown methods for NotificationClient
func notificationClientQueryInterface(this unsafe.Pointer, riid *ole.GUID, ppvObject *unsafe.Pointer) uintptr {
	client := (*NotificationClient)(this)
	if ole.IsEqualGUID(riid, ole.IID_IUnknown) || ole.IsEqualGUID(riid, IID_IMMNotificationClient) {
		notificationClientAddRef(this)
		*ppvObject = this
		return 0
	}
	*ppvObject = nil
	return 0x80004002 // E_NOINTERFACE
}

func notificationClientAddRef(this unsafe.Pointer) uintptr {
	client := (*NotificationClient)(this)
	ref := syscall.AddInt32(&client.ref, 1)
	return uintptr(ref)
}

func notificationClientRelease(this unsafe.Pointer) uintptr {
	client := (*NotificationClient)(this)
	ref := syscall.AddInt32(&client.ref, -1)
	if ref == 0 {
		// Cleanup if needed
	}
	return uintptr(ref)
}

// IMMNotificationClient methods
func notificationClientOnDeviceStateChanged(this unsafe.Pointer, deviceID *uint16, newState uint32) uintptr {
	client := (*NotificationClient)(this)
	if client.listener.callback == nil {
		return 0
	}
	
	id := syscall.UTF16ToString((*[1024]uint16)(unsafe.Pointer(deviceID))[:])
	
	// Get device info
	device := getDeviceInfo(id)
	
	eventType := DeviceAdded
	if newState != DEVICE_STATE_ACTIVE {
		eventType = DeviceDisconnected
	}
	
	client.listener.callback(DeviceEvent{
		Type:     eventType,
		DeviceID: id,
		Device:   device,
	})
	
	return 0
}

func notificationClientOnDeviceAdded(this unsafe.Pointer, deviceID *uint16) uintptr {
	client := (*NotificationClient)(this)
	if client.listener.callback == nil {
		return 0
	}
	
	id := syscall.UTF16ToString((*[1024]uint16)(unsafe.Pointer(deviceID))[:])
	device := getDeviceInfo(id)
	
	client.listener.callback(DeviceEvent{
		Type:     DeviceAdded,
		DeviceID: id,
		Device:   device,
	})
	
	return 0
}

func notificationClientOnDeviceRemoved(this unsafe.Pointer, deviceID *uint16) uintptr {
	client := (*NotificationClient)(this)
	if client.listener.callback == nil {
		return 0
	}
	
	id := syscall.UTF16ToString((*[1024]uint16)(unsafe.Pointer(deviceID))[:])
	
	client.listener.callback(DeviceEvent{
		Type:     DeviceRemoved,
		DeviceID: id,
		Device:   nil,
	})
	
	return 0
}

func notificationClientOnDefaultDeviceChanged(this unsafe.Pointer, flow EDataFlow, role ERole, deviceID *uint16) uintptr {
	client := (*NotificationClient)(this)
	if client.listener.callback == nil {
		return 0
	}
	
	if deviceID == nil {
		return 0
	}
	
	id := syscall.UTF16ToString((*[1024]uint16)(unsafe.Pointer(deviceID))[:])
	device := getDeviceInfo(id)
	
	client.listener.callback(DeviceEvent{
		Type:     ActiveDeviceChanged,
		DeviceID: id,
		Device:   device,
	})
	
	return 0
}

func notificationClientOnPropertyValueChanged(this unsafe.Pointer, deviceID *uint16, key PROPERTYKEY) uintptr {
	client := (*NotificationClient)(this)
	// Not implemented for now
	return 0
}

// getDeviceInfo retrieves information about a device by ID
func getDeviceInfo(deviceID string) *AudioDevice {
	enumerator, err := CreateDeviceEnumerator()
	if err != nil {
		return nil
	}
	defer enumerator.Release()
	
	// Get device by ID
	device, err := enumerator.GetDevice(deviceID)
	if err != nil {
		return nil
	}
	defer device.Release()
	
	name, _ := GetDeviceName(device)
	state, _ := device.GetState()
	
	// Determine if it's input or output by trying to get it as default
	isOutput := false
	isInput := false
	isActive := false
	
	// Check if it's the default output
	defaultOutput, err := enumerator.GetDefaultAudioEndpoint(eRender, eConsole)
	if err == nil {
		defaultID, _ := defaultOutput.GetId()
		if defaultID == deviceID {
			isOutput = true
			isActive = true
		}
		defaultOutput.Release()
	}
	
	// Check if it's the default input
	defaultInput, err := enumerator.GetDefaultAudioEndpoint(eCapture, eConsole)
	if err == nil {
		defaultID, _ := defaultInput.GetId()
		if defaultID == deviceID {
			isInput = true
			isActive = true
		}
		defaultInput.Release()
	}
	
	// If not default, try to determine type by enumeration
	if !isOutput && !isInput {
		// Check in output devices
		outputCollection, err := enumerator.EnumAudioEndpoints(eRender, DEVICE_STATEMASK_ALL)
		if err == nil {
			count, _ := outputCollection.GetCount()
			for i := uint32(0); i < count; i++ {
				d, err := outputCollection.Item(i)
				if err == nil {
					id, _ := d.GetId()
					if id == deviceID {
						isOutput = true
					}
					d.Release()
				}
			}
			outputCollection.Release()
		}
		
		// Check in input devices
		inputCollection, err := enumerator.EnumAudioEndpoints(eCapture, DEVICE_STATEMASK_ALL)
		if err == nil {
			count, _ := inputCollection.GetCount()
			for i := uint32(0); i < count; i++ {
				d, err := inputCollection.Item(i)
				if err == nil {
					id, _ := d.GetId()
					if id == deviceID {
						isInput = true
					}
					d.Release()
				}
			}
			inputCollection.Release()
		}
	}
	
	return &AudioDevice{
		ID:          deviceID,
		Name:        name,
		IsInput:     isInput,
		IsOutput:    isOutput,
		IsActive:    isActive,
		IsConnected: state == DEVICE_STATE_ACTIVE,
	}
}