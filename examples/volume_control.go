package main

import (
	"fmt"
	"log"

	audiocontrol "github.com/audi70r/go-audio-control"
)

func main() {
	// Create audio controller
	controller, err := audiocontrol.New()
	if err != nil {
		log.Fatal("Failed to create audio controller:", err)
	}

	// Get default output device
	device, err := controller.GetDefaultDevice(audiocontrol.DeviceTypeOutput)
	if err != nil {
		log.Fatal("Failed to get default output device:", err)
	}

	fmt.Printf("Controlling volume for: %s\n", device.Name)

	// Get current volume
	currentVolume, err := controller.GetVolume(device.ID)
	if err != nil {
		log.Fatal("Failed to get volume:", err)
	}
	fmt.Printf("Current volume: %.0f%%\n", currentVolume*100)

	// Get mute state
	isMuted, err := controller.GetMute(device.ID)
	if err != nil {
		log.Fatal("Failed to get mute state:", err)
	}
	fmt.Printf("Muted: %v\n", isMuted)

	// Set volume to 50%
	fmt.Println("\nSetting volume to 50%...")
	err = controller.SetVolume(device.ID, 0.5)
	if err != nil {
		log.Fatal("Failed to set volume:", err)
	}

	// Toggle mute
	fmt.Println("Toggling mute...")
	err = controller.SetMute(device.ID, !isMuted)
	if err != nil {
		log.Fatal("Failed to toggle mute:", err)
	}

	// Verify changes
	newVolume, _ := controller.GetVolume(device.ID)
	newMute, _ := controller.GetMute(device.ID)
	fmt.Printf("\nNew volume: %.0f%%\n", newVolume*100)
	fmt.Printf("New mute state: %v\n", newMute)
}