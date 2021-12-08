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

// A MountStatsNFS is a MountStats implementation for NFSv3 and v4 mounts.
type MountStatsNFS struct {
	// The version of statistics provided.
	StatVersion string
	// The age of the NFS mount.
	Age time.Duration
	// Statistics related to byte counters for various operations.
	Bytes NFSBytesStats
	// Statistics related to various NFS event occurrences.
	Events NFSEventsStats
	// Statistics broken down by filesystem operation.
	Operations []NFSOperationStats
	// Statistics about the NFS RPC transport.
	Transport NFSTransportStats
}

// mountStats implements MountStats.
func (m MountStatsNFS) mountStats() {}

// A NFSBytesStats contains statistics about the number of bytes read and written
// by an NFS client to and from an NFS server.
type NFSBytesStats struct {
	// Number of bytes read using the read() syscall.
	Read uint64
	// Number of bytes written using the write() syscall.
	Write uint64
	// Number of bytes read using the read() syscall in O_DIRECT mode.
	DirectRead uint64
	// Number of bytes written using the write() syscall in O_DIRECT mode.
	DirectWrite uint64
	// Number of bytes read from the NFS server, in total.
	ReadTotal uint64
	// Number of bytes written to the NFS server, in total.
	WriteTotal uint64
	// Number of pages read directly via mmap()'d files.
	ReadPages uint64
	// Number of pages written directly via mmap()'d files.
	WritePages uint64
}

// A NFSEventsStats contains statistics about NFS event occurrences.
type NFSEventsStats struct {
	// Number of times cached inode attributes are re-validated from the server.
	InodeRevalidate uint64
	// Number of times cached dentry nodes are re-validated from the server.
	DnodeRevalidate uint64
	// Number of times an inode cache is cleared.
	DataInvalidate uint64
	// Number of times cached inode attributes are invalidated.
	AttributeInvalidate uint64
	// Number of times files or directories have been open()'d.
	VFSOpen uint64
	// Number of times a directory lookup has occurred.
	VFSLookup uint64
	// Number of times permissions have been checked.
	VFSAccess uint64
	// Number of updates (and potential writes) to pages.
	VFSUpdatePage uint64
	// Number of pages read directly via mmap()'d files.
	VFSReadPage uint64
	// Number of times a group of pages have been read.
	VFSReadPages uint64
	// Number of pages written directly via mmap()'d files.
	VFSWritePage uint64
	// Number of times a group of pages have been written.
	VFSWritePages uint64
	// Number of times directory entries have been read with getdents().
	VFSGetdents uint64
	// Number of times attributes have been set on inodes.
	VFSSetattr uint64
	// Number of pending writes that have been forcefully flushed to the server.
	VFSFlush uint64
	// Number of times fsync() has been called on directories and files.
	VFSFsync uint64
	// Number of times locking has been attempted on a file.
	VFSLock uint64
	// Number of times files have been closed and released.
	VFSFileRelease uint64
	// Unknown.  Possibly unused.
	CongestionWait uint64
	// Number of times files have been truncated.
	Truncation uint64
	// Number of times a file has been grown due to writes beyond its existing end.
	WriteExtension uint64
	// Number of times a file was removed while still open by another process.
	SillyRename uint64
	// Number of times the NFS server gave less data than expected while reading.
	ShortRead uint64
	// Number of times the NFS server wrote less data than expected while writing.
	ShortWrite uint64
	// Number of times the NFS server indicated EJUKEBOX; retrieving data from
	// offline storage.
	JukeboxDelay uint64
	// Number of NFS v4.1+ pNFS reads.
	PNFSRead uint64
	// Number of NFS v4.1+ pNFS writes.
	PNFSWrite uint64
}

// A NFSOperationStats contains statistics for a single operation.
type NFSOperationStats struct {
	// The name of the operation.
	Operation string
	// Number of requests performed for this operation.
	Requests uint64
	// Number of times an actual RPC request has been transmitted for this operation.
	Transmissions uint64
	// Number of times a request has had a major timeout.
	MajorTimeouts uint64
	// Number of bytes sent for this operation, including RPC headers and payload.
	BytesSent uint64
	// Number of bytes received for this operation, including RPC headers and payload.
	BytesReceived uint64
	// Duration all requests spent queued for transmission before they were sent.
	CumulativeQueueTime time.Duration
	// Duration it took to get a reply back after the request was transmitted.
	CumulativeTotalResponseTime time.Duration
	// Duration from when a request was enqueued to when it was completely handled.
	CumulativeTotalRequestTime time.Duration
}

// A NFSTransportStats contains statistics for the NFS mount RPC requests and
// responses.
type NFSTransportStats struct {
	// The local port used for the NFS mount.
	Port uint64
	// Number of times the client has had to establish a connection from scratch
	// to the NFS server.
	Bind uint64
	// Number of times the client has made a TCP connection to the NFS server.
	Connect uint64
	// Duration (in jiffies, a kernel internal unit of time) the NFS mount has
	// spent waiting for connections to the server to be established.
	ConnectIdleTime uint64
	// Duration since the NFS mount last saw any RPC traffic.
	IdleTime time.Duration
	// Number of RPC requests for this mount sent to the NFS server.
	Sends uint64
	// Number of RPC responses for this mount received from the NFS server.
	Receives uint64
	// Number of times the NFS server sent a response with a transaction ID
	// unknown to this client.
	BadTransactionIDs uint64
	// A running counter, incremented on each request as the current difference
	// ebetween sends and receives.
	CumulativeActiveRequests uint64
	// A running counter, incremented on each request by the current backlog
	// queue size.
	CumulativeBacklog uint64

	// Stats below only available with stat version 1.1.

	// Maximum number of simultaneously active RPC requests ever used.
	MaximumRPCSlotsUsed uint64
	// A running counter, incremented on each request as the current size of the
	// sending queue.
	CumulativeSendingQueue uint64
	// A running counter, incremented on each request as the current size of the
	// pending queue.
	CumulativePendingQueue uint64
}

// parseMountStats parses a /proc/[pid]/mountstats file and returns a slice
// of Mount structures containing detailed information about each mount.
// If available, statistics for each mount are parsed as well.
func parseMountStats(r io.Reader) ([]*Mount, error) {
	const (
		device            = "device"
		statVersionPrefix = "statvers="

		nfs3Type = "nfs"
		nfs4Type = "nfs4"
	)

	var mounts []*Mount

	s := bufio.NewScanner(r)
	for s.Scan() {
		// Only look for device entries in this function
		ss := strings.Fields(string(s.Bytes()))
		if len(ss) == 0 || ss[0] != device {
			continue
		}

		m, err := parseMount(ss)
		if err != nil {
			return nil, err
		}

		// Does this mount also possess statistics information?
		if len(ss) > deviceEntryLen {
