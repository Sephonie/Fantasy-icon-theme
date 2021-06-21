package collectors

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context/ctxhttp"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/kubelet/apis/stats/v1alpha1"
)

const (
	volumeStatsCapacityBytesKey  = "kubelet_volume_stats_capacity_bytes"
	volumeStatsAvailableBytesKey = "kubelet_volume_stats_available_bytes"
	volumeStatsUsedBytesKey      = "kubelet_volume_stats_used_bytes"
	volumeStatsInodesKey         = "kubelet_volume_stats_inodes"
	volumeStatsInodesFreeKey     = "kubelet_volume_stats_inodes_free"
	volumeStatsInodesUsedKey     = "kubelet_volume_stats_inodes_used"
)

var (
	volumeStatsCapacityBytes = prometheus.NewDesc(
		volumeStatsCapacityBytesKey,
		"Capacity in bytes of the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
	volumeStatsAvailableBytes = prometheus.NewDesc(
		volumeStatsAvailableBytesKey,
		"Number of available bytes in the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
	volumeStatsUsedBytes = prometheus.NewDesc(
		volumeStatsUsedBytesKey,
		"Number of used bytes in the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
	volumeStatsInodes = prometheus.NewDesc(
		volumeStatsInodesKey,
		"Maximum number of inodes in the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
	volumeStatsInodesFree = prometheus.NewDesc(
		volumeStatsInodesFreeKey,
		"Number of free inodes in the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
	volumeStatsInodesUsed = prometheus.NewDesc(
		volumeStatsInodesUsedKey,
		"Number of used inodes in the volume",
		[]string{"namespace", "persistentvolumeclaim"}, nil,
	)
)

// volumeStatsCollector collects metrics from kubelet stats summary.
type volumeStatsCollector struct {
	host string
}

// NewVolumeStatsCollector creates a new volume stats prometheus collector.
func NewVolumeStatsCollector(host string) prometheus.Collector {
	return &volumeStatsCollector{host: host}
}

// Describe implements the prometheus.Collector interface.
func (collector *volumeStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- volumeStatsCapacityBytes
	ch <- volumeStatsAvailableBytes
	ch <- volumeStatsUsedBytes
	ch <- volumeStatsInodes
	ch <- volumeStatsInodesFree
	ch <- volumeStatsInodesUsed
}

// Collect implements the prometheus.Collector interface.
func (collector *volumeSta