package block

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	sysfsBlockRoot  = "/sys/block"
	sectorSizeBytes = 512
)

type Type int

const (
	TypeUnknown Type = iota
	TypeDisk
	TypeRAID
	TypeDeviceMapper
)

// Device represents a blockdevice
type Device struct {
	Name string
	Size uint64
	Type Type
}

// NewDevice creates a Device type.
// The device properties are discovered from sysfs.
func NewDevice(devicePath string) (*Device, error) {
	name := path.Base(devicePath)

	sysfsPath := path.Join(sysfsBlockRoot, name)
	if _, err := os.Stat(sysfsPath); os.IsNotExist(err) {
		return nil, errors.Errorf("device %s does not exist", sysfsPath)
	}

	// Discover device size from /sys/block/<name>/size
	sizeFilePath := path.Join(sysfsPath, "size")
	sizeContent, err := ioutil.ReadFile(sizeFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %v for size", err)
	}

	sizeString := strings.Trim(string(sizeContent), "\n")
	size, err := strconv.ParseUint(sizeString, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse device size")
	}

	// Device size in sysfs is always shown in 512 bytes sectors
	size = size * sectorSizeBytes

	typ, err := discoverDeviceType(sysfsPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse device size")
	}

	return &Device{name, size, typ}, nil
}

func discoverDeviceType(sysfsPath string) (Type, error) {
	mdPath := path.Join(sysfsPath, "md")
	mdPathExists, err := exists(mdPath)
	if err != nil {
		return TypeUnknown, errors.Errorf("failed to discover device type for %s", sysfsPath)
	}

	if mdPathExists {
		return TypeRAID, nil
	}

	dmPath := path.Join(sysfsPath, "dm")
	dmPathExists, err := exists(dmPath)
	if err != nil {
		return TypeUnknown, errors.Errorf("failed to discover device type for %s", sysfsPath)
	}

	if dmPathExists {
		return TypeDeviceMapper, nil
	}

	devicePath := path.Join(sysfsPath, "device")
	devicePathExists, err := exists(devicePath)
	if err != nil {
		return TypeUnknown, errors.Errorf("failed to discover device type for %s", sysfsPath)
	}

	if devicePathExists {
		return TypeDisk, nil
	}

	return TypeUnknown, nil
}

// exists returns whether the given path exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

// ListDevices returns block devices found in the system.
// Block devices are discovered by quering sysfs hierarchy.
func ListDevices() ([]Device, error) {
	root, err := os.Open(sysfsBlockRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %v", sysfsBlockRoot)
	}

	diskNames, err := root.Readdirnames(-1)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %v", root.Name())
	}

	return NewDevicesFromPaths(diskNames)
}

// NewDevicesFromPaths creates Device types from a given slice of paths.
// Each path will be checked to exist in the system.
// Paths can be provided as base device names like ["sda", "sdb"] or with any
// prefix like ["/dev/sda", "/dev/sdb"] - only base name will be used.
func NewDevicesFromPaths(paths []string) ([]Device, error) {
	ds := make([]Device, 0, len(paths))
	for _, p := range paths {
		d, err := NewDevice(p)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create disk")
		}
		ds = append(ds, *d)
	}

	return ds, nil
}
