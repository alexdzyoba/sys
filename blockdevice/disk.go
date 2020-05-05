package blockdevice

import (
	"fmt"
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

// Disk represents a disk device
type Disk struct {
	Name string
	Size uint64
}

// NewDisk creates a Disk type.
// The disk properties are discovered from sysfs.
func NewDisk(diskPath string) (*Disk, error) {
	name := path.Base(diskPath)

	sysfsName := path.Join(sysfsBlockRoot, name)
	if _, err := os.Stat(sysfsName); os.IsNotExist(err) {
		return nil, errors.Errorf("device %s does not exist", sysfsName)
	}

	// Discover disk size from /sys/block/<name>/size
	sizeFilePath := path.Join(sysfsName, "size")
	sizeContent, err := ioutil.ReadFile(sizeFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %v for size", err)
	}

	sizeString := strings.Trim(string(sizeContent), "\n")
	size, err := strconv.ParseUint(sizeString, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse device size")
	}

	// Disk size in sysfs is always shown in 512 bytes sectors
	size = size * sectorSizeBytes

	return &Disk{name, size}, nil
}

func (d *Disk) DeviceName() string {
	return fmt.Sprintf("/dev/%s", d.Name)
}
