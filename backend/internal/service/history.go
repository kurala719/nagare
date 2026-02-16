package service

import (
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
	IPAddr            string    `json:"ip_addr"`
	SampledAt         time.Time `json:"sampled_at"`
}

// NetworkStatusHistoryResp represents network status history for API responses.
type NetworkStatusHistoryResp struct {
	Score         int       `json:"score"`
	MonitorTotal  int       `json:"monitor_total"`
	MonitorActive int       `json:"monitor_active"`
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
		HostID:    item.HID,
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
	_ = repository.AddHostHistoryDAO(model.HostHistory{
		HostID:            host.ID,
		Status:            host.Status,
		StatusDescription: host.StatusDescription,
		IPAddr:            host.IPAddr,
		SampledAt:         sampledAt,
	})
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
