package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	audiocontrol "github.com/audi70r/go-audio-control"
)

func main() {
	// Create audio controller
	controller, err := audiocontrol.New()
	if err != nil {
		log.Fatal("Failed to create audio controller:", err)
	}

	// Start monitoring device changes
	fmt.Println("Monitoring audio device changes...")
	fmt.Println("Press Ctrl+C to exit")
	fmt.Println()

	err = controller.MonitorDeviceChanges(func(event audiocontrol.DeviceEvent) {
		var eventType string
		switch event.Type {
		case audiocontrol.EventTypeDeviceAdded:
			eventType = "Device Added"
		case audiocontrol.EventTypeDeviceRemoved:
			eventType = "Device Removed"
		case audiocontrol.EventTypeDefaultChanged:
			eventType = "Default Device Changed"
		case audiocontrol.EventTypeVolumeChanged:
			eventType = "Volume Changed"
		}

		deviceType := "Output"
		if event.Device.Type == audiocontrol.DeviceTypeInput {
			deviceType = "Input"
		}

		fmt.Printf("[%s] %s: %s (%s)\n", 
			event.Timestamp.Format("15:04:05"),
			eventType,
			event.Device.Name,
			deviceType,
		)
	})

	if err != nil {
		log.Fatal("Failed to start monitoring:", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	// Stop monitoring
	fmt.Println("\nStopping monitor...")
	err = controller.StopMonitoring()
	if err != nil {
		log.Fatal("Failed to stop monitoring:", err)
	}

	fmt.Println("Monitor stopped.")
}