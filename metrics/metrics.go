package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sync"
)

var instance *Metrics
var once sync.Once

// Metrics implements instrumentation of metrics for kube-iptables-tailer using Prometheus
// registry is used by Prometheus to collect metrics
// packetDropsCount is the Counters Collector in Prometheus having variable labels related to an iptables packet drop
type Metrics struct {
	registry         *prometheus.Registry
	packetDropsCount *prometheus.CounterVec
}

// Return the singleton instance of metrics
func GetInstance() *Metrics {
	once.Do(initMetricsSingleton) // thread-safe way to construct the singleton instance
	return instance
}

// Helper function to init singleton object of Metrics
func initMetricsSingleton() {
	packetDropCountsVec := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "packet_drops_count",
		Help: "Counter for number of packet drops handled; excludes expired and duplicates.",
	},
		[]string{
			"src_namespace",
			"src_pod",
			"dst_namespace",
			"dst_pod",
			"dst_port",
		},
	)

	// registry the count vector in prometheus
	r := prometheus.NewRegistry()
	r.MustRegister(packetDropCountsVec)

	instance = &Metrics{packetDropsCount: packetDropCountsVec, registry: r}
}

// Return the handler of metrics
func (m *Metrics) GetHandler() http.Handler {
	// need to specify registry to avoid getting extra data sent in prometheus
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// Update the metrics by given service name
func (m *Metrics) ProcessPacketDrop(srcNamespace, srcPod, dstNamespace, dstPod string, dstPort string) {
	m.packetDropsCount.With(prometheus.Labels{
		"src_namespace": srcNamespace,
		"src_pod":       srcPod,
		"dst_namespace": dstNamespace,
		"dst_pod":       dstPod,
		"dst_port":      dstPort,
	}).Inc()
}
