package main

import (
	"fmt"
	"log"

	audiocontrol "github.com/audi70r/go-audio-control"
)

func main() {
	// List all audio devices
	devices, err := audiocontrol.ListAudioDevices()
	if err != nil {
		log.Fatal("Failed to list audio devices:", err)
	}

	fmt.Println("Audio Devices:")
	fmt.Println("==============")
	for _, device := range devices {
		fmt.Printf("ID: %s\n", device.ID)
		fmt.Printf("  Name: %s\n", device.Name)
		fmt.Printf("  Input: %v, Output: %v\n", device.IsInput, device.IsOutput)
		fmt.Printf("  Active: %v, Connected: %v\n", device.IsActive, device.IsConnected)
		fmt.Println()
	}

	// Get active output device
	activeDevice, err := audiocontrol.GetActiveOutputDevice()
	if err != nil {
		log.Println("Failed to get active output device:", err)
	} else {
		fmt.Println("Active Output Device:")
		fmt.Println("====================")
		fmt.Printf("ID: %s\n", activeDevice.ID)
		fmt.Printf("Name: %s\n", activeDevice.Name)
		fmt.Println()
	}
}