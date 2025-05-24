//go:build windows
// +build windows

package audiocontrol

import "github.com/audi70r/go-audio-control/platform/windows"

// Platform-specific function implementations

func listAudioDevices() ([]AudioDevice, error) {
	devices, err := windows.ListAudioDevices()
	if err != nil {
		return nil, err
	}

	// Convert platform-specific devices to generic AudioDevice
	result := make([]AudioDevice, len(devices))
	for i, d := range devices {
		result[i] = AudioDevice{
			ID:          d.ID,
			Name:        d.Name,
			IsInput:     d.IsInput,
			IsOutput:    d.IsOutput,
			IsActive:    d.IsActive,
			IsConnected: d.IsConnected,
		}
	}

	return result, nil
}

func getActiveOutputDevice() (AudioDevice, error) {
	device, err := windows.GetActiveOutputDevice()
	if err != nil {
		return AudioDevice{}, err
	}

	return AudioDevice{
		ID:          device.ID,
		Name:        device.Name,
		IsInput:     device.IsInput,
		IsOutput:    device.IsOutput,
		IsActive:    device.IsActive,
		IsConnected: device.IsConnected,
	}, nil
}

func setActiveOutputDevice(deviceID string) error {
	return windows.SetActiveOutputDevice(deviceID)
}

func onDeviceChange(callback func(Event)) {
	// Create listener
	listener, err := windows.NewDeviceListener()
	if err != nil {
		// Handle error - for now just return
		return
	}

	// Start listening with callback wrapper
	listener.Start(func(e windows.DeviceEvent) {
		event := Event{
			Type:     EventType(e.Type),
			DeviceID: e.DeviceID,
		}

		if e.Device != nil {
			event.Info = &AudioDevice{
				ID:          e.Device.ID,
				Name:        e.Device.Name,
				IsInput:     e.Device.IsInput,
				IsOutput:    e.Device.IsOutput,
				IsActive:    e.Device.IsActive,
				IsConnected: e.Device.IsConnected,
			}
		}

		callback(event)
	})
}
