package service

import (
	"fmt"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
)

// ItemHistoryResp represents item history for API responses.
type ItemHistoryResp struct {
	ItemID    uint      `json:"item_id"`
	HostID    uint      `json:"host_id"`
	Value     string    `json:"value"`
	Units     string    `json:"units"`
	Status    int       `json:"status"`
	SampledAt time.Time `json:"sampled_at"`
}

// HostHistoryResp represents host history for API responses.
type HostHistoryResp struct {
	HostID            uint      `json:"host_id"`
	Status            int       `json:"status"`
	StatusDescription string    `json:"status_description"`
	ItemTotal         int       `json:"item_total"`
	ItemActive        int       `json:"item_active"`
	IPAddr            string    `json:"ip_addr"`
	SampledAt         time.Time `json:"sampled_at"`
}

// NetworkStatusHistoryResp represents network status history for API responses.
type NetworkStatusHistoryResp struct {
	Score         int       `json:"score"`
	MonitorTotal  int       `json:"monitor_total"`
	MonitorActive int       `json:"monitor_active"`
	GroupTotal    int       `json:"group_total"`
	GroupActive   int       `json:"group_active"`
	GroupImpacted int       `json:"group_impacted"`
	HostTotal     int       `json:"host_total"`
	HostActive    int       `json:"host_active"`
	ItemTotal     int       `json:"item_total"`
	ItemActive    int       `json:"item_active"`
	SampledAt     time.Time `json:"sampled_at"`
}

func GetItemHistoryServ(itemID uint, from, to *time.Time, limit int) ([]ItemHistoryResp, error) {
	rows, err := repository.ListItemHistoryDAO(itemID, from, to, limit)
	if err != nil {
		return nil, err
	}
	reverseItemHistory(rows)
	resp := make([]ItemHistoryResp, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, ItemHistoryResp{
			ItemID:    row.ItemID,
			HostID:    row.HostID,
			Value:     row.Value,
			Units:     row.Units,
			Status:    row.Status,
			SampledAt: row.SampledAt,
		})
	}
	return resp, nil
}

func GetHostHistoryServ(hostID uint, from, to *time.Time, limit int) ([]HostHistoryResp, error) {
	rows, err := repository.ListHostHistoryDAO(hostID, from, to, limit)
	if err != nil {
		return nil, err
	}
	reverseHostHistory(rows)
	resp := make([]HostHistoryResp, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, HostHistoryResp{
			HostID:            row.HostID,
			Status:            row.Status,
			StatusDescription: row.StatusDescription,
			ItemTotal:         row.ItemTotal,
			ItemActive:        row.ItemActive,
			IPAddr:            row.IPAddr,
			SampledAt:         row.SampledAt,
		})
	}
	return resp, nil
}

func GetNetworkStatusHistoryServ(from, to *time.Time, limit int) ([]NetworkStatusHistoryResp, error) {
	rows, err := repository.ListNetworkStatusHistoryDAO(from, to, limit)
	if err != nil {
		return nil, err
	}
	reverseNetworkStatusHistory(rows)
	resp := make([]NetworkStatusHistoryResp, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, NetworkStatusHistoryResp{
			Score:         row.Score,
			MonitorTotal:  row.MonitorTotal,
			MonitorActive: row.MonitorActive,
			GroupTotal:    row.GroupTotal,
			GroupActive:   row.GroupActive,
			GroupImpacted: row.GroupImpacted,
			HostTotal:     row.HostTotal,
			HostActive:    row.HostActive,
			ItemTotal:     row.ItemTotal,
			ItemActive:    row.ItemActive,
			SampledAt:     row.SampledAt,
		})
	}
	return resp, nil
}

func recordItemHistory(item model.Item, sampledAt time.Time) {
	if sampledAt.IsZero() {
		sampledAt = time.Now().UTC()
	}
	_ = repository.AddItemHistoryDAO(model.ItemHistory{
		ItemID:    item.ID,
		HostID:    item.HostID,
		Value:     item.LastValue,
		Units:     item.Units,
		Status:    item.Status,
		SampledAt: sampledAt,
	})
}

func recordHostHistory(host model.Host, sampledAt time.Time) {
	if sampledAt.IsZero() {
		sampledAt = time.Now().UTC()
	}

	totalCount := 0
	activeCount := 0
	items, err := repository.GetItemsByHIDDAO(host.ID)
	if err != nil {
		fmt.Printf("[ERROR] recordHostHistory: GetItemsByHIDDAO failed for HostID=%d: %v\n", host.ID, err)
	} else {
		totalCount = len(items)
		for _, it := range items {
			if it.Status == 1 { // Assuming 1 is Active
				activeCount++
			}
		}
	}

	// Safety check: Don't record history with 0 items if the host should have items,
	// or if we are in the middle of a sync where items haven't been populated yet.
	if totalCount == 0 && activeCount == 0 {
		fmt.Printf("[DEBUG] recordHostHistory: Skipping snapshot for HostID=%d because item counts are 0\n", host.ID)
		return
	}

	h := model.HostHistory{
		HostID:            host.ID,
		Status:            host.Status,
		StatusDescription: host.StatusDescription,
		ItemTotal:         totalCount,
		ItemActive:        activeCount,
		IPAddr:            host.IPAddr,
		SampledAt:         sampledAt,
	}
	_ = repository.AddHostHistoryDAO(h)
}

func recordNetworkStatusSnapshot(sampledAt time.Time) {
	if sampledAt.IsZero() {
		sampledAt = time.Now().UTC()
	}
	score, err := GetHealthScoreServ()
	if err != nil {
		return
	}
	_ = repository.AddNetworkStatusHistoryDAO(model.NetworkStatusHistory{
		Score:         score.Score,
		MonitorTotal:  score.MonitorTotal,
		MonitorActive: score.MonitorActive,
		GroupTotal:    score.GroupTotal,
		GroupActive:   score.GroupActive,
		GroupImpacted: score.GroupImpacted,
		HostTotal:     score.HostTotal,
		HostActive:    score.HostActive,
		ItemTotal:     score.ItemTotal,
		ItemActive:    score.ItemActive,
		SampledAt:     sampledAt,
	})
}

func reverseItemHistory(rows []model.ItemHistory) {
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
}

func reverseHostHistory(rows []model.HostHistory) {
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
}

func reverseNetworkStatusHistory(rows []model.NetworkStatusHistory) {
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
}

// GenerateTestHistoryServ - DEVELOPMENT ONLY: Generates sample history data for testing charts
func GenerateTestHistoryServ() error {
	// Get first item with a host
	items, err := repository.GetAllItemsDAO()
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil // No items, nothing to do
	}

	item := items[0]
	if item.ID == 0 {
		return nil // No valid item
	}

	// Generate 24 hours of sample data (one hour intervals)
	now := time.Now().UTC()
	for i := 0; i < 24; i++ {
		sampledAt := now.Add(time.Duration(-i) * time.Hour)
		// Generate varying values (e.g., CPU 30-80%, Memory 40-90%, Speed 100-500)
		baseValue := float64(50 + (i % 30) + i)
		if baseValue > 95 {
			baseValue = 95
		}

		if err := repository.AddItemHistoryDAO(model.ItemHistory{
			ItemID:    item.ID,
			HostID:    item.HostID,
			Value:     fmt.Sprintf("%.2f", baseValue),
			Units:     item.Units,
			Status:    1,
			SampledAt: sampledAt,
		}); err != nil {
			return err
		}
	}
	return nil
}
