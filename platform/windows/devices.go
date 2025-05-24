//go:build windows
// +build windows

package windows

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"

	"github.com/go-ole/go-ole"
)

// COM GUIDs
var (
	CLSID_MMDeviceEnumerator = &ole.GUID{0xBCDE0395, 0xE52F, 0x467C, [8]byte{0x8E, 0x3D, 0xC4, 0x57, 0x92, 0x91, 0x69, 0x2E}}
	IID_IMMDeviceEnumerator  = &ole.GUID{0xA95664D2, 0x9614, 0x4F35, [8]byte{0xA7, 0x46, 0xDE, 0x8D, 0xB6, 0x36, 0x17, 0xE6}}
	IID_IMMDevice            = &ole.GUID{0xD666063F, 0x1587, 0x4E43, [8]byte{0x81, 0xF1, 0xB9, 0x48, 0xE8, 0x07, 0x36, 0x3F}}
	IID_IMMDeviceCollection  = &ole.GUID{0x0BD7A1BE, 0x7DA1, 0x44AB, [8]byte{0x82, 0x41, 0x5F, 0x9D, 0xCF, 0x94, 0x28, 0x2E}}
	IID_IPropertyStore       = &ole.GUID{0x886D8EEB, 0x8CF2, 0x4446, [8]byte{0x8D, 0x02, 0xCD, 0xBA, 0x1D, 0xBD, 0xCF, 0x99}}
	IID_IMMNotificationClient = &ole.GUID{0x7991EEC9, 0x7E89, 0x4D85, [8]byte{0x83, 0x90, 0x6C, 0x70, 0x3C, 0xEC, 0x60, 0xC0}}
	IID_IPolicyConfig        = &ole.GUID{0xF8679F50, 0x850A, 0x41CF, [8]byte{0x9C, 0x72, 0x43, 0x0F, 0x29, 0x02, 0x90, 0xC8}}
	CLSID_PolicyConfigClient = &ole.GUID{0x870AF99C, 0x171D, 0x4F9E, [8]byte{0xAF, 0x0D, 0xE6, 0x3D, 0xF4, 0x0C, 0x2B, 0xC9}}
)

// Property keys
type PROPERTYKEY struct {
	fmtid ole.GUID
	pid   uint32
}

var (
	PKEY_Device_FriendlyName = PROPERTYKEY{
		fmtid: ole.GUID{0xA45C254E, 0xDF1C, 0x4EFD, [8]byte{0x80, 0x20, 0x67, 0xD1, 0x46, 0xA8, 0x50, 0xE0}},
		pid:   14,
	}
	PKEY_Device_DeviceDesc = PROPERTYKEY{
		fmtid: ole.GUID{0xA45C254E, 0xDF1C, 0x4EFD, [8]byte{0x80, 0x20, 0x67, 0xD1, 0x46, 0xA8, 0x50, 0xE0}},
		pid:   2,
	}
)

// EDataFlow enum
type EDataFlow uint32

const (
	eRender EDataFlow = iota
	eCapture
	eAll
	EDataFlow_enum_count
)

// ERole enum
type ERole uint32

const (
	eConsole ERole = iota
	eMultimedia
	eCommunications
	ERole_enum_count
)

// DEVICE_STATE constants
const (
	DEVICE_STATE_ACTIVE     = 0x00000001
	DEVICE_STATE_DISABLED   = 0x00000002
	DEVICE_STATE_NOTPRESENT = 0x00000004
	DEVICE_STATE_UNPLUGGED  = 0x00000008
	DEVICE_STATEMASK_ALL    = 0x0000000F
)

// PropVariant structure
type PropVariant struct {
	Vt   uint16
	_    uint16
	_    uint32
	Data [16]byte
}

// IMMDeviceEnumerator interface
type IMMDeviceEnumerator struct {
	vtbl *IMMDeviceEnumeratorVtbl
}

type IMMDeviceEnumeratorVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	EnumAudioEndpoints          uintptr
	GetDefaultAudioEndpoint     uintptr
	GetDevice                   uintptr
	RegisterEndpointNotificationCallback   uintptr
	UnregisterEndpointNotificationCallback uintptr
}

// IMMDeviceCollection interface
type IMMDeviceCollection struct {
	vtbl *IMMDeviceCollectionVtbl
}

type IMMDeviceCollectionVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetCount uintptr
	Item     uintptr
}

// IMMDevice interface
type IMMDevice struct {
	vtbl *IMMDeviceVtbl
}

type IMMDeviceVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	Activate          uintptr
	OpenPropertyStore uintptr
	GetId             uintptr
	GetState          uintptr
}

// IPropertyStore interface
type IPropertyStore struct {
	vtbl *IPropertyStoreVtbl
}

type IPropertyStoreVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetCount uintptr
	GetAt    uintptr
	GetValue uintptr
	SetValue uintptr
	Commit   uintptr
}

// IPolicyConfig interface for setting default devices
type IPolicyConfig struct {
	vtbl *IPolicyConfigVtbl
}

type IPolicyConfigVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr

	GetMixFormat                     uintptr
	GetDeviceFormat                  uintptr
	ResetDeviceFormat                uintptr
	SetDeviceFormat                  uintptr
	GetProcessingPeriod              uintptr
	SetProcessingPeriod              uintptr
	GetShareMode                     uintptr
	SetShareMode                     uintptr
	GetPropertyValue                 uintptr
	SetPropertyValue                 uintptr
	SetDefaultEndpoint               uintptr
	SetEndpointVisibility            uintptr
}

// Helper functions
func (e *IMMDeviceEnumerator) Release() {
	syscall.Syscall(e.vtbl.Release, 1, uintptr(unsafe.Pointer(e)), 0, 0)
}

func (c *IMMDeviceCollection) Release() {
	syscall.Syscall(c.vtbl.Release, 1, uintptr(unsafe.Pointer(c)), 0, 0)
}

func (d *IMMDevice) Release() {
	syscall.Syscall(d.vtbl.Release, 1, uintptr(unsafe.Pointer(d)), 0, 0)
}

func (s *IPropertyStore) Release() {
	syscall.Syscall(s.vtbl.Release, 1, uintptr(unsafe.Pointer(s)), 0, 0)
}

func (p *IPolicyConfig) Release() {
	syscall.Syscall(p.vtbl.Release, 1, uintptr(unsafe.Pointer(p)), 0, 0)
}

// AudioDevice represents an audio device
type AudioDevice struct {
	ID          string
	Name        string
	IsInput     bool
	IsOutput    bool
	IsActive    bool
	IsConnected bool
}

// Initialize COM
func InitializeCOM() error {
	return ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
}

// Uninitialize COM
func UninitializeCOM() {
	ole.CoUninitialize()
}

// CreateDeviceEnumerator creates an IMMDeviceEnumerator instance
func CreateDeviceEnumerator() (*IMMDeviceEnumerator, error) {
	var enumerator *IMMDeviceEnumerator
	hr := ole.CoCreateInstance(
		CLSID_MMDeviceEnumerator,
		nil,
		ole.CLSCTX_ALL,
		IID_IMMDeviceEnumerator,
		(*unsafe.Pointer)(unsafe.Pointer(&enumerator)),
	)
	if hr != 0 {
		return nil, ole.NewError(hr)
	}
	return enumerator, nil
}

// EnumAudioEndpoints enumerates audio endpoints
func (e *IMMDeviceEnumerator) EnumAudioEndpoints(dataFlow EDataFlow, stateMask uint32) (*IMMDeviceCollection, error) {
	var collection *IMMDeviceCollection
	hr, _, _ := syscall.Syscall6(
		e.vtbl.EnumAudioEndpoints,
		4,
		uintptr(unsafe.Pointer(e)),
		uintptr(dataFlow),
		uintptr(stateMask),
		uintptr(unsafe.Pointer(&collection)),
		0,
		0,
	)
	if hr != 0 {
		return nil, ole.NewError(hr)
	}
	return collection, nil
}

// GetDefaultAudioEndpoint gets the default audio endpoint
func (e *IMMDeviceEnumerator) GetDefaultAudioEndpoint(dataFlow EDataFlow, role ERole) (*IMMDevice, error) {
	var device *IMMDevice
	hr, _, _ := syscall.Syscall6(
		e.vtbl.GetDefaultAudioEndpoint,
		4,
		uintptr(unsafe.Pointer(e)),
		uintptr(dataFlow),
		uintptr(role),
		uintptr(unsafe.Pointer(&device)),
		0,
		0,
	)
	if hr != 0 {
		return nil, ole.NewError(hr)
	}
	return device, nil
}

// GetCount gets the number of devices in the collection
func (c *IMMDeviceCollection) GetCount() (uint32, error) {
	var count uint32
	hr, _, _ := syscall.Syscall(
		c.vtbl.GetCount,
		2,
		uintptr(unsafe.Pointer(c)),
		uintptr(unsafe.Pointer(&count)),
		0,
	)
	if hr != 0 {
		return 0, ole.NewError(hr)
	}
	return count, nil
}

// Item gets a device from the collection by index
func (c *IMMDeviceCollection) Item(index uint32) (*IMMDevice, error) {
	var device *IMMDevice
	hr, _, _ := syscall.Syscall(
		c.vtbl.Item,
		3,
		uintptr(unsafe.Pointer(c)),
		uintptr(index),
		uintptr(unsafe.Pointer(&device)),
	)
	if hr != 0 {
		return nil, ole.NewError(hr)
	}
	return device, nil
}

// GetId gets the device ID
func (d *IMMDevice) GetId() (string, error) {
	var idPtr *uint16
	hr, _, _ := syscall.Syscall(
		d.vtbl.GetId,
		2,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(&idPtr)),
		0,
	)
	if hr != 0 {
		return "", ole.NewError(hr)
	}
	defer ole.CoTaskMemFree(uintptr(unsafe.Pointer(idPtr)))
	
	return syscall.UTF16ToString((*[1024]uint16)(unsafe.Pointer(idPtr))[:]), nil
}

// GetState gets the device state
func (d *IMMDevice) GetState() (uint32, error) {
	var state uint32
	hr, _, _ := syscall.Syscall(
		d.vtbl.GetState,
		2,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(&state)),
		0,
	)
	if hr != 0 {
		return 0, ole.NewError(hr)
	}
	return state, nil
}

// OpenPropertyStore opens the property store for the device
func (d *IMMDevice) OpenPropertyStore(stgmAccess uint32) (*IPropertyStore, error) {
	var store *IPropertyStore
	hr, _, _ := syscall.Syscall(
		d.vtbl.OpenPropertyStore,
		3,
		uintptr(unsafe.Pointer(d)),
		uintptr(stgmAccess),
		uintptr(unsafe.Pointer(&store)),
	)
	if hr != 0 {
		return nil, ole.NewError(hr)
	}
	return store, nil
}

// GetValue gets a property value
func (s *IPropertyStore) GetValue(key *PROPERTYKEY, propVar *PropVariant) error {
	hr, _, _ := syscall.Syscall(
		s.vtbl.GetValue,
		3,
		uintptr(unsafe.Pointer(s)),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(propVar)),
	)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}

// GetDeviceName gets the friendly name of the device
func GetDeviceName(device *IMMDevice) (string, error) {
	const STGM_READ = 0
	
	store, err := device.OpenPropertyStore(STGM_READ)
	if err != nil {
		return "", err
	}
	defer store.Release()
	
	var propVar PropVariant
	err = store.GetValue(&PKEY_Device_FriendlyName, &propVar)
	if err != nil {
		return "", err
	}
	
	// VT_LPWSTR = 31
	if propVar.Vt != 31 {
		return "", errors.New("unexpected property type")
	}
	
	// Extract the string pointer from PropVariant
	strPtr := *(**uint16)(unsafe.Pointer(&propVar.Data[0]))
	if strPtr == nil {
		return "", errors.New("null string pointer")
	}
	
	// Convert to Go string
	name := syscall.UTF16ToString((*[1024]uint16)(unsafe.Pointer(strPtr))[:])
	
	// Clear the PropVariant
	PropVariantClear(&propVar)
	
	return name, nil
}

// ListAudioDevices enumerates all audio devices
func ListAudioDevices() ([]AudioDevice, error) {
	enumerator, err := CreateDeviceEnumerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create device enumerator: %w", err)
	}
	defer enumerator.Release()
	
	var devices []AudioDevice
	
	// Get default output device ID
	defaultOutputID := ""
	defaultOutput, err := enumerator.GetDefaultAudioEndpoint(eRender, eConsole)
	if err == nil {
		defaultOutputID, _ = defaultOutput.GetId()
		defaultOutput.Release()
	}
	
	// Get default input device ID
	defaultInputID := ""
	defaultInput, err := enumerator.GetDefaultAudioEndpoint(eCapture, eConsole)
	if err == nil {
		defaultInputID, _ = defaultInput.GetId()
		defaultInput.Release()
	}
	
	// Enumerate output devices
	outputCollection, err := enumerator.EnumAudioEndpoints(eRender, DEVICE_STATE_ACTIVE)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate output devices: %w", err)
	}
	defer outputCollection.Release()
	
	outputCount, err := outputCollection.GetCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get output device count: %w", err)
	}
	
	for i := uint32(0); i < outputCount; i++ {
		device, err := outputCollection.Item(i)
		if err != nil {
			continue
		}
		
		id, err := device.GetId()
		if err != nil {
			device.Release()
			continue
		}
		
		name, err := GetDeviceName(device)
		if err != nil {
			device.Release()
			continue
		}
		
		state, _ := device.GetState()
		
		devices = append(devices, AudioDevice{
			ID:          id,
			Name:        name,
			IsInput:     false,
			IsOutput:    true,
			IsActive:    id == defaultOutputID,
			IsConnected: state == DEVICE_STATE_ACTIVE,
		})
		
		device.Release()
	}
	
	// Enumerate input devices
	inputCollection, err := enumerator.EnumAudioEndpoints(eCapture, DEVICE_STATE_ACTIVE)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate input devices: %w", err)
	}
	defer inputCollection.Release()
	
	inputCount, err := inputCollection.GetCount()
	if err != nil {
		return nil, fmt.Errorf("failed to get input device count: %w", err)
	}
	
	for i := uint32(0); i < inputCount; i++ {
		device, err := inputCollection.Item(i)
		if err != nil {
			continue
		}
		
		id, err := device.GetId()
		if err != nil {
			device.Release()
			continue
		}
		
		name, err := GetDeviceName(device)
		if err != nil {
			device.Release()
			continue
		}
		
		state, _ := device.GetState()
		
		devices = append(devices, AudioDevice{
			ID:          id,
			Name:        name,
			IsInput:     true,
			IsOutput:    false,
			IsActive:    id == defaultInputID,
			IsConnected: state == DEVICE_STATE_ACTIVE,
		})
		
		device.Release()
	}
	
	return devices, nil
}

// GetActiveOutputDevice returns the currently active output device
func GetActiveOutputDevice() (*AudioDevice, error) {
	enumerator, err := CreateDeviceEnumerator()
	if err != nil {
		return nil, fmt.Errorf("failed to create device enumerator: %w", err)
	}
	defer enumerator.Release()
	
	device, err := enumerator.GetDefaultAudioEndpoint(eRender, eConsole)
	if err != nil {
		return nil, fmt.Errorf("failed to get default output device: %w", err)
	}
	defer device.Release()
	
	id, err := device.GetId()
	if err != nil {
		return nil, fmt.Errorf("failed to get device ID: %w", err)
	}
	
	name, err := GetDeviceName(device)
	if err != nil {
		return nil, fmt.Errorf("failed to get device name: %w", err)
	}
	
	state, _ := device.GetState()
	
	return &AudioDevice{
		ID:          id,
		Name:        name,
		IsInput:     false,
		IsOutput:    true,
		IsActive:    true,
		IsConnected: state == DEVICE_STATE_ACTIVE,
	}, nil
}

// CreatePolicyConfig creates an IPolicyConfig instance
func CreatePolicyConfig() (*IPolicyConfig, error) {
	var policyConfig *IPolicyConfig
	hr := ole.CoCreateInstance(
		CLSID_PolicyConfigClient,
		nil,
		ole.CLSCTX_ALL,
		IID_IPolicyConfig,
		(*unsafe.Pointer)(unsafe.Pointer(&policyConfig)),
	)
	if hr != 0 {
		return nil, ole.NewError(hr)
	}
	return policyConfig, nil
}

// SetDefaultEndpoint sets the default audio endpoint
func (p *IPolicyConfig) SetDefaultEndpoint(deviceID string, role ERole) error {
	deviceIDPtr, err := syscall.UTF16PtrFromString(deviceID)
	if err != nil {
		return err
	}
	
	hr, _, _ := syscall.Syscall(
		p.vtbl.SetDefaultEndpoint,
		3,
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(deviceIDPtr)),
		uintptr(role),
	)
	if hr != 0 {
		return ole.NewError(hr)
	}
	return nil
}

// PropVariantClear clears a PropVariant structure
func PropVariantClear(pv *PropVariant) {
	modOle32 := syscall.NewLazyDLL("ole32.dll")
	procPropVariantClear := modOle32.NewProc("PropVariantClear")
	procPropVariantClear.Call(uintptr(unsafe.Pointer(pv)))
}

// SetActiveOutputDevice sets the active output device by ID
func SetActiveOutputDevice(deviceID string) error {
	policyConfig, err := CreatePolicyConfig()
	if err != nil {
		return fmt.Errorf("failed to create policy config: %w", err)
	}
	defer policyConfig.Release()
	
	// Set for all roles
	if err := policyConfig.SetDefaultEndpoint(deviceID, eConsole); err != nil {
		return fmt.Errorf("failed to set console endpoint: %w", err)
	}
	if err := policyConfig.SetDefaultEndpoint(deviceID, eMultimedia); err != nil {
		return fmt.Errorf("failed to set multimedia endpoint: %w", err)
	}
	if err := policyConfig.SetDefaultEndpoint(deviceID, eCommunications); err != nil {
		return fmt.Errorf("failed to set communications endpoint: %w", err)
	}
	
	return nil
}