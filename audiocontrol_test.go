package audiocontrol

import (
	"fmt"
	"testing"
	"time"
)

func TestListAudioDevices(t *testing.T) {
	devices, err := ListAudioDevices()
	if err != nil {
		t.Fatalf("Failed to list audio devices: %v", err)
	}

	if len(devices) == 0 {
		t.Fatal("No audio devices found")
	}

	fmt.Println("Found audio devices:")
	for _, device := range devices {
		fmt.Printf("- %s (ID: %s, Input: %v, Output: %v, Active: %v)\n",
			device.Name, device.ID, device.IsInput, device.IsOutput, device.IsActive)
	}
}

func TestGetActiveOutputDevice(t *testing.T) {
	device, err := GetActiveOutputDevice()
	if err != nil {
		t.Fatalf("Failed to get active output device: %v", err)
	}

	fmt.Printf("Active output device: %s (ID: %s)\n", device.Name, device.ID)
}

func TestDeviceMonitoring(t *testing.T) {
	eventReceived := make(chan bool, 1)

	OnDeviceChange(func(event Event) {
		fmt.Printf("Device event: Type=%v, DeviceID=%s\n", event.Type, event.DeviceID)
		if event.Info != nil {
			fmt.Printf("  Device: %s\n", event.Info.Name)
		}
		select {
		case eventReceived <- true:
		default:
		}
	})

	// Wait a bit to see if any events occur
	select {
	case <-eventReceived:
		fmt.Println("Received at least one device event")
	case <-time.After(2 * time.Second):
		fmt.Println("No device events received in 2 seconds (this is normal if no devices were changed)")
	}
}
