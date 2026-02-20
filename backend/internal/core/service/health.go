package service

import (
	"fmt"
	"sync"

	"nagare/internal/adapter/repository"
	"nagare/internal/core/domain"
)

type HealthScore struct {
	Score         int `json:"score"`
	MonitorTotal  int `json:"monitor_total"`
	MonitorActive int `json:"monitor_active"`
	GroupTotal    int `json:"group_total"`
	GroupActive   int `json:"group_active"`
	GroupImpacted int `json:"group_impacted"`
	HostTotal     int `json:"host_total"`
	HostActive    int `json:"host_active"`
	ItemTotal     int `json:"item_total"`
	ItemActive    int `json:"item_active"`
}

func GetHealthScoreServ() (HealthScore, error) {
	var (
		monitors               []domain.Monitor
		groups                 []domain.Group
		hosts                  []domain.Host
		items                  []domain.Item
		errM, errG, errH, errI error
		wg                     sync.WaitGroup
	)

	wg.Add(4)
	go func() {
		defer wg.Done()
		monitors, errM = repository.GetAllMonitorsDAO()
	}()
	go func() {
		defer wg.Done()
		groups, errG = repository.GetAllGroupsDAO()
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
	if errG != nil {
		return HealthScore{}, fmt.Errorf("failed to load groups: %w", errG)
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
	groupScore, groupActive, groupTotal := statusScore(len(groups), func(i int) (enabled bool, status int) {
		return groups[i].Enabled != 0, groups[i].Status
	})
	hostScore, hostActive, hostTotal := statusScore(len(hosts), func(i int) (enabled bool, status int) {
		return hosts[i].Enabled != 0, hosts[i].Status
	})
	itemScore, itemActive, itemTotal := statusScore(len(items), func(i int) (enabled bool, status int) {
		return items[i].Enabled != 0, items[i].Status
	})
	groupImpacted := countImpactedGroups(groups, hosts)

	weighted := int(0.3*float64(monitorScore) + 0.2*float64(groupScore) + 0.3*float64(hostScore) + 0.2*float64(itemScore))
	return HealthScore{
		Score:         weighted,
		MonitorTotal:  monitorTotal,
		MonitorActive: monitorActive,
		GroupTotal:    groupTotal,
		GroupActive:   groupActive,
		GroupImpacted: groupImpacted,
		HostTotal:     hostTotal,
		HostActive:    hostActive,
		ItemTotal:     itemTotal,
		ItemActive:    itemActive,
	}, nil
}

func countImpactedGroups(groups []domain.Group, hosts []domain.Host) int {
	if len(groups) == 0 || len(hosts) == 0 {
		return 0
	}
	enabledGroups := make(map[uint]struct{}, len(groups))
	for _, group := range groups {
		if group.Enabled != 0 {
			enabledGroups[group.ID] = struct{}{}
		}
	}
	impacted := map[uint]struct{}{}
	for _, host := range hosts {
		if host.Status != 2 || host.GroupID == 0 {
			continue
		}
		if _, ok := enabledGroups[host.GroupID]; ok {
			impacted[host.GroupID] = struct{}{}
		}
	}
	return len(impacted)
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
