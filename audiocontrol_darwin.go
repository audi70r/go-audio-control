//go:build darwin
// +build darwin

package audiocontrol

import darwin "github.com/audi70r/go-audio-control/platform/darwin"

// Platform-specific function implementations

func listAudioDevices() ([]AudioDevice, error) {
	devices, err := darwin.ListAudioDevices()
	if err != nil {
		return nil, err
	}

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
	device, err := darwin.GetActiveOutputDevice()
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
	return darwin.SetActiveOutputDevice(deviceID)
}

func onDeviceChange(callback func(Event)) {
	darwin.OnDeviceChange(func(e darwin.Event) {
		var info *AudioDevice
		if e.Info != nil {
			info = &AudioDevice{
				ID:          e.Info.ID,
				Name:        e.Info.Name,
				IsInput:     e.Info.IsInput,
				IsOutput:    e.Info.IsOutput,
				IsActive:    e.Info.IsActive,
				IsConnected: e.Info.IsConnected,
			}
		}

		callback(Event{
			Type:     EventType(e.Type),
			DeviceID: e.DeviceID,
			Info:     info,
		})
	})
}
