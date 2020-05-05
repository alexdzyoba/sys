package blockdevice

import (
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type BlockDevicer interface {
	DeviceName() string
}

// Attributes holds device attributes as reported by blkid
type Attributes struct {
	UUID  string
	Type  string
	Label string
}

const (
	attributeUUIDKey  = "UUID"
	attributeTypeKey  = "TYPE"
	attributeLabelKey = "LABEL"
)

// GetAttributes returns block device attributes by invoking blkid and parsing
// its output
func GetAttributes(bd BlockDevicer) (*Attributes, error) {
	// `-o export` will output block device attributes as KEY=VALUE lines
	blkid := exec.Command("blkid", "-o", "export", bd.DeviceName())
	blkid.Stderr = os.Stderr

	output, err := blkid.Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to exec blkid")
	}

	lines := strings.Split(string(output), "\n")

	var attrs Attributes
	for _, line := range lines {
		if line == "" {
			continue // skip last empty line
		}

		kv := strings.SplitN(line, "=", 2)
		key, value := kv[0], kv[1]

		switch key {
		case attributeUUIDKey:
			attrs.UUID = value

		case attributeLabelKey:
			attrs.Label = value

		case attributeTypeKey:
			attrs.Type = value
		}
	}

	return &attrs, nil
}
