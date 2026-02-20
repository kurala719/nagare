package monitors

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

type SnmpProvider struct {
	config Config
}

func NewSnmpProvider(cfg Config) (Provider, error) {
	return &SnmpProvider{config: cfg}, nil
}

func (p *SnmpProvider) Authenticate(ctx context.Context) error { return nil }
func (p *SnmpProvider) GetAuthToken() string { return "" }
func (p *SnmpProvider) SetAuthToken(token string) {}
func (p *SnmpProvider) GetHosts(ctx context.Context) ([]Host, error) { return []Host{}, nil }
func (p *SnmpProvider) GetHostsByGroupID(ctx context.Context, groupID string) ([]Host, error) { return []Host{}, nil }
func (p *SnmpProvider) GetHostByName(ctx context.Context, name string) (*Host, error) { return nil, nil }
func (p *SnmpProvider) GetHostByID(ctx context.Context, hostID string) (*Host, error) { return nil, nil }
func (p *SnmpProvider) CreateHost(ctx context.Context, host Host) (Host, error) { return host, nil }
func (p *SnmpProvider) UpdateHost(ctx context.Context, host Host) (Host, error) { return host, nil }
func (p *SnmpProvider) DeleteHost(ctx context.Context, hostID string) error { return nil }

type SnmpConfig struct {
	Community           string
	Version             string
	Port                int
	V3User              string
	V3AuthPass          string
	V3PrivPass          string
	V3AuthProtocol      string
	V3PrivProtocol      string
	V3SecurityLevel     string
	CustomOIDs          map[string]string // map[OID]Name
}

func normalizeOID(oid string) string {
	if len(oid) > 0 && oid[0] == '.' { return oid[1:] }
	return oid
}

func (p *SnmpProvider) GetItems(ctx context.Context, hostID string) ([]Item, error) {
	config, ok := ctx.Value("snmp_config").(SnmpConfig)
	if !ok {
		config = SnmpConfig{Community: "public", Version: "v2c", Port: 161}
	}
	if config.Port == 0 { config.Port = 161 }
	if config.Community == "" { config.Community = "public" }
	if config.Version == "" { config.Version = "v2c" }

	fmt.Printf("SNMP Debug: Starting high-detail V200 poll for %s\n", hostID)

	oidNames := make(map[string]string)
	addOID := func(oid, name string) { oidNames[normalizeOID(oid)] = name }

	// Essential Hardware Metrics
	addOID(".1.3.6.1.2.1.1.1.0", "sysDescr")
	addOID(".1.3.6.1.2.1.1.3.0", "sysUpTime")
	addOID(".1.3.6.1.2.1.1.5.0", "sysName")
	addOID(".1.3.6.1.4.1.2011.6.3.4.1.3", "hwEntityCpuUsage")
	addOID(".1.3.6.1.4.1.2011.6.1.2.1.1.2", "hwMemoryDevSize")
	addOID(".1.3.6.1.4.1.2011.6.1.2.1.1.3", "hwMemoryDevFree")
	addOID(".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.11", "hwEntityTemperature")
	addOID(".1.3.6.1.4.1.2011.5.25.31.1.1.10.1.7", "hwEntityFanState")

	// Merge Custom OIDs
	for k, v := range config.CustomOIDs {
		if _, exists := oidNames[normalizeOID(k)]; !exists { addOID(k, v) }
	}

	gs := &gosnmp.GoSNMP{
		Target: hostID, Port: uint16(config.Port), Timeout: 10 * time.Second, Retries: 1, MaxOids: 20,
	}

	// Auth Config (v1/v2c/v3 logic)
	if config.Version == "v1" {
		gs.Version = gosnmp.Version1; gs.Community = config.Community
	} else if config.Version == "v3" {
		gs.Version = gosnmp.Version3; gs.SecurityModel = gosnmp.UserSecurityModel
		level := config.V3SecurityLevel
		if level == "" { level = "NoAuthNoPriv" }
		switch level {
		case "AuthNoPriv": gs.MsgFlags = gosnmp.AuthNoPriv
		case "AuthPriv": gs.MsgFlags = gosnmp.AuthPriv
		default: gs.MsgFlags = gosnmp.NoAuthNoPriv
		}
		sp := &gosnmp.UsmSecurityParameters{UserName: config.V3User}
		if level != "NoAuthNoPriv" {
			sp.AuthenticationPassphrase = config.V3AuthPass
			switch config.V3AuthProtocol {
			case "MD5": sp.AuthenticationProtocol = gosnmp.MD5
			default: sp.AuthenticationProtocol = gosnmp.SHA
			}
		}
		if level == "AuthPriv" {
			sp.PrivacyPassphrase = config.V3PrivPass
			switch config.V3PrivProtocol {
			case "DES": sp.PrivacyProtocol = gosnmp.DES
			default: sp.PrivacyProtocol = gosnmp.AES
			}
		}
		gs.SecurityParameters = sp
	} else {
		gs.Version = gosnmp.Version2c; gs.Community = config.Community
	}

	err := gs.Connect()
	if err != nil { return nil, fmt.Errorf("SNMP connect failed: %w", err) }
	defer gs.Conn.Close()

	var targetOids []string
	for rawOid := range oidNames { targetOids = append(targetOids, "."+rawOid) }

	var items []Item
	var memSize, memFree float64
	validBitrateCount := 0

	// Batch Fetch
	for i := 0; i < len(targetOids); i += 20 {
		end := i + 20
		if end > len(targetOids) { end = len(targetOids) }
		result, err := gs.Get(targetOids[i:end])
		if err != nil { continue }

		for _, variable := range result.Variables {
			normalized := normalizeOID(variable.Name)
			name := oidNames[normalized]
			if name == "" { name = normalized }
			val := formatSnmpValue(variable)
			
			if (val == "N/A" || val == "" || val == "0") && isV200HardwareMetric(name) {
				discoveredVal := p.discoverV200Metric(gs, normalized)
				if discoveredVal != "" { val = discoveredVal }
			}

			if strings.Contains(name, "Bitrate") && val != "N/A" && val != "" { validBitrateCount++ }
			if name == "hwMemoryDevSize" { memSize = float64(gosnmp.ToBigInt(variable.Value).Int64()) }
			if name == "hwMemoryDevFree" { memFree = float64(gosnmp.ToBigInt(variable.Value).Int64()) }

			items = append(items, Item{
				ID: normalized, Name: name, Key: normalized, Value: val, Timestamp: time.Now().Unix(), Status: "0",
			})
		}
	}

	if len(items) == 0 && len(targetOids) > 0 {
		walkItems, _ := p.pollViaWalk(gs, oidNames, hostID)
		items = append(items, walkItems...)
	}

	// Memory Percentage
	if memSize > 0 && memFree > 0 && memFree <= memSize {
		usage := ((memSize - memFree) / memSize) * 100
		items = append(items, Item{
			ID: "calculated.mem.usage", Name: "Memory Usage (%)", Key: "mem_usage_pct", Value: fmt.Sprintf("%.2f", usage), Timestamp: time.Now().Unix(), Status: "0",
		})
	}

	// Dynamic Deep Discovery (Run if bitrates are missing or first time)
	hasCustomBitrates := false
	expectedBitrateCount := 0
	for k := range config.CustomOIDs {
		if strings.Contains(strings.ToLower(config.CustomOIDs[k]), "bitrate") {
			hasCustomBitrates = true; expectedBitrateCount++
		}
	}

	if len(config.CustomOIDs) == 0 || (hasCustomBitrates && validBitrateCount < (expectedBitrateCount/2)) {
		fmt.Printf("SNMP Debug: Starting deep V200 discovery for %s...\n", hostID)
		ifaceItems, _ := p.discoverInterfaces(gs)
		routingItems, _ := p.discoverRoutingMetrics(gs)
		lldpItems, _ := p.discoverLLDPNeighbors(gs)
		sfpItems, _ := p.discoverSFPDetails(gs)
		invItems, _ := p.discoverInventory(gs)
		
		discovered := append(ifaceItems, routingItems...)
		discovered = append(discovered, lldpItems...)
		discovered = append(discovered, sfpItems...)
		discovered = append(discovered, invItems...)
		
		existingIDs := make(map[string]bool)
		for _, it := range items { existingIDs[it.ID] = true }
		for _, it := range discovered {
			if !existingIDs[it.ID] { items = append(items, it) }
		}
	}

	return items, nil
}

func isV200HardwareMetric(name string) bool {
	return name == "hwCpuDevDuty" || name == "hwMemoryDevSize" || name == "hwMemoryDevFree" || 
	       name == "hwEntityTemperature" || name == "hwEntityFanState" ||
	       name == "hwEntityCpuUsage" || name == "hwEntityMemUsage"
}

func (p *SnmpProvider) discoverV200Metric(gs *gosnmp.GoSNMP, baseOid string) string {
	val := ""
	_ = gs.Walk("."+baseOid, func(v gosnmp.SnmpPDU) error {
		walkVal := formatSnmpValue(v)
		if walkVal != "N/A" && walkVal != "" && walkVal != "0" {
			val = walkVal; return fmt.Errorf("found") 
		}
		return nil
	})
	return val
}

func (p *SnmpProvider) discoverInventory(gs *gosnmp.GoSNMP) ([]Item, error) {
	var items []Item
	// entPhysicalSerialNum (.1.3.6.1.2.1.47.1.1.1.1.11)
	// Usually index 1 is the main chassis
	_ = gs.Walk(".1.3.6.1.2.1.47.1.1.1.1.11", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		if val == "" || val == "N/A" { return nil }
		items = append(items, Item{
			ID: normalizeOID(v.Name), Name: "Hardware Serial Number", Key: normalizeOID(v.Name), Value: val, Status: "0", Timestamp: time.Now().Unix(),
		})
		return fmt.Errorf("found") // Just take the first valid serial (chassis)
	})
	// entPhysicalModelName (.1.3.6.1.2.1.47.1.1.1.1.13)
	_ = gs.Walk(".1.3.6.1.2.1.47.1.1.1.1.13", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		if val == "" || val == "N/A" { return nil }
		items = append(items, Item{
			ID: normalizeOID(v.Name), Name: "Hardware Model", Key: normalizeOID(v.Name), Value: val, Status: "0", Timestamp: time.Now().Unix(),
		})
		return fmt.Errorf("found")
	})
	return items, nil
}

func (p *SnmpProvider) discoverInterfaces(gs *gosnmp.GoSNMP) ([]Item, error) {
	var items []Item
	ifNames := make(map[string]string)
	handlePDU := func(v gosnmp.SnmpPDU) error {
		if v.Type == gosnmp.OctetString {
			if bytes, ok := v.Value.([]byte); ok {
				val := string(bytes)
				if val != "" {
					idx := v.Name[strings.LastIndex(v.Name, ".")+1:]
					ifNames[idx] = val
				}
			}
		}
		return nil
	}
	_ = gs.BulkWalk(".1.3.6.1.2.1.31.1.1.1.1", handlePDU)
	if len(ifNames) == 0 { _ = gs.Walk(".1.3.6.1.2.1.2.2.1.2", handlePDU) }
	if len(ifNames) == 0 { return items, nil }

	var fetchOids []string
	oidToMeta := make(map[string]struct{Name, Type string})
	for idx, name := range ifNames {
		lowerName := strings.ToLower(name)
		if !strings.Contains(lowerName, "gigabit") && !strings.Contains(lowerName, "ten-gigabit") && 
		   !strings.Contains(lowerName, "eth-trunk") && !strings.Contains(lowerName, "ge") && 
		   !strings.Contains(lowerName, "xge") && !strings.Contains(lowerName, "meth") {
			continue
		}
		inOid := ".1.3.6.1.2.1.31.1.1.1.6." + idx
		outOid := ".1.3.6.1.2.1.31.1.1.1.10." + idx
		statOid := ".1.3.6.1.2.1.2.2.1.8." + idx
		speedOid := ".1.3.6.1.2.1.31.1.1.1.15." + idx

		fetchOids = append(fetchOids, inOid, outOid, statOid, speedOid)
		oidToMeta[normalizeOID(inOid)] = struct{Name, Type string}{name + " Inbound Bitrate", "traffic"}
		oidToMeta[normalizeOID(outOid)] = struct{Name, Type string}{name + " Outbound Bitrate", "traffic"}
		oidToMeta[normalizeOID(statOid)] = struct{Name, Type string}{name + " Status", "status"}
		oidToMeta[normalizeOID(speedOid)] = struct{Name, Type string}{name + " Speed (Mbps)", "speed"}
	}

	for i := 0; i < len(fetchOids); i += 20 {
		end := i + 20
		if end > len(fetchOids) { end = len(fetchOids) }
		result, err := gs.Get(fetchOids[i:end])
		if err != nil { continue }
		for _, v := range result.Variables {
			val := formatSnmpValue(v)
			normOid := normalizeOID(v.Name)
			meta := oidToMeta[normOid]
			if meta.Name == "" { continue }
			if meta.Type == "status" {
				switch val {
				case "1": val = "Up"; case "2": val = "Down"; default: val = "Unknown"
				}
			}
			items = append(items, Item{
				ID: normOid, Name: meta.Name, Key: normOid, Value: val, Status: "0", Timestamp: time.Now().Unix(),
			})
		}
	}
	return items, nil
}

func (p *SnmpProvider) discoverRoutingMetrics(gs *gosnmp.GoSNMP) ([]Item, error) {
	var items []Item
	_ = gs.Walk(".1.3.6.1.2.1.14.10.1.6", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v); normOid := normalizeOID(v.Name)
		parts := strings.Split(normOid, "."); nbrIp := "Unknown"
		if len(parts) >= 4 { nbrIp = strings.Join(parts[len(parts)-4:], ".") }
		switch val {
		case "8": val = "Full"; case "1": val = "Down"; default: val = "Syncing (" + val + ")"
		}
		items = append(items, Item{
			ID: normOid, Name: "OSPF Neighbor " + nbrIp + " State", Key: normOid, Value: val, Status: "0", Timestamp: time.Now().Unix(),
		})
		return nil
	})
	return items, nil
}

func (p *SnmpProvider) discoverLLDPNeighbors(gs *gosnmp.GoSNMP) ([]Item, error) {
	var items []Item
	_ = gs.Walk(".1.0.8802.1.1.2.1.4.1.1.9", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v); normOid := normalizeOID(v.Name)
		if val == "" || val == "N/A" { return nil }
		items = append(items, Item{
			ID: normOid, Name: "LLDP Neighbor " + val, Key: normOid, Value: "Connected", Status: "0", Timestamp: time.Now().Unix(),
		})
		return nil
	})
	return items, nil
}

func (p *SnmpProvider) discoverSFPDetails(gs *gosnmp.GoSNMP) ([]Item, error) {
	var items []Item
	_ = gs.Walk(".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.22", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v); normOid := normalizeOID(v.Name)
		if val == "" || val == "N/A" || val == "0" { return nil }
		
		// Scale dBm: VRP usually sends centi-dBm (e.g. -1500 = -15.00)
		numVal := 0.0
		fmt.Sscanf(val, "%f", &numVal)
		if numVal != 0 { val = fmt.Sprintf("%.2f dBm", numVal/100.0) }

		idx := normOid[strings.LastIndex(normOid, ".")+1:]
		items = append(items, Item{
			ID: normOid, Name: "SFP Port " + idx + " RX Power", Key: normOid, Value: val, Status: "0", Timestamp: time.Now().Unix(),
		})
		// TX Power
		txOid := "1.3.6.1.4.1.2011.5.25.31.1.1.1.1.17." + idx
		items = append(items, Item{
			ID: txOid, Name: "SFP Port " + idx + " TX Power", Key: txOid, Value: "Checking...", Status: "0", Timestamp: time.Now().Unix(),
		})
		return nil
	})
	return items, nil
}

func (p *SnmpProvider) pollViaWalk(gs *gosnmp.GoSNMP, oidNames map[string]string, hostID string) ([]Item, error) {
	var items []Item
	for rawOid, name := range oidNames {
		_ = gs.Walk("."+rawOid, func(v gosnmp.SnmpPDU) error {
			norm := normalizeOID(v.Name)
			items = append(items, Item{
				ID: norm, Name: name, Key: norm, Value: formatSnmpValue(v), Timestamp: time.Now().Unix(), Status: "0",
			})
			return fmt.Errorf("found") 
		})
	}
	return items, nil
}

func formatSnmpValue(variable gosnmp.SnmpPDU) string {
	switch variable.Type {
	case gosnmp.OctetString:
		if bytes, ok := variable.Value.([]byte); ok { return string(bytes) }
		return fmt.Sprintf("%v", variable.Value)
	case gosnmp.Integer, gosnmp.Counter32, gosnmp.Gauge32, gosnmp.Counter64, gosnmp.TimeTicks:
		return fmt.Sprintf("%v", gosnmp.ToBigInt(variable.Value))
	default: return "N/A"
	}
}

func (p *SnmpProvider) GetItemByID(ctx context.Context, itemID string) (*Item, error) { return nil, nil }
func (p *SnmpProvider) GetItemHistory(ctx context.Context, itemID string, from, to int64) ([]Item, error) { return []Item{}, nil }
func (p *SnmpProvider) CreateItem(ctx context.Context, item Item) (Item, error) { return item, nil }
func (p *SnmpProvider) UpdateItem(ctx context.Context, item Item) (Item, error) { return item, nil }
func (p *SnmpProvider) DeleteItem(ctx context.Context, itemID string) error { return nil }
func (p *SnmpProvider) GetAlerts(ctx context.Context) ([]Alert, error) { return []Alert{}, nil }
func (p *SnmpProvider) GetAlertsByHost(ctx context.Context, hostID string) ([]Alert, error) { return []Alert{}, nil }
func (p *SnmpProvider) GetTriggers(ctx context.Context) ([]Trigger, error) { return []Trigger{}, nil }
func (p *SnmpProvider) GetTriggersByHost(ctx context.Context, hostID string) ([]Trigger, error) { return []Trigger{}, nil }
func (p *SnmpProvider) GetTemplateidByName(ctx context.Context, name string) ([]string, error) { return []string{}, nil }
func (p *SnmpProvider) GetHostGroups(ctx context.Context) ([]string, error) { return []string{}, nil }
func (p *SnmpProvider) GetHostGroupsDetails(ctx context.Context) ([]struct{ ID, Name string }, error) { return nil, nil }
func (p *SnmpProvider) GetHostGroupByName(ctx context.Context, name string) (string, error) { return "", nil }
func (p *SnmpProvider) CreateHostGroup(ctx context.Context, name string) (string, error) { return "", nil }
func (p *SnmpProvider) UpdateHostGroup(ctx context.Context, id, name string) error { return nil }
func (p *SnmpProvider) DeleteHostGroup(ctx context.Context, id string) error { return nil }
func (p *SnmpProvider) CreateMediaType(ctx context.Context, name string, script string, params map[string]string) error { return nil }
func (p *SnmpProvider) GetMediaTypeIDByName(ctx context.Context, name string) (string, error) { return "", nil }
func (p *SnmpProvider) GetUserIDByUsername(ctx context.Context, username string) (string, error) { return "", nil }
func (p *SnmpProvider) EnsureUserMedia(ctx context.Context, userID string, mediaTypeID string, sendTo string) error { return nil }
func (p *SnmpProvider) EnsureActionWithMedia(ctx context.Context, name string, userID string, mediaTypeID string) error { return nil }
func (p *SnmpProvider) Name() string { return p.config.Name }
func (p *SnmpProvider) Type() MonitorType { return MonitorSNMP }
