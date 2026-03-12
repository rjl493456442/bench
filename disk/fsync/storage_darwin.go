//go:build darwin

package main

// StorageInfo holds detected storage device information.
type StorageInfo struct {
	Device    string
	Model     string
	Serial    string
	Firmware  string
	Transport string
	// PCIe / NVMe link fields
	LinkSpeed    string
	LinkWidth    string
	MaxLinkSpeed string
	MaxLinkWidth string
}

// Interface returns a human-readable interface description.
func (s *StorageInfo) Interface() string { return "" }

// PCIeGen returns the PCIe generation label.
func (s *StorageInfo) PCIeGen() string { return "" }

// detectStorage is not implemented on macOS.
func detectStorage(_ string) (*StorageInfo, error) {
	return nil, nil
}
