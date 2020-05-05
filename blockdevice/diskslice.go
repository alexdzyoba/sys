package blockdevice

import (
	"os"
	"strings"

	"github.com/pkg/errors"
)

// DiskSlice represents a collection of disks.
//
// It's a distinct type from []Disk because of 2 things:
// 1. For convenient discovery of all disks in the system.
// 2. To implement sort.Interface for sorting by size.
type DiskSlice []Disk

// NewDiskSlice creates DiskSlice from all disks in the system.
// Disks are discovered by quering sysfs hierarchy.
func NewDiskSlice() (DiskSlice, error) {
	root, err := os.Open(sysfsBlockRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open %v", sysfsBlockRoot)
	}

	diskNames, err := root.Readdirnames(-1)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read directory %v", root.Name())
	}

	return NewDiskSliceFromPaths(diskNames)
}

// NewDiskSliceFromPaths creates DiskSlice from a given slice of paths.
// Each path will be checked to exist in the system.
// Paths can be provided as base device names like ["sda", "sdb"] or with any
// prefix like ["/dev/sda", "/dev/sdb"] - only base name will be used.
func NewDiskSliceFromPaths(paths []string) (DiskSlice, error) {
	ds := make(DiskSlice, 0, len(paths))
	for _, p := range paths {
		d, err := NewDisk(p)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create disk")
		}
		ds = append(ds, *d)
	}

	return ds, nil
}

// Implement sort.Interface to allow sorting of the disk slice.
// Disk slice is sorted by size.
func (ds DiskSlice) Len() int           { return len(ds) }
func (ds DiskSlice) Less(i, j int) bool { return ds[i].Size < ds[j].Size }
func (ds DiskSlice) Swap(i, j int)      { ds[i], ds[j] = ds[j], ds[i] }

func (ds DiskSlice) String() string {
	return strings.Join(ds.StringSlice(), " ")
}

func (ds DiskSlice) StringSlice() []string {
	var ss []string
	for _, d := range ds {
		ss = append(ss, d.DeviceName())
	}

	return ss
}
