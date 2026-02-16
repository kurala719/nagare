package service

import (
	"fmt"
	"sync"

	"nagare/internal/model"
	"nagare/internal/repository"
)

type HealthScore struct {
	Score         int `json:"score"`
	MonitorTotal  int `json:"monitor_total"`
	MonitorActive int `json:"monitor_active"`
	HostTotal     int `json:"host_total"`
	HostActive    int `json:"host_active"`
	ItemTotal     int `json:"item_total"`
	ItemActive    int `json:"item_active"`
}

func GetHealthScoreServ() (HealthScore, error) {
	var (
		monitors         []model.Monitor
		hosts            []model.Host
		items            []model.Item
		errM, errH, errI error
		wg               sync.WaitGroup
	)

	wg.Add(3)
	go func() {
		defer wg.Done()
		monitors, errM = repository.GetAllMonitorsDAO()
	}()
	go func() {
		defer wg.Done()
		hosts, errH = repository.GetAllHostsDAO()
	}()
	go func() {
		defer wg.Done()
		items, errI = repository.GetItems()
	}()
	wg.Wait()

	if errM != nil {
		return HealthScore{}, fmt.Errorf("failed to load monitors: %w", errM)
	}
	if errH != nil {
		return HealthScore{}, fmt.Errorf("failed to load hosts: %w", errH)
	}
	if errI != nil {
		return HealthScore{}, fmt.Errorf("failed to load items: %w", errI)
	}

	monitorScore, monitorActive, monitorTotal := statusScore(len(monitors), func(i int) (enabled bool, status int) {
		return monitors[i].Enabled != 0, monitors[i].Status
	})
	hostScore, hostActive, hostTotal := statusScore(len(hosts), func(i int) (enabled bool, status int) {
		return hosts[i].Enabled != 0, hosts[i].Status
	})
	itemScore, itemActive, itemTotal := statusScore(len(items), func(i int) (enabled bool, status int) {
		return items[i].Enabled != 0, items[i].Status
	})

	weighted := int(0.4*float64(monitorScore) + 0.4*float64(hostScore) + 0.2*float64(itemScore))
	return HealthScore{
		Score:         weighted,
		MonitorTotal:  monitorTotal,
		MonitorActive: monitorActive,
		HostTotal:     hostTotal,
		HostActive:    hostActive,
		ItemTotal:     itemTotal,
		ItemActive:    itemActive,
	}, nil
}

func statusScore(total int, statusFn func(index int) (enabled bool, status int)) (score int, active int, enabledTotal int) {
	if total == 0 {
		return 100, 0, 0
	}
	var weighted float64
	for i := 0; i < total; i++ {
		enabled, status := statusFn(i)
		if !enabled {
			continue
		}
		enabledTotal++
		switch status {
		case 1:
			weighted += 1.0
			active++
		case 3:
			weighted += 0.5
		default:
			weighted += 0.0
		}
	}
	if enabledTotal == 0 {
		return 100, 0, 0
	}
	return int((weighted / float64(enabledTotal)) * 100.0), active, enabledTotal
}
