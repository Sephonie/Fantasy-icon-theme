// Copyright 2017 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package xfs provides access to statistics exposed by the XFS filesystem.
package xfs

// Stats contains XFS filesystem runtime statistics, parsed from
// /proc/fs/xfs/stat.
//
// The names and meanings of each statistic were taken from
// http://xfs.org/index.php/Runtime_Stats and xfs_stats.h in the Linux
// kernel source. Most counters are uint32s (same data types used in
// xfs_stats.h), but some of the "extended precision stats" are uint64s.
type Stats struct {
	// The name of the filesystem used to source these statistics.
	// If empty, this indicates aggregated statistics for all XFS
	// filesystems on the host.
	Name string

	ExtentAllocation   ExtentAllocationStats
	AllocationBTree    BTreeStats
	BlockMapping       BlockMappingStats
	BlockMapBTree      BTreeStats
	DirectoryOperation DirectoryOperationStats
	Transaction        TransactionStats
	InodeOperation     InodeOperationStats
	LogOperation       LogOperationStats
	ReadWrite          ReadWriteStats
	AttributeOperation AttributeOperationStats
	InodeClustering    InodeClusteringStats
	Vnode              VnodeStats
	Buffer             BufferStats
	ExtendedPrecision  ExtendedPrecisionStats
}

// ExtentAllocationStats contains statistics regarding XFS extent allocations.
type ExtentAllocationStats struct {
	ExtentsAllocated uint32
	BlocksAllocated  uint32
	ExtentsFreed     uint32
	BlocksFreed      uint32
}

// BTreeStats contains statistics regarding an XFS internal B-tree.
type BTreeStats struct {
	Lookups         uint32
	Compares        uint32
	RecordsInserted uint32
	RecordsDeleted  uint32
}

// BlockMappingStats contains statistics regarding XFS block maps.
type BlockMappingStats struct {
	Reads                uint32
	Writes               uint32
	Unmaps               uint32
	ExtentListInsertions uint32
	ExtentListDeletions  uint32
	ExtentListLookups    uint32
	ExtentListCompares   uint32
}

// DirectoryOperationStats contains statistics regarding XFS directory entries.
type DirectoryOperationStats str