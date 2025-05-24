//go:build !darwin && !windows
// +build !darwin,!windows

package audiocontrol

// newPlatformController returns an error for unsupported platforms
func newPlatformController() (AudioController, error) {
	return nil, ErrNotImplemented
}
