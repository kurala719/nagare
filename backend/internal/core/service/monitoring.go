package service

import (
	"strings"
	"sync"
	"time"

	"nagare/internal/adapter/repository"
	"nagare/internal/core/domain"
)

type MetricSnapshot struct {
	HostID    uint      `json:"host_id"`
	HostName  string    `json:"host_name"`
	ItemID    uint      `json:"item_id"`
	ItemName  string    `json:"item_name"`
	Value     string    `json:"value"`
	Units     string    `json:"units"`
	Status    int       `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

func GetNetworkMetricsServ(query string, limit int) ([]MetricSnapshot, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	var items []domain.Item
	var hosts []domain.Host
	var errI, errH error
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		items, errI = loadMetricsItems(query)
	}()
	go func() {
		defer wg.Done()
		hosts, errH = repository.GetAllHostsDAO()
	}()
	wg.Wait()

	if errI != nil {
		return nil, errI
	}
	if errH != nil {
		return nil, errH
	}
	hostNames := make(map[uint]string, len(hosts))
	for _, host := range hosts {
		hostNames[host.ID] = host.Name
	}

	results := make([]MetricSnapshot, 0, len(items))
	for _, item := range items {
		results = append(results, MetricSnapshot{
			HostID:    item.HID,
			HostName:  hostNames[item.HID],
			ItemID:    item.ID,
			ItemName:  item.Name,
			Value:     item.LastValue,
			Units:     item.Units,
			Status:    item.Status,
			UpdatedAt: item.UpdatedAt,
		})
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

func loadMetricsItems(query string) ([]domain.Item, error) {
	trimmed := strings.TrimSpace(query)
	if trimmed != "" {
		filter := domain.ItemFilter{Query: trimmed, Limit: 1000}
		return repository.SearchItemsDAO(filter)
	}
	items, err := repository.GetAllItemsDAO()
	if err != nil {
		return nil, err
	}
	filtered := make([]domain.Item, 0, len(items))
	for _, item := range items {
		if isNetworkMetric(item.Name) {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) > 0 {
		return filtered, nil
	}
	// Fallback: include any active items with values so network summaries are not empty.
	for _, item := range items {
		if item.Status == 1 && item.LastValue != "" {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func isNetworkMetric(name string) bool {
	lower := strings.ToLower(name)
	keywords := []string{
		"cpu", "processor", "load", "util", "utilization",
		"mem", "memory", "ram",
		"disk", "io",
		"network", "net", "traffic", "bandwidth", "throughput",
		"interface", "ifin", "ifout", "inbound", "outbound", "rx", "tx",
		"inbps", "outbps", "octet", "packet", "loss", "icmp", "ping",
		"latency", "rtt", "delay",
		"availability", "uptime", "temperature",
	}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}
