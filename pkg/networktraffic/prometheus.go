package networktraffic

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/klog/v2"
)

const (
	// nodeMeasureQueryTemplate is the template string to get the query for the node used bandwidth
	nodeMeasureQueryTemplate        = "sum_over_time(node_network_receive_bytes_total{node=\"%s\",device=\"%s\"}[%s])"
	nodeMemActualUsageQueryTemplate = "100 - (avg(node_memory_MemAvailable_bytes{node=\"%s\"}) / avg(node_memory_MemTotal_bytes{node=\"%s\"}) * 100)"
	nodeLoad1mQueryTemplate         = "node_load1{node=\"%s\"}"
	nodeLatency                     = `probe_duration_seconds{instance="172.25.216.162"}`
)

// Handles the interaction of the networkplugin with Prometheus
type PrometheusHandle struct {
	networkInterface string
	timeRange        time.Duration
	address          string
	api              v1.API
}

func NewPrometheus(address, networkInterface string, timeRange time.Duration) *PrometheusHandle {
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		klog.Fatalf("[NetworkTraffic] Error creating prometheus client: %s", err.Error())
	}

	return &PrometheusHandle{
		networkInterface: networkInterface,
		timeRange:        timeRange,
		address:          address,
		api:              v1.NewAPI(client),
	}
}

func (p *PrometheusHandle) GetNodeBandwidthMeasure(node string) (*model.Sample, error) {
	query := getNodeBandwidthQuery(node, p.networkInterface, p.timeRange)
	klog.Infoln(query)
	res, err := p.query(query)
	if err != nil {
		return nil, fmt.Errorf("[NetworkTraffic] Error querying prometheus: %w", err)
	}

	nodeMeasure := res.(model.Vector)
	if len(nodeMeasure) != 1 {
		return nil, fmt.Errorf("[NetworkTraffic] Invalid response, expected 1 value, got %d", len(nodeMeasure))
	}

	return nodeMeasure[0], nil
}

func (p *PrometheusHandle) GetNodeActualMemoryMeasure(node string) (*model.Sample, error) {
	query := getNodeActualMemoryQuery(node)
	res, err := p.query(query)
	if err != nil {
		return nil, fmt.Errorf("[NetworkTraffic] Error querying prometheus: %w", err)
	}

	nodeMeasure := res.(model.Vector)
	if nodeMeasure == nil {
		return nil, fmt.Errorf("[NetworkTraffic] Invalid response, expected 1 value, got %d", len(nodeMeasure))
	}

	return nodeMeasure[0], nil
}

func (p *PrometheusHandle) GetNode1mLoadMeasure(node string) (*model.Sample, error) {
	query := getNode1mLoadQuery(node)
	res, err := p.query(query)
	if err != nil {
		return nil, fmt.Errorf("[NetworkTraffic] Error querying prometheus: %w", err)
	}

	nodeMeasure := res.(model.Vector)
	if len(nodeMeasure) != 1 {
		return nil, fmt.Errorf("[NetworkTraffic] Invalid response, expected 1 value, got %d", len(nodeMeasure))
	}

	return nodeMeasure[0], nil
}

func getNodeActualMemoryQuery(node string) string {
	return fmt.Sprintf(nodeMemActualUsageQueryTemplate, node, node)
}

func getNodeBandwidthQuery(node, networkInterface string, timeRange time.Duration) string {
	return fmt.Sprintf(nodeMeasureQueryTemplate, node, networkInterface, timeRange)
}

func getNode1mLoadQuery(node string) string {
	return fmt.Sprintf(nodeLoad1mQueryTemplate, node)
}

func (p *PrometheusHandle) query(query string) (model.Value, error) {
	results, warnings, err := p.api.Query(context.Background(), query, time.Now())

	if len(warnings) > 0 {
		klog.Warningf("[NetworkTraffic] Warnings: %v\n", warnings)
	}

	return results, err
}
