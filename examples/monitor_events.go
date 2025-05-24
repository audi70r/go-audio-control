package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	audiocontrol "github.com/audi70r/go-audio-control"
)

func main() {
	fmt.Println("Monitoring audio device events...")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()

	// Set up event monitoring
	audiocontrol.OnDeviceChange(func(event audiocontrol.Event) {
		switch event.Type {
		case audiocontrol.DeviceAdded:
			fmt.Printf("Device Added: %s (%s)\n", event.Info.Name, event.DeviceID)
		case audiocontrol.DeviceRemoved:
			fmt.Printf("Device Removed: %s (%s)\n", event.Info.Name, event.DeviceID)
		case audiocontrol.ActiveDeviceChanged:
			fmt.Printf("Active Device Changed: %s (%s)\n", event.Info.Name, event.DeviceID)
		case audiocontrol.DeviceDisconnected:
			fmt.Printf("Device Disconnected: %s (%s)\n", event.Info.Name, event.DeviceID)
		}
	})

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nExiting...")
}
