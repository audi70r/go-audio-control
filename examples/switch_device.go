package main

import (
	"fmt"
	"log"
	"os"

	audiocontrol "github.com/audi70r/go-audio-control"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: switch_device <device-id>")
		fmt.Println("\nAvailable devices:")
		
		devices, err := audiocontrol.ListAudioDevices()
		if err != nil {
			log.Fatal("Failed to list devices:", err)
		}
		
		for _, device := range devices {
			if device.IsOutput {
				active := ""
				if device.IsActive {
					active = " (active)"
				}
				fmt.Printf("  %s - %s%s\n", device.ID, device.Name, active)
			}
		}
		os.Exit(1)
	}

	deviceID := os.Args[1]
	
	// Get current active device
	current, err := audiocontrol.GetActiveOutputDevice()
	if err == nil {
		fmt.Printf("Current active device: %s\n", current.Name)
	}

	// Switch to new device
	fmt.Printf("Switching to device: %s\n", deviceID)
	err = audiocontrol.SetActiveOutputDevice(deviceID)
	if err != nil {
		log.Fatal("Failed to switch device:", err)
	}

	// Verify the switch
	newActive, err := audiocontrol.GetActiveOutputDevice()
	if err == nil {
		fmt.Printf("New active device: %s\n", newActive.Name)
	}
}