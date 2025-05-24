package audiocontrol

// AudioDevice represents an audio device on the system
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

// ListAudioDevices enumerates all audio devices on the system
func ListAudioDevices() ([]AudioDevice, error) {
	return listAudioDevices()
}

// GetActiveOutputDevice returns the currently active output device
func GetActiveOutputDevice() (AudioDevice, error) {
	return getActiveOutputDevice()
}

// SetActiveOutputDevice sets the active output device by ID
func SetActiveOutputDevice(deviceID string) error {
	return setActiveOutputDevice(deviceID)
}

// OnDeviceChange registers a callback for audio device events
func OnDeviceChange(callback func(Event)) {
	onDeviceChange(callback)
}
