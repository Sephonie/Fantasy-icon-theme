package procfs

// While implementing parsing of /proc/[pid]/mountstats, this blog was used
// heavily as a reference:
//   https://utcc.utoronto.ca/~cks/space/blog/linux/NFSMountstatsIndex
//
// Special thanks to Chris Siebenmann for all of his posts explaining the
// various statistics available for NFS.

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Constants shared between multiple functions.
const (
	deviceEntryLen = 8

	fieldBytesLen  = 8
	fieldEventsLen = 27

	statVersion10 = "1.0"
	statVersion11 = "1.1"

	fieldTransport10Len = 10
	fieldTransport11Len = 13
)

// A Mount is a device mount parsed from /proc/[pid]/mountstats.
type Mount struct {
	// Name of the device.
	Device string
	// The mount point of the device.
	Mount string
	// The filesystem type used by the device.
	Type string
	// If available additional statistics related to this Mount.
	// Use a type assertion to determine if additional statistics are available.
	Stats MountStats
}

// A MountStats is a type which contains detailed statistics for a specific
// type of Mount.
type MountStats interface {
	mountStats()
}

