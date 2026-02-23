package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/monitors"
	"nagare/internal/service/utils"

	"github.com/gosnmp/gosnmp"
)

type SNMPOIDProbeResult struct {
	HostID       uint   `json:"host_id"`
	HostName     string `json:"host_name"`
	Target       string `json:"target"`
	OID          string `json:"oid"`
	Version      string `json:"version"`
	Port         int    `json:"port"`
	Success      bool   `json:"success"`
	Value        string `json:"value,omitempty"`
	ValueType    string `json:"value_type,omitempty"`
	RawType      string `json:"raw_type,omitempty"`
	DurationMs   int64  `json:"duration_ms"`
	Error        string `json:"error,omitempty"`
	CommunityLen int    `json:"community_len,omitempty"`
	Reachable    bool   `json:"reachable"`
	ReachableBy  string `json:"reachable_by,omitempty"`
}

func normalizeProbeOID(oid string) string {
	oid = strings.TrimSpace(oid)
	if oid == "" {
		return oid
	}
	if strings.HasPrefix(oid, ".") {
		return oid
	}
	return "." + oid
}

func formatProbePDU(v gosnmp.SnmpPDU) (string, string) {
	typeName := v.Type.String()
	switch v.Type {
	case gosnmp.OctetString:
		if bytes, ok := v.Value.([]byte); ok {
			isBinary := false
			for _, b := range bytes {
				if (b < 32 && b != 9 && b != 10 && b != 13) || b > 126 {
					isBinary = true
					break
				}
			}
			if isBinary {
				return fmt.Sprintf("%x", bytes), typeName
			}
			return string(bytes), typeName
		}
		return "", typeName
	case gosnmp.Integer, gosnmp.Counter32, gosnmp.Gauge32, gosnmp.Counter64, gosnmp.TimeTicks:
		return fmt.Sprintf("%d", gosnmp.ToBigInt(v.Value)), typeName
	case gosnmp.ObjectIdentifier, gosnmp.IPAddress:
		return fmt.Sprintf("%v", v.Value), typeName
	case gosnmp.NoSuchObject:
		return "NoSuchObject", typeName
	case gosnmp.NoSuchInstance:
		return "NoSuchInstance", typeName
	case gosnmp.EndOfMibView:
		return "EndOfMibView", typeName
	case gosnmp.Null:
		return "N/A", typeName
	default:
		return fmt.Sprintf("%v", v.Value), typeName
	}
}

func probeSingleOID(gs *gosnmp.GoSNMP, oid string) (*gosnmp.SnmpPDU, error) {
	packet, err := gs.Get([]string{normalizeProbeOID(oid)})
	if err != nil {
		return nil, err
	}
	if packet == nil || len(packet.Variables) == 0 {
		return nil, fmt.Errorf("empty SNMP response")
	}
	return &packet.Variables[0], nil
}

// ProbeSNMPOIDServ probes one SNMP OID on a host and returns raw debug output.
func ProbeSNMPOIDServ(hid uint, oid string) (SNMPOIDProbeResult, error) {
	host, err := repository.GetHostByIDDAO(hid)
	if err != nil {
		return SNMPOIDProbeResult{}, err
	}

	markHostFromProbe := func(success bool, reason string) {
		if success {
			_ = repository.UpdateHostStatusAndDescriptionDAO(host.ID, 1, "")
		} else {
			if strings.TrimSpace(reason) == "" {
				reason = "SNMP probe failed"
			}
			_ = repository.UpdateHostStatusAndDescriptionDAO(host.ID, 2, reason)
		}
		if refreshed, rErr := repository.GetHostByIDDAO(host.ID); rErr == nil {
			recordHostHistory(refreshed, time.Now().UTC())
		}
	}

	normOID := normalizeProbeOID(oid)
	if normOID == "" {
		return SNMPOIDProbeResult{}, fmt.Errorf("oid is required")
	}

	target := strings.TrimSpace(host.IPAddr)
	if target == "" {
		target = strings.TrimSpace(host.Hostid)
	}
	if target == "" {
		return SNMPOIDProbeResult{}, fmt.Errorf("host has no SNMP target (ip_addr/hostid empty)")
	}

	version := strings.TrimSpace(host.SNMPVersion)
	if version == "" {
		version = "v2c"
	}
	port := host.SNMPPort
	if port == 0 {
		port = 161
	}
	community := strings.TrimSpace(host.SNMPCommunity)
	if community == "" {
		community = "public"
	}

	res := SNMPOIDProbeResult{
		HostID:       host.ID,
		HostName:     host.Name,
		Target:       target,
		OID:          normOID,
		Version:      version,
		Port:         port,
		CommunityLen: len(community),
		Reachable:    false,
	}

	gs := &gosnmp.GoSNMP{
		Target:  target,
		Port:    uint16(port),
		Timeout: 5 * time.Second,
		Retries: 1,
		MaxOids: 1,
	}

	switch version {
	case "v3":
		gs.Version = gosnmp.Version3
		gs.SecurityModel = gosnmp.UserSecurityModel
		level := host.SNMPV3SecurityLevel
		if level == "" {
			level = "NoAuthNoPriv"
		}
		switch level {
		case "AuthNoPriv":
			gs.MsgFlags = gosnmp.AuthNoPriv
		case "AuthPriv":
			gs.MsgFlags = gosnmp.AuthPriv
		default:
			gs.MsgFlags = gosnmp.NoAuthNoPriv
		}
		authPass := host.SNMPV3AuthPass
		if authPass != "" {
			if decrypted, decErr := utils.Decrypt(authPass); decErr == nil {
				authPass = decrypted
			}
		}
		privPass := host.SNMPV3PrivPass
		if privPass != "" {
			if decrypted, decErr := utils.Decrypt(privPass); decErr == nil {
				privPass = decrypted
			}
		}
		sp := &gosnmp.UsmSecurityParameters{UserName: host.SNMPV3User}
		if level != "NoAuthNoPriv" {
			sp.AuthenticationPassphrase = authPass
			switch host.SNMPV3AuthProtocol {
			case "MD5":
				sp.AuthenticationProtocol = gosnmp.MD5
			default:
				sp.AuthenticationProtocol = gosnmp.SHA
			}
		}
		if level == "AuthPriv" {
			sp.PrivacyPassphrase = privPass
			switch host.SNMPV3PrivProtocol {
			case "DES":
				sp.PrivacyProtocol = gosnmp.DES
			default:
				sp.PrivacyProtocol = gosnmp.AES
			}
		}
		gs.SecurityParameters = sp
	case "v1":
		gs.Version = gosnmp.Version1
		gs.Community = community
	default:
		gs.Version = gosnmp.Version2c
		gs.Community = community
	}

	start := time.Now()
	if err := gs.Connect(); err != nil {
		res.DurationMs = time.Since(start).Milliseconds()
		res.Error = fmt.Sprintf("connect failed: %v", err)
		res.Success = false
		markHostFromProbe(false, res.Error)
		return res, nil
	}
	defer gs.Conn.Close()

	pdu, err := probeSingleOID(gs, normOID)
	res.DurationMs = time.Since(start).Milliseconds()
	if err != nil {
		requestedErr := err
		origTimeout := gs.Timeout
		origRetries := gs.Retries
		gs.Timeout = 2 * time.Second
		gs.Retries = 0
		fallbackOids := []string{".1.3.6.1.2.1.1.3.0", ".1.3.6.1.2.1.1.5.0"}
		for _, fallbackOID := range fallbackOids {
			fallbackPDU, fallbackErr := probeSingleOID(gs, fallbackOID)
			if fallbackErr == nil && fallbackPDU != nil {
				fallbackVal, fallbackType := formatProbePDU(*fallbackPDU)
				res.Success = true
				res.Reachable = true
				res.ReachableBy = normalizeProbeOID(fallbackOID)
				res.Value = fmt.Sprintf("reachable via %s = %s", normalizeProbeOID(fallbackOID), fallbackVal)
				res.ValueType = fallbackType
				res.RawType = fallbackPDU.Type.String()
				res.Error = fmt.Sprintf("requested OID %s failed: %v", normOID, requestedErr)
				markHostFromProbe(true, "")
				gs.Timeout = origTimeout
				gs.Retries = origRetries
				return res, nil
			}
		}
		gs.Timeout = origTimeout
		gs.Retries = origRetries

		res.Error = requestedErr.Error()
		res.Success = false
		res.Reachable = false
		markHostFromProbe(false, res.Error)
		return res, nil
	}

	value, valueType := formatProbePDU(*pdu)
	res.Value = value
	res.ValueType = valueType
	res.RawType = pdu.Type.String()
	res.Success = true
	res.Reachable = true
	res.ReachableBy = normOID
	markHostFromProbe(true, "")
	return res, nil
}

// HostReq represents a host request
type HostReq struct {
	Name        string `json:"name" binding:"required"`
	MID         uint   `json:"m_id"`
	GroupID     uint   `json:"group_id"`
	Hostid      string `json:"hostid"`
	Description string `json:"description"`
	Enabled     int    `json:"enabled"`
	IPAddr      string `json:"ip_addr"`
	Comment     string `json:"comment"`
	SSHUser     string `json:"ssh_user"`
	SSHPassword string `json:"ssh_password"`
	SSHPort     int    `json:"ssh_port"`
	// SNMP Configuration
	SNMPCommunity       string     `json:"snmp_community"`
	SNMPVersion         string     `json:"snmp_version"`
	SNMPPort            int        `json:"snmp_port"`
	SNMPV3User          string     `json:"snmp_v3_user"`
	SNMPV3AuthPass      string     `json:"snmp_v3_auth_pass"`
	SNMPV3PrivPass      string     `json:"snmp_v3_priv_pass"`
	SNMPV3AuthProtocol  string     `json:"snmp_v3_auth_protocol"`
	SNMPV3PrivProtocol  string     `json:"snmp_v3_priv_protocol"`
	SNMPV3SecurityLevel string     `json:"snmp_v3_security_level"`
	LastSyncAt          *time.Time `json:"last_sync_at,omitempty"`
	ExternalSource      string     `json:"external_source,omitempty"`
}

// HostResp represents a host response
type HostResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	MID         uint   `json:"m_id"`
	GroupID     uint   `json:"group_id"`
	GroupName   string `json:"group_name"`
	MonitorName string `json:"monitor_name"`
	Hostid      string `json:"hostid"`
	Description string `json:"description"`
	Enabled     int    `json:"enabled"`
	Status      int    `json:"status"`
	StatusDesc  string `json:"status_description"`
	IPAddr      string `json:"ip_addr"`
	Comment     string `json:"comment"`
	SSHUser     string `json:"ssh_user"`
	SSHPort     int    `json:"ssh_port"`
	// SNMP Configuration
	SNMPCommunity       string     `json:"snmp_community"`
	SNMPVersion         string     `json:"snmp_version"`
	SNMPPort            int        `json:"snmp_port"`
	SNMPV3User          string     `json:"snmp_v3_user"`
	SNMPV3AuthProtocol  string     `json:"snmp_v3_auth_protocol"`
	SNMPV3PrivProtocol  string     `json:"snmp_v3_priv_protocol"`
	SNMPV3SecurityLevel string     `json:"snmp_v3_security_level"`
	LastSyncAt          *time.Time `json:"last_sync_at"`
	ExternalSource      string     `json:"external_source"`
	HealthScore         int        `json:"health_score"`
}

// GetAllHostsServ retrieves all hosts
func GetAllHostsServ() ([]HostResp, error) {
	hosts, err := repository.GetAllHostsDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get hosts: %w", err)
	}

	result := make([]HostResp, 0, len(hosts))
	for _, h := range hosts {
		result = append(result, hostToResp(h))
	}
	return result, nil
}

// SearchHostsServ retrieves hosts by filter
func SearchHostsServ(filter model.HostFilter) ([]HostResp, error) {
	hosts, err := repository.SearchHostsDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search hosts: %w", err)
	}
	result := make([]HostResp, 0, len(hosts))
	for _, h := range hosts {
		result = append(result, hostToResp(h))
	}
	return result, nil
}

// CountHostsServ returns total count for hosts by filter
func CountHostsServ(filter model.HostFilter) (int64, error) {
	return repository.CountHostsDAO(filter)
}

// GetHostByIDServ retrieves a host by ID
func GetHostByIDServ(id uint) (HostResp, error) {
	host, err := repository.GetHostByIDDAO(id)
	if err != nil {
		return HostResp{}, fmt.Errorf("failed to get host: %w", err)
	}
	return hostToResp(host), nil
}

// DeleteHostsByMIDServ deletes all hosts by monitor ID
func DeleteHostsByMIDServ(mid uint) error {
	return repository.DeleteHostByMIDDAO(mid)
}

// AddHostServ creates a new host
func AddHostServ(h HostReq) (HostResp, error) {
	newHost := model.Host{
		Name:                h.Name,
		Hostid:              h.Hostid,
		MonitorID:           h.MID,
		GroupID:             h.GroupID,
		Description:         h.Description,
		Enabled:             h.Enabled,
		IPAddr:              h.IPAddr,
		Comment:             h.Comment,
		SSHUser:             h.SSHUser,
		SSHPort:             h.SSHPort,
		SNMPCommunity:       h.SNMPCommunity,
		SNMPVersion:         h.SNMPVersion,
		SNMPPort:            h.SNMPPort,
		SNMPV3User:          h.SNMPV3User,
		SNMPV3AuthPass:      h.SNMPV3AuthPass,
		SNMPV3PrivPass:      h.SNMPV3PrivPass,
		SNMPV3AuthProtocol:  h.SNMPV3AuthProtocol,
		SNMPV3PrivProtocol:  h.SNMPV3PrivProtocol,
		SNMPV3SecurityLevel: h.SNMPV3SecurityLevel,
	}
	if h.SNMPPort == 0 {
		newHost.SNMPPort = 161
	}
	if h.SNMPVersion == "" {
		newHost.SNMPVersion = "v2c"
	}
	if h.SNMPCommunity == "" {
		newHost.SNMPCommunity = "public"
	}
	if h.SSHPort == 0 {
		newHost.SSHPort = 22
	}
	if h.SSHPassword != "" {
		encrypted, err := utils.Encrypt(h.SSHPassword)
		if err == nil {
			newHost.SSHPassword = encrypted
		}
	}
	if h.SNMPV3AuthPass != "" {
		encrypted, err := utils.Encrypt(h.SNMPV3AuthPass)
		if err == nil {
			newHost.SNMPV3AuthPass = encrypted
		}
	}
	if h.SNMPV3PrivPass != "" {
		encrypted, err := utils.Encrypt(h.SNMPV3PrivPass)
		if err == nil {
			newHost.SNMPV3PrivPass = encrypted
		}
	}
	if h.MID == 0 {
		internalMonitors, sErr := repository.SearchMonitorsDAO(model.MonitorFilter{Query: "Nagare Internal"})
		if sErr == nil && len(internalMonitors) > 0 {
			newHost.MonitorID = internalMonitors[0].ID
		}
	}

	if newHost.MonitorID > 0 {
		monitorStatus := 1
		if monitor, err := repository.GetMonitorByIDDAO(newHost.MonitorID); err == nil {
			monitorStatus = determineMonitorStatus(monitor)
		}
		
		groupStatus := 1
		if newHost.GroupID > 0 {
			if g, err := repository.GetGroupByIDDAO(newHost.GroupID); err == nil {
				groupStatus = g.Status
			}
		}
		newHost.Status = determineHostStatus(newHost, monitorStatus, groupStatus)
	} else {
		groupStatus := 1
		if newHost.GroupID > 0 {
			if g, err := repository.GetGroupByIDDAO(newHost.GroupID); err == nil {
				groupStatus = g.Status
			}
		}
		newHost.Status = determineHostStatus(newHost, 1, groupStatus)
	}

	if err := repository.AddHostDAO(&newHost); err != nil {
		return HostResp{}, fmt.Errorf("failed to add host: %w", err)
	}

	// Auto-push to monitor if MID is set (either from request or resolved to "Nagare Internal")
	if newHost.MonitorID > 0 {
		LogService("info", "auto-pushing new host to monitor", map[string]interface{}{"host_id": newHost.ID, "monitor_id": newHost.MonitorID}, nil, "")
		if _, pushErr := PushHostToMonitorServ(newHost.MonitorID, newHost.ID); pushErr != nil {
			LogService("error", "auto-push failed for new host", map[string]interface{}{"host_id": newHost.ID, "error": pushErr.Error()}, nil, "")
		} else {
			// Reload to get the external Hostid if it was updated during push
			if refreshed, err := repository.GetHostByIDDAO(newHost.ID); err == nil {
				newHost = refreshed
			}
		}
	}

	_, _ = recomputeHostStatus(newHost.ID)
	
	// Final reload to ensure we have all fields including any from recompute
	if refreshed, err := repository.GetHostByIDDAO(newHost.ID); err == nil {
		newHost = refreshed
	}

	return HostResp{
		ID:          int(newHost.ID),
		Name:        newHost.Name,
		MID:         newHost.MonitorID,
		GroupID:     newHost.GroupID,
		Hostid:      newHost.Hostid,
		Description: newHost.Description,
		Enabled:     newHost.Enabled,
		Status:      newHost.Status,
		StatusDesc:  newHost.StatusDescription,
		IPAddr:      newHost.IPAddr,
		Comment:     newHost.Comment,
	}, nil
}

// UpdateHostServ updates a host
func UpdateHostServ(id uint, h HostReq) error {
	existing, err := repository.GetHostByIDDAO(id)
	if err != nil {
		return err
	}
	monitorID := h.MID
	if monitorID == 0 {
		monitorID = existing.MonitorID
	}
	updated := model.Host{
		Name:                h.Name,
		Hostid:              h.Hostid,
		MonitorID:           monitorID,
		GroupID:             h.GroupID,
		Description:         h.Description,
		Enabled:             h.Enabled,
		IPAddr:              h.IPAddr,
		Comment:             h.Comment,
		SSHUser:             h.SSHUser,
		SSHPort:             h.SSHPort,
		SNMPCommunity:       h.SNMPCommunity,
		SNMPVersion:         h.SNMPVersion,
		SNMPPort:            h.SNMPPort,
		SNMPV3User:          h.SNMPV3User,
		SNMPV3AuthPass:      h.SNMPV3AuthPass,
		SNMPV3PrivPass:      h.SNMPV3PrivPass,
		SNMPV3AuthProtocol:  h.SNMPV3AuthProtocol,
		SNMPV3PrivProtocol:  h.SNMPV3PrivProtocol,
		SNMPV3SecurityLevel: h.SNMPV3SecurityLevel,
		ActiveAvailable:     existing.ActiveAvailable,
		LastSyncAt:          existing.LastSyncAt,
		ExternalSource:      existing.ExternalSource,
		Status:              existing.Status,
		StatusDescription:   existing.StatusDescription,
	}
	if h.LastSyncAt != nil {
		updated.LastSyncAt = h.LastSyncAt
	}
	if h.ExternalSource != "" {
		updated.ExternalSource = h.ExternalSource
	}
	if h.SSHPort == 0 {
		updated.SSHPort = 22
	}
	if h.SSHPassword != "" {
		encrypted, err := utils.Encrypt(h.SSHPassword)
		if err == nil {
			updated.SSHPassword = encrypted
		}
	} else {
		updated.SSHPassword = existing.SSHPassword
	}
	if h.SNMPV3AuthPass != "" {
		encrypted, err := utils.Encrypt(h.SNMPV3AuthPass)
		if err == nil {
			updated.SNMPV3AuthPass = encrypted
		}
	} else {
		updated.SNMPV3AuthPass = existing.SNMPV3AuthPass
	}
	if h.SNMPV3PrivPass != "" {
		encrypted, err := utils.Encrypt(h.SNMPV3PrivPass)
		if err == nil {
			updated.SNMPV3PrivPass = encrypted
		}
	} else {
		updated.SNMPV3PrivPass = existing.SNMPV3PrivPass
	}
	// Preserve status and description unless enabled state changed
	if h.Enabled != existing.Enabled {
		monitorStatus := 1
		if monitorID > 0 {
			if monitor, err := repository.GetMonitorByIDDAO(monitorID); err == nil {
				monitorStatus = determineMonitorStatus(monitor)
			}
		}

		groupStatus := 1
		if updated.GroupID > 0 {
			if g, err := repository.GetGroupByIDDAO(updated.GroupID); err == nil {
				groupStatus = g.Status
			}
		}

		updated.Status = determineHostStatus(updated, monitorStatus, groupStatus)
		updated.StatusDescription = ""
	}
	if err := repository.UpdateHostDAO(id, updated); err != nil {
		return err
	}

	// Auto-push to monitor if MID is set and critical fields changed
	if updated.MonitorID > 0 && (updated.Name != existing.Name || updated.IPAddr != existing.IPAddr || updated.MonitorID != existing.MonitorID) {
		_, _ = PushHostToMonitorServ(updated.MonitorID, id)
	}

	if refreshed, err := repository.GetHostByIDDAO(id); err == nil {
		recordHostHistory(refreshed, time.Now().UTC())
	}
	_ = recomputeItemsForHost(id)
	_, _ = recomputeHostStatus(id)
	return nil
}

// GetHostConnectionDetails retrieves host connection details including decrypted password
func GetHostConnectionDetails(id uint) (string, int, string, string, error) {
	host, err := repository.GetHostByIDDAO(id)
	if err != nil {
		return "", 0, "", "", fmt.Errorf("failed to get host: %w", err)
	}

	if host.SSHUser == "" {
		return "", 0, "", "", fmt.Errorf("SSH user not configured")
	}

	password := ""
	if host.SSHPassword != "" {
		decrypted, err := utils.Decrypt(host.SSHPassword)
		if err != nil {
			return "", 0, "", "", fmt.Errorf("failed to decrypt SSH password: %w", err)
		}
		password = decrypted
	}

	return host.IPAddr, host.SSHPort, host.SSHUser, password, nil
}

// DeleteHostByIDServ deletes a host by ID
func DeleteHostByIDServ(id uint) error {
	return repository.DeleteHostByIDDAO(id)
}

// GetHostsFromMonitorServ retrieves hosts from an external monitor
func GetHostsFromMonitorServ(mid uint) ([]HostResp, error) {
	monitor, err := GetMonitorByIDServ(mid)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitor: %w", err)
	}

	client, err := createMonitorClient(monitor)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitor client: %w", err)
	}

	// Use existing auth token if available
	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else {
		if err := client.Authenticate(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	monitorHosts, err := client.GetHosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get hosts from monitor: %w", err)
	}

	hosts := make([]HostResp, 0, len(monitorHosts))
	for _, h := range monitorHosts {
		activeAvailable := ""
		statusDesc := ""
		if h.Metadata != nil {
			if value, ok := h.Metadata["active_available"]; ok {
				activeAvailable = value
			}
			if value, ok := h.Metadata["status_description"]; ok {
				statusDesc = value
			}
		}
		status := mapMonitorHostStatus(h.Status, activeAvailable)
		hosts = append(hosts, HostResp{
			Name:        h.Name,
			Hostid:      h.ID,
			Description: h.Description,
			IPAddr:      h.IPAddress,
			Enabled:     1,
			Status:      status,
			StatusDesc:  statusDesc,
		})
	}
	return hosts, nil
}

func mapMonitorHostStatus(status string, activeAvailable string) int {
	// Check Zabbix active_available first: 0=unknown, 1=available, 2=not_available
	if activeAvailable == "2" {
		return 2
	}
	if activeAvailable == "1" {
		return 1
	}
	if activeAvailable == "0" {
		return 1 // Default to active if unknown but enabled
	}

	if status == "down" {
		return 2
	}
	if status == "up" {
		return 1
	}
	return 0
}

// createMonitorClient creates a monitor client from a MonitorResp
func createMonitorClient(monitor MonitorResp) (*monitors.Client, error) {
	monitorType := monitors.ParseMonitorType(monitor.Type)
	monitorURL := strings.TrimSpace(monitor.URL)
	urlLower := strings.ToLower(monitorURL)

	if monitorType == monitors.MonitorSNMP && monitor.ID != 1 {
		LogService("warn", "monitor type is SNMP but not internal; forcing zabbix provider", map[string]interface{}{
			"monitor_id":   monitor.ID,
			"monitor_name": monitor.Name,
			"monitor_type": monitor.Type,
		}, nil, "")
		monitorType = monitors.MonitorZabbix
	}

	if monitorType == monitors.MonitorOther {
		if monitorURL != "" && (strings.Contains(urlLower, "zabbix") || monitor.Username != "" || monitor.Password != "" || monitor.AuthToken != "") {
			LogService("warn", "monitor type is other but looks like zabbix; forcing zabbix provider", map[string]interface{}{
				"monitor_id":   monitor.ID,
				"monitor_name": monitor.Name,
				"monitor_type": monitor.Type,
				"monitor_url":  monitor.URL,
			}, nil, "")
			monitorType = monitors.MonitorZabbix
		}
	}

	if monitorType == monitors.MonitorZabbix {
		if strings.Contains(urlLower, "api_jsonrpc.php") {
			monitorURL = strings.TrimSuffix(monitorURL, "/")
			monitorURL = strings.TrimSuffix(monitorURL, "/api_jsonrpc.php")
			monitorURL = strings.TrimSuffix(monitorURL, "/API_JSONRPC.PHP")
		}
		if !strings.Contains(urlLower, "api_jsonrpc.php") && !strings.Contains(urlLower, "/zabbix") {
			LogService("warn", "zabbix monitor URL missing api_jsonrpc.php path", map[string]interface{}{
				"monitor_id":   monitor.ID,
				"monitor_name": monitor.Name,
				"monitor_url":  monitor.URL,
			}, nil, "")
		}
	}

	cfg := monitors.Config{
		Name: monitor.Name,
		Type: monitorType,
		Auth: monitors.AuthConfig{
			URL:      monitorURL,
			Username: monitor.Username,
			Password: monitor.Password,
			Token:    monitor.AuthToken,
		},
		Timeout: 30,
	}
	return monitors.NewClient(cfg)
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	Added   int `json:"added"`
	Updated int `json:"updated"`
	Failed  int `json:"failed"`
	Total   int `json:"total"`
}

// hostToResp converts a domain Host to HostResp
func hostToResp(h model.Host) HostResp {
	return HostResp{
		ID:                  int(h.ID),
		Name:                h.Name,
		MID:                 h.MonitorID,
		GroupID:             h.GroupID,
		GroupName:           h.GroupName,
		MonitorName:         h.MonitorName,
		Hostid:              h.Hostid,
		Description:         h.Description,
		Enabled:             h.Enabled,
		Status:              h.Status,
		StatusDesc:          h.StatusDescription,
		IPAddr:              h.IPAddr,
		Comment:             h.Comment,
		SSHUser:             h.SSHUser,
		SSHPort:             h.SSHPort,
		SNMPCommunity:       h.SNMPCommunity,
		SNMPVersion:         h.SNMPVersion,
		SNMPPort:            h.SNMPPort,
		SNMPV3User:          h.SNMPV3User,
		SNMPV3AuthProtocol:  h.SNMPV3AuthProtocol,
		SNMPV3PrivProtocol:  h.SNMPV3PrivProtocol,
		SNMPV3SecurityLevel: h.SNMPV3SecurityLevel,
		LastSyncAt:          h.LastSyncAt,
		ExternalSource:      h.ExternalSource,
		HealthScore:         h.HealthScore,
	}
}

// findGroupByExternalIDOrName tries to find a group first by external ID, then by name
func findGroupByExternalIDOrName(externalID, groupName string, mid uint) (uint, error) {
	// First try by external ID
	if externalID != "" {
		if group, err := repository.GetGroupByExternalIDDAO(externalID, mid); err == nil {
			return group.ID, nil
		}
	}

	// Then try by group name if provided
	if groupName != "" {
		groups, err := repository.SearchGroupsDAO(model.GroupFilter{Query: groupName})
		if err == nil {
			for _, g := range groups {
				if g.Name == groupName && (g.MonitorID == mid || g.MonitorID == 0) {
					return g.ID, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("group not found")
}

func parseMetadataCSVValues(metadata map[string]string, listKey, singleKey string) []string {
	if metadata == nil {
		return nil
	}

	values := make([]string, 0)
	seen := make(map[string]struct{})
	appendValue := func(value string) {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return
		}
		if _, ok := seen[trimmed]; ok {
			return
		}
		seen[trimmed] = struct{}{}
		values = append(values, trimmed)
	}

	if list, ok := metadata[listKey]; ok && list != "" {
		for _, part := range strings.Split(list, ",") {
			appendValue(part)
		}
	}

	if single, ok := metadata[singleKey]; ok && single != "" {
		appendValue(single)
	}

	return values
}

func resolveHostGroupIDFromMetadata(mid uint, metadata map[string]string, fallbackGroupID uint) uint {
	groupID := fallbackGroupID
	if metadata == nil {
		return groupID
	}

	externalGroupIDs := parseMetadataCSVValues(metadata, "groupids", "groupid")
	externalGroupNames := parseMetadataCSVValues(metadata, "groupnames", "groupname")

	for _, externalGroupID := range externalGroupIDs {
		if group, err := repository.GetGroupByExternalIDDAO(externalGroupID, mid); err == nil {
			return group.ID
		}
	}

	for idx, groupName := range externalGroupNames {
		externalGroupID := ""
		if idx < len(externalGroupIDs) {
			externalGroupID = externalGroupIDs[idx]
		}

		if matchedGroupID, err := findGroupByExternalIDOrName(externalGroupID, groupName, mid); err == nil {
			if externalGroupID != "" {
				if matchedGroup, err := repository.GetGroupByIDDAO(matchedGroupID); err == nil {
					if matchedGroup.ExternalID != externalGroupID || matchedGroup.MonitorID != mid {
						matchedGroup.ExternalID = externalGroupID
						matchedGroup.MonitorID = mid
						_ = repository.UpdateGroupDAO(matchedGroup.ID, matchedGroup)
					}
				}
			}
			return matchedGroupID
		}
	}

	if len(externalGroupIDs) == 0 {
		for _, groupName := range externalGroupNames {
			if matchedGroupID, err := findGroupByExternalIDOrName("", groupName, mid); err == nil && matchedGroupID != 0 {
				return matchedGroupID
			}
		}

		if len(externalGroupNames) > 0 {
			newGroupName := strings.TrimSpace(externalGroupNames[0])
			if newGroupName == "" {
				return groupID
			}
			newGroup := model.Group{
				Name:       newGroupName,
				MonitorID:  mid,
				Enabled:    1,
				Status:     1,
				ExternalID: "",
			}
			if err := repository.AddGroupDAO(newGroup); err == nil {
				if createdGroup, err := repository.SearchGroupsDAO(model.GroupFilter{Query: newGroupName}); err == nil && len(createdGroup) > 0 {
					return createdGroup[len(createdGroup)-1].ID
				}
			}
		}

		return groupID
	}

	newGroupName := fmt.Sprintf("Group %s", externalGroupIDs[0])
	if len(externalGroupNames) > 0 && strings.TrimSpace(externalGroupNames[0]) != "" {
		newGroupName = strings.TrimSpace(externalGroupNames[0])
	}

	newGroup := model.Group{
		Name:       newGroupName,
		ExternalID: externalGroupIDs[0],
		MonitorID:  mid,
		Enabled:    1,
		Status:     1,
	}
	if err := repository.AddGroupDAO(newGroup); err == nil {
		if createdGroup, err := repository.GetGroupByExternalIDDAO(externalGroupIDs[0], mid); err == nil {
			LogService("info", "created group from external metadata", map[string]interface{}{"external_id": externalGroupIDs[0], "group_name": newGroupName, "monitor_id": mid}, nil, "")
			return createdGroup.ID
		}
	}

	return groupID
}

func PullHostsFromMonitorServ(mid uint) (SyncResult, error) {
	return pullHostsFromMonitorServ(mid, true)
}

func pullHostsFromMonitorServ(mid uint, recordHistory bool) (SyncResult, error) {
	LogService("info", "host sync started", map[string]interface{}{"monitor_id": mid}, nil, "")
	result := SyncResult{}
	setMonitorStatusSyncing(mid)

	// Nagare Internal (ID 1) is always healthy
	if mid != 1 {
		_, _ = TestMonitorStatusServ(mid)
	}

	monitor, err := GetMonitorByIDServ(mid)
	if err != nil {
		LogService("error", "pull hosts failed to load monitor", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get monitor: %w", err)
	}

	// Monitor must be active (status == 1 or syncing) to pull hosts
	monitorDomain, _ := repository.GetMonitorByIDDAO(mid)

	// Skip status block for Nagare Internal
	if mid != 1 && (monitorDomain.Status == 0 || monitorDomain.Status == 2) {
		reason := "monitor is not active"
		if monitorDomain.StatusDescription != "" {
			reason = monitorDomain.StatusDescription
		}
		if recordHistory {
			setMonitorStatusErrorWithReason(mid, reason)

			// Mark all hosts as error with monitor down reason
			hosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
			if err == nil {
				for _, host := range hosts {
					setHostStatusErrorWithReason(host.ID, reason)
					items, err := repository.GetItemsByHIDDAO(host.ID)
					if err == nil {
						for _, item := range items {
							_ = repository.UpdateItemStatusAndDescriptionDAO(item.ID, 2, reason)
						}
					}
				}
			}

			LogService("warn", "pull hosts skipped due to monitor not active", map[string]interface{}{"monitor_id": mid, "monitor_status": monitorDomain.Status, "monitor_status_description": reason}, nil, "")
			return result, fmt.Errorf("monitor is not active (status: %d)", monitorDomain.Status)
		}

		LogService("warn", "auto sync proceeding despite monitor status", map[string]interface{}{"monitor_id": mid, "monitor_status": monitorDomain.Status, "monitor_status_description": reason}, nil, "")
	}

	client, err := createMonitorClient(monitor)
	if err != nil {
		LogService("error", "pull hosts failed to create monitor client", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to create monitor client: %w", err)
	}

	// Always attempt authentication to refresh token if needed
	if err := client.Authenticate(context.Background()); err != nil {
		LogService("warn", "authentication failed, attempting with existing token", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		if monitor.AuthToken != "" {
			client.SetAuthToken(monitor.AuthToken)
		} else {
			return result, fmt.Errorf("authentication failed and no existing token: %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	monitorHosts, err := client.GetHosts(ctx)
	if err != nil {
		LogService("error", "pull hosts failed to fetch hosts", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get hosts from monitor: %w", err)
	}
	if len(monitorHosts) == 0 {
		LogService("warn", "monitor returned zero hosts", map[string]interface{}{
			"monitor_id":   mid,
			"monitor_name": monitor.Name,
			"monitor_type": monitor.Type,
			"monitor_url":  monitor.URL,
			"monitor_user": monitor.Username,
		}, nil, "")
	}
	monitorHostIDs := make(map[string]struct{}, len(monitorHosts))
	for _, h := range monitorHosts {
		monitorHostIDs[h.ID] = struct{}{}

		// Dedup check: if this host exists in Nagare Internal (ID 1) AND we already have it in the target monitor, delete the internal one
		if mid != 1 {
			internalHost, err := repository.GetHostByMIDAndHostIDDAO(1, h.ID)
			if err == nil {
				targetHost, err := repository.GetHostByMIDAndHostIDDAO(mid, h.ID)
				if err == nil && targetHost.ID != internalHost.ID {
					LogService("info", "deleting duplicate internal host", map[string]interface{}{"host_name": h.Name, "internal_id": internalHost.ID, "target_id": targetHost.ID}, nil, "")
					_ = repository.DeleteHostByIDDAO(internalHost.ID)
				}
			}
		}
	}

	result.Total = len(monitorHosts)

	now := time.Now().UTC()
	for _, h := range monitorHosts {
		// Get active_available from metadata
		activeAvailable := ""
		if h.Metadata != nil {
			if value, ok := h.Metadata["active_available"]; ok {
				activeAvailable = value
			}
		}
		status := mapMonitorHostStatus(h.Status, activeAvailable)
		statusDesc := ""
		if h.Metadata != nil {
			if value, ok := h.Metadata["status_description"]; ok {
				statusDesc = value
			}
		}

		existingHost, err := repository.GetHostByMIDAndHostIDDAO(mid, h.ID)
		if err != nil {
			// Try global search by external ID to prevent duplicates from other monitors (like Nagare Internal)
			if globalHost, gErr := repository.GetHostByHostIDDAO(h.ID); gErr == nil {
				// If found globally and it's either in Nagare Internal (ID 1) or has no monitor, adopt it
				if globalHost.MonitorID == 1 || globalHost.MonitorID == 0 {
					existingHost = globalHost
					err = nil
					LogService("info", "adopting host from another monitor", map[string]interface{}{"host_name": h.Name, "old_mid": globalHost.MonitorID, "new_mid": mid}, nil, "")
				}
			}
		}

		groupID := uint(0)
		if err == nil {
			groupID = existingHost.GroupID
		}
		groupID = resolveHostGroupIDFromMetadata(mid, h.Metadata, groupID)

		if err == nil {
			// Host exists, update it
			if err := repository.UpdateHostDAO(existingHost.ID, model.Host{
				Name:              h.Name,
				Hostid:            h.ID,
				MonitorID:         mid,
				GroupID:           groupID,
				Description:       h.Description,
				Enabled:           h.Enabled,
				Status:            status,
				StatusDescription: statusDesc,
				ActiveAvailable:   activeAvailable,
				IPAddr:            h.IPAddress,
				SSHUser:           existingHost.SSHUser,
				SSHPassword:       "", // UpdateHostDAO won't update if empty
				SSHPort:           existingHost.SSHPort,
				LastSyncAt:        &now,
				ExternalSource:    monitor.Name,
			}); err != nil {
				setHostStatusErrorWithReason(existingHost.ID, err.Error())
				LogService("error", "pull hosts failed to update host", map[string]interface{}{"monitor_id": mid, "host_id": existingHost.ID, "error": err.Error()}, nil, "")
				result.Failed++
				continue
			}
			if recordHistory {
				if refreshed, err := repository.GetHostByIDDAO(existingHost.ID); err == nil {
					recordHostHistory(refreshed, time.Now().UTC())
				}
			}
			result.Updated++
		} else {
			// Host doesn't exist, add it
			hNew := model.Host{
				Name:              h.Name,
				Hostid:            h.ID,
				MonitorID:         mid,
				GroupID:           groupID,
				Description:       h.Description,
				Enabled:           h.Enabled,
				Status:            status,
				StatusDescription: statusDesc,
				ActiveAvailable:   activeAvailable,
				IPAddr:            h.IPAddress,
				LastSyncAt:        &now,
				ExternalSource:    monitor.Name,
			}
			if err := repository.AddHostDAO(&hNew); err != nil {
				setMonitorStatusError(mid)
				LogService("error", "pull hosts failed to add host", map[string]interface{}{"monitor_id": mid, "host_name": h.Name, "host_external_id": h.ID, "error": err.Error()}, nil, "")
				result.Failed++
				continue
			}
			if recordHistory {
				if created, err := repository.GetHostByMIDAndHostIDDAO(mid, h.ID); err == nil {
					recordHostHistory(created, time.Now().UTC())
				}
			}
			result.Added++
		}

		// Cascade to items for this host in parallel to improve performance
		hostInDB, err := repository.GetHostByMIDAndHostIDDAO(mid, h.ID)
		if err == nil {
			LogService("debug", "triggering parallel item sync for host", map[string]interface{}{"host_id": hostInDB.ID, "name": hostInDB.Name}, nil, "")

			// We use a goroutine for each host's items, but we need to track completion for the final result
			// For simplicity and to avoid race conditions on the 'result' struct,
			// we can use a small concurrency limit for host items sync as well.
			go func(hid uint, hName string) {
				itemsRes, err := PullItemsFromHostServ(mid, hid)
				if err != nil {
					LogService("error", "cascaded item sync failed", map[string]interface{}{"host_id": hid, "name": hName, "error": err.Error()}, nil, "")
				} else {
					LogService("debug", "cascaded item sync finished", map[string]interface{}{"host_id": hid, "name": hName, "added": itemsRes.Added, "updated": itemsRes.Updated}, nil, "")
				}
			}(hostInDB.ID, hostInDB.Name)
		}
	}

	localHosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
	if err == nil {
		// Skip 'not found' check for SNMP monitors as they don't provide a master host list
		if monitors.ParseMonitorType(monitor.Type) != monitors.MonitorSNMP {
			for _, localHost := range localHosts {
				if _, ok := monitorHostIDs[localHost.Hostid]; ok {
					continue
				}
				reason := "host not found on monitor"
				setHostStatusErrorWithReason(localHost.ID, reason)
				items, err := repository.GetItemsByHIDDAO(localHost.ID)
				if err == nil {
					for _, item := range items {
						_ = repository.UpdateItemStatusAndDescriptionDAO(item.ID, 2, reason)
					}
				}
			}
		}
	}
	_ = recomputeMonitorRelated(mid)
	recordNetworkStatusSnapshot(time.Now().UTC())
	SyncEvent("hosts", mid, 0, result)
	LogService("info", "host sync finished", map[string]interface{}{"monitor_id": mid, "added": result.Added, "updated": result.Updated}, nil, "")
	return result, nil
}

func PullHostFromMonitorServ(mid, id uint) (SyncResult, error) {
	result := SyncResult{}
	host, err := repository.GetHostByIDDAO(id)
	if err != nil {
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "pull host failed to load host", map[string]interface{}{"host_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get host: %w", err)
	}
	setHostStatusSyncing(id)
	if host.MonitorID != mid {
		setHostStatusErrorWithReason(id, "host does not belong to the specified monitor")
		LogService("error", "pull host failed due to monitor mismatch", map[string]interface{}{"host_id": id, "monitor_id": mid}, nil, "")
		return SyncResult{}, fmt.Errorf("host does not belong to the specified monitor")
	}

	monitor, err := GetMonitorByIDServ(mid)
	if err != nil {
		setMonitorStatusError(mid)
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "pull host failed to load monitor", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get monitor: %w", err)
	}
	setMonitorStatusSyncing(mid)

	client, err := createMonitorClient(monitor)
	if err != nil {
		setMonitorStatusError(mid)
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "pull host failed to create monitor client", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to create monitor client: %w", err)
	}

	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else {
		if err := client.Authenticate(context.Background()); err != nil {
			setMonitorStatusError(mid)
			setHostStatusErrorWithReason(id, err.Error())
			LogService("error", "pull host failed to authenticate", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
			return SyncResult{}, fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}

	h, err := client.GetHostByID(context.Background(), host.Hostid)
	if err != nil {
		setMonitorStatusError(mid)
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "pull host failed to fetch host", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
		return SyncResult{}, fmt.Errorf("failed to get host from monitor: %w", err)
	}

	if h == nil {
		return SyncResult{}, fmt.Errorf("host %s not found on monitor", host.Hostid)
	}

	// Get active_available from metadata
	activeAvailable := ""
	if h.Metadata != nil {
		if value, ok := h.Metadata["active_available"]; ok {
			activeAvailable = value
		}
	}

	// Check if host already exists
	existingHost, err := repository.GetHostByMIDAndHostIDDAO(mid, h.ID)
	if err != nil {
		// Try global search by external ID to prevent duplicates
		if globalHost, gErr := repository.GetHostByHostIDDAO(h.ID); gErr == nil {
			if globalHost.MonitorID == 1 || globalHost.MonitorID == 0 {
				existingHost = globalHost
				err = nil
				LogService("info", "adopting host during single sync", map[string]interface{}{"host_name": h.Name, "new_mid": mid}, nil, "")
			}
		}
	}

	groupID := uint(0)
	if err == nil {
		groupID = existingHost.GroupID
	}
	groupID = resolveHostGroupIDFromMetadata(mid, h.Metadata, groupID)

	if err == nil {
		// Host exists, update it
		if err := repository.UpdateHostDAO(existingHost.ID, model.Host{
			Name:            h.Name,
			Hostid:          h.ID,
			MonitorID:       mid,
			GroupID:         groupID,
			Enabled:         h.Enabled,
			Status:          mapMonitorHostStatus(h.Status, activeAvailable),
			ActiveAvailable: activeAvailable,
			IPAddr:          h.IPAddress,
			SSHUser:         existingHost.SSHUser,
			SSHPort:         existingHost.SSHPort,
		}); err != nil {
			setHostStatusErrorWithReason(existingHost.ID, err.Error())
			LogService("error", "pull host failed to update host", map[string]interface{}{"monitor_id": mid, "host_id": existingHost.ID, "error": err.Error()}, nil, "")
			return SyncResult{}, fmt.Errorf("failed to update host: %w", err)
		}
		if refreshed, err := repository.GetHostByIDDAO(existingHost.ID); err == nil {
			recordHostHistory(refreshed, time.Now().UTC())
		}
		result.Updated++
			} else {
			// Host doesn't exist, add it
			newHost := model.Host{
				Name:            h.Name,
				Hostid:          h.ID,
				MonitorID:       mid,
				GroupID:         groupID,
				Description:     h.Description,
				Enabled:         h.Enabled,
				Status:          mapMonitorHostStatus(h.Status, activeAvailable),
				ActiveAvailable: activeAvailable,
				IPAddr:          h.IPAddress,
			}
	
		if err := repository.AddHostDAO(&newHost); err != nil {
			setMonitorStatusError(mid)
			LogService("error", "pull host failed to add host", map[string]interface{}{"monitor_id": mid, "host_name": h.Name, "host_external_id": h.ID, "error": err.Error()}, nil, "")
			return SyncResult{}, fmt.Errorf("failed to add host: %w", err)
		}
		if created, err := repository.GetHostByMIDAndHostIDDAO(mid, h.ID); err == nil {
			recordHostHistory(created, time.Now().UTC())
		}
		result.Added++
	}

	_ = recomputeMonitorRelated(mid)
	recordNetworkStatusSnapshot(time.Now().UTC())
	result.Total = 1
	SyncEvent("hosts", mid, 0, result)
	return result, nil
}

// PushHostToMonitorServ pushes a host from local database to remote monitor
func PushHostToMonitorServ(mid uint, id uint) (SyncResult, error) {
	result := SyncResult{}
	host, err := repository.GetHostByIDDAO(id)
	if err != nil {
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "push host failed to load host", map[string]interface{}{"host_id": id, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get host: %w", err)
	}
	setHostStatusSyncing(id)

	if host.MonitorID != mid {
		if host.MonitorID == 0 {
			host.MonitorID = mid
			_ = repository.UpdateHostDAO(host.ID, host)
		} else {
			setHostStatusErrorWithReason(id, "host does not belong to the specified monitor")
			LogService("error", "push host failed due to monitor mismatch", map[string]interface{}{"host_id": id, "monitor_id": mid}, nil, "")
			return result, fmt.Errorf("host does not belong to the specified monitor")
		}
	}

	monitor, err := GetMonitorByIDServ(mid)
	if err != nil {
		setMonitorStatusError(mid)
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "push host failed to load monitor", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get monitor: %w", err)
	}

	if mid != 1 && monitor.Status == 2 {
		reason := "monitor is in error state"
		if monitor.StatusDesc != "" {
			reason = monitor.StatusDesc
		}
		setHostStatusErrorWithReason(id, reason)
		return result, fmt.Errorf("monitor is in error state: %s", reason)
	}

	client, err := createMonitorClient(monitor)
	if err != nil {
		setMonitorStatusError(mid)
		setHostStatusErrorWithReason(id, err.Error())
		LogService("error", "push host failed to create monitor client", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to create monitor client: %w", err)
	}

	if monitor.AuthToken != "" {
		client.SetAuthToken(monitor.AuthToken)
	} else {
		if err := client.Authenticate(context.Background()); err != nil {
			setMonitorStatusError(mid)
			setHostStatusErrorWithReason(id, err.Error())
			LogService("error", "push host failed to authenticate", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
			return result, fmt.Errorf("failed to authenticate with monitor: %w", err)
		}
	}
	// Create host group (group name) then create host in monitor
	extGroupID := ""
	groupName := "Default"
	if host.GroupID > 0 {
		if group, err := repository.GetGroupByIDDAO(host.GroupID); err == nil {
			if group.Name != "" {
				groupName = group.Name
			}
			if group.ExternalID != "" && group.MonitorID == mid {
				extGroupID = group.ExternalID
			} else if group.Name != "" {
				// Fallback: Check/Create by name if ExternalID missing or mismatched
				gid, err := client.CreateHostGroup(context.Background(), group.Name)
				if err != nil {
					setMonitorStatusError(mid)
					setHostStatusErrorWithReason(id, err.Error())
					LogService("error", "push host failed to create host group", map[string]interface{}{"monitor_id": mid, "host_id": id, "group": group.Name, "error": err.Error()}, nil, "")
					return result, fmt.Errorf("failed to create host group: %w", err)
				}
				extGroupID = gid
				// Update group with new ExternalID
				group.ExternalID = extGroupID
				group.MonitorID = mid
				_ = repository.UpdateGroupDAO(group.ID, group)
			}
		}
	}
	// Default group if no group or group creation failed
	if extGroupID == "" {
		gid, err := client.CreateHostGroup(context.Background(), "Default")
		if err != nil {
			return result, fmt.Errorf("failed to create default host group: %w", err)
		}
		extGroupID = gid
	}

	monitorHost := monitors.Host{
		ID:          host.Hostid,
		Name:        host.Name,
		IPAddress:   host.IPAddr,
		Description: host.Description,
		Enabled:     host.Enabled,
		Metadata: map[string]string{
			"groupid":        extGroupID,
			"monitor_type":   fmt.Sprintf("%d", monitor.Type),
			"snmp_community": host.SNMPCommunity,
			"snmp_version":   host.SNMPVersion,
			"snmp_port":      fmt.Sprintf("%d", host.SNMPPort),
		},
	}
	if host.Hostid == "" {
		// Try to find host by name first to avoid duplicates
		if existing, err := client.GetHostByName(context.Background(), host.Name); err == nil && existing != nil && existing.ID != "" {
			host.Hostid = existing.ID
			_ = repository.UpdateHostDAO(host.ID, host)
			monitorHost.ID = existing.ID

			// Update the host since it exists
			_, err = client.UpdateHost(context.Background(), monitorHost)
			if err != nil {
				setMonitorStatusError(mid)
				setHostStatusErrorWithReason(id, err.Error())
				LogService("error", "push host failed to update existing host found by name", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
				return result, fmt.Errorf("failed to update existing host in monitor: %w", err)
			}
		} else {
			// Create new host
			created, err := client.CreateHost(context.Background(), monitorHost)
			if err != nil {
				setMonitorStatusError(mid)
				setHostStatusErrorWithReason(id, err.Error())
				LogService("error", "push host failed to create host", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
				return result, fmt.Errorf("failed to create host in monitor: %w", err)
			}
			if created.ID != "" {
				host.Hostid = created.ID
				_ = repository.UpdateHostDAO(host.ID, host)
			}
		}
	} else {
		if _, err := client.UpdateHost(context.Background(), monitorHost); err != nil {
			setMonitorStatusError(mid)
			setHostStatusErrorWithReason(id, err.Error())
			LogService("error", "push host failed to update host", map[string]interface{}{"monitor_id": mid, "host_id": id, "error": err.Error()}, nil, "")
			return result, fmt.Errorf("failed to update host in monitor: %w", err)
		}
	}
	LogService("info", "push host to monitor", map[string]interface{}{"host_name": host.Name, "host_id": host.Hostid, "monitor": monitor.Name, "group": groupName}, nil, "")

	result.Added++
	result.Total = 1

	_, _ = recomputeHostStatus(id)
	_ = recomputeMonitorRelated(mid)
	return result, nil
}

// PushHostsFromMonitorServ pushes all hosts from local database to remote monitor
// TestSNMPServ tests SNMP connectivity for a host
func TestSNMPServ(hid uint) (SyncResult, error) {
	host, err := repository.GetHostByIDDAO(hid)
	if err != nil {
		return SyncResult{}, err
	}

	if host.MonitorID == 0 {
		return SyncResult{}, fmt.Errorf("host has no monitor assigned")
	}

	monitor, err := repository.GetMonitorByIDDAO(host.MonitorID)
	if err != nil {
		// Fallback: If no monitor assigned, try to find "Nagare Internal"
		internalMonitors, sErr := repository.SearchMonitorsDAO(model.MonitorFilter{Query: "Nagare Internal"})
		if sErr == nil && len(internalMonitors) > 0 {
			monitor = internalMonitors[0]
		} else {
			return SyncResult{}, err
		}
	}

	if monitors.ParseMonitorType(monitor.Type) != monitors.MonitorSNMP {
		return SyncResult{}, fmt.Errorf("host is not monitored via SNMP (monitor type: %d)", monitor.Type)
	}

	// Re-use pullItemsFromHostServ but with recordHistory=false
	return pullItemsFromHostServ(monitor.ID, host.ID, false)
}

func PushHostsFromMonitorServ(mid uint) (SyncResult, error) {
	result := SyncResult{}
	setMonitorStatusSyncing(mid)
	if mid != 1 {
		_, _ = TestMonitorStatusServ(mid)
	}

	// Check monitor status before proceeding with host push
	monitor, err := repository.GetMonitorByIDDAO(mid)
	if err != nil {
		setMonitorStatusError(mid)
		LogService("error", "push hosts failed to load monitor", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get monitor: %w", err)
	}

	// Monitor must be active or inactive (not error) to push hosts
	if mid != 1 && monitor.Status == 2 {
		reason := "monitor is in error state"
		if monitor.StatusDescription != "" {
			reason = monitor.StatusDescription
		}
		setMonitorStatusErrorWithReason(mid, reason)
		LogService("warn", "push hosts skipped due to monitor error", map[string]interface{}{"monitor_id": mid, "monitor_status": monitor.Status, "monitor_status_description": reason}, nil, "")
		return result, fmt.Errorf("monitor is in error state (status: %d)", monitor.Status)
	}

	hosts, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
	if err != nil {
		setMonitorStatusError(mid)
		LogService("error", "push hosts failed to load hosts", map[string]interface{}{"monitor_id": mid, "error": err.Error()}, nil, "")
		return result, fmt.Errorf("failed to get hosts: %w", err)
	}

	result.Total = len(hosts)

	for _, host := range hosts {
		hostResult, err := PushHostToMonitorServ(mid, host.ID)
		if err != nil {
			setHostStatusErrorWithReason(host.ID, err.Error())
			LogService("error", "push hosts failed to push host", map[string]interface{}{"monitor_id": mid, "host_id": host.ID, "error": err.Error()}, nil, "")
			result.Failed++
			continue
		}
		result.Added += hostResult.Added
		result.Updated += hostResult.Updated
		result.Failed += hostResult.Failed
		result.Total += hostResult.Total
	}

	_ = recomputeMonitorRelated(mid)
	return result, nil
}
