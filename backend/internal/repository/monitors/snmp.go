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

func (p *SnmpProvider) Authenticate(ctx context.Context) error       { return nil }
func (p *SnmpProvider) GetAuthToken() string                         { return "" }
func (p *SnmpProvider) SetAuthToken(token string)                    {}
func (p *SnmpProvider) GetHosts(ctx context.Context) ([]Host, error) { return []Host{}, nil }
func (p *SnmpProvider) GetHostsByGroupID(ctx context.Context, groupID string) ([]Host, error) {
	return []Host{}, nil
}
func (p *SnmpProvider) GetHostByName(ctx context.Context, name string) (*Host, error) {
	return nil, nil
}
func (p *SnmpProvider) GetHostByID(ctx context.Context, hostID string) (*Host, error) {
	return nil, nil
}
func (p *SnmpProvider) CreateHost(ctx context.Context, host Host) (Host, error) { return host, nil }
func (p *SnmpProvider) UpdateHost(ctx context.Context, host Host) (Host, error) { return host, nil }
func (p *SnmpProvider) DeleteHost(ctx context.Context, hostID string) error     { return nil }

type SnmpConfig struct {
	Community       string
	Version         string
	Port            int
	V3User          string
	V3AuthPass      string
	V3PrivPass      string
	V3AuthProtocol  string
	V3PrivProtocol  string
	V3SecurityLevel string
	CustomOIDs      map[string]string // map[OID]Name
}

func normalizeOID(oid string) string {
	if len(oid) > 0 && oid[0] == '.' {
		return oid[1:]
	}
	return oid
}

func (p *SnmpProvider) GetItems(ctx context.Context, hostID string) ([]Item, error) {
	config, ok := ctx.Value("snmp_config").(SnmpConfig)
	if !ok {
		config = SnmpConfig{Community: "public", Version: "v2c", Port: 161}
	}
	if config.Port == 0 {
		config.Port = 161
	}
	if config.Community == "" {
		config.Community = "public"
	}
	if config.Version == "" {
		config.Version = "v2c"
	}

	oidNames := make(map[string]string)

	// MANDATORY NAMING MAP (Forces professional English names in Frontend)
	translate := map[string]string{
		"1.3.6.1.2.1.1.1.0":                   "System Version Details",
		"1.3.6.1.2.1.1.3.0":                   "System Uptime",
		"1.3.6.1.2.1.1.5.0":                   "System Hostname",
		"1.3.6.1.4.1.2011.6.3.4.1.3":          "CPU Usage (%)",
		"1.3.6.1.4.1.2011.6.1.2.1.1.2":        "Physical Memory Capacity",
		"1.3.6.1.4.1.2011.6.1.2.1.1.3":        "Physical Memory Available",
		"1.3.6.1.4.1.2011.6.3.4.1.1":          "Core Temperature",
		"1.3.6.1.4.1.2011.5.25.31.1.1.10.1.7": "Fan Tray Status",
	}

	for k, v := range translate {
		oidNames[k] = v
	}
	for k, v := range config.CustomOIDs {
		norm := normalizeOID(k)
		if _, exists := oidNames[norm]; !exists {
			oidNames[norm] = v
		}
	}

	gs := &gosnmp.GoSNMP{
		Target: hostID, Port: uint16(config.Port), Timeout: 10 * time.Second, Retries: 1, MaxOids: 20,
	}

	// Auth Config (v1/v2c/v3 logic)
	switch config.Version {
	case "v3":
		gs.Version = gosnmp.Version3
		gs.SecurityModel = gosnmp.UserSecurityModel
		level := config.V3SecurityLevel
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
		sp := &gosnmp.UsmSecurityParameters{UserName: config.V3User}
		if level != "NoAuthNoPriv" {
			sp.AuthenticationPassphrase = config.V3AuthPass
			switch config.V3AuthProtocol {
			case "MD5":
				sp.AuthenticationProtocol = gosnmp.MD5
			default:
				sp.AuthenticationProtocol = gosnmp.SHA
			}
		}
		if level == "AuthPriv" {
			sp.PrivacyPassphrase = config.V3PrivPass
			switch config.V3PrivProtocol {
			case "DES":
				sp.PrivacyProtocol = gosnmp.DES
			default:
				sp.PrivacyProtocol = gosnmp.AES
			}
		}
		gs.SecurityParameters = sp
	case "v1":
		gs.Version = gosnmp.Version1
		gs.Community = config.Community
	default:
		gs.Version = gosnmp.Version2c
		gs.Community = config.Community
	}

	err := gs.Connect()
	if err != nil {
		return nil, fmt.Errorf("SNMP connect failed: %w", err)
	}
	defer gs.Conn.Close()

	// Entity Map for hardware discovery
	entityMap := make(map[string]string)
	_ = gs.BulkWalk(".1.3.6.1.2.1.47.1.1.1.1.7", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		if val != "" && val != "N/A" {
			idx := v.Name[strings.LastIndex(v.Name, ".")+1:]
			entityMap[idx] = val
		}
		return nil
	})

	var targetOids []string
	for rawOid := range oidNames {
		targetOids = append(targetOids, "."+rawOid)
	}

	var items []Item
	var memSize, memFree float64

	// Resilient Batch Fetch
	for i := 0; i < len(targetOids); i += 20 {
		end := i + 20
		if end > len(targetOids) {
			end = len(targetOids)
		}
		batch := targetOids[i:end]
		result, err := gs.Get(batch)

		if err != nil {
			for _, singleOid := range batch {
				sRes, sErr := gs.Get([]string{singleOid})
				if sErr == nil && len(sRes.Variables) > 0 {
					p.processPDU(sRes.Variables[0], oidNames, &items, &memSize, &memFree, gs)
				}
			}
			continue
		}

		for _, variable := range result.Variables {
			p.processPDU(variable, oidNames, &items, &memSize, &memFree, gs)
		}
	}

	// Memory Percentage (S5700 Fix)
	if memSize > 0 && memFree >= 0 {
		if memFree > memSize {
			memFree = memFree / 1024
		}
		// Ensure memFree is not larger than memSize after correction
		if memFree > memSize {
			memFree = memSize
		}
		usage := ((memSize - memFree) / memSize) * 100
		items = append(items, Item{
			ID: "calculated.mem.usage", Name: "Memory Usage (%)", Key: "mem_usage_pct", Value: fmt.Sprintf("%.2f", usage), Units: "%", Timestamp: time.Now().Unix(), Status: "0",
		})
	}

	// ENFORCED DEEP forensic DISCOVERY

	ifaceItems, _ := p.discoverInterfaces(gs)
	routingItems, _ := p.discoverRoutingMetrics(gs)
	bgpItems, _ := p.discoverBGPMetrics(gs)
	stpItems, _ := p.discoverSTPMetrics(gs)
	lldpItems, _ := p.discoverLLDPNeighbors(gs)
	sfpItems, _ := p.discoverSFPDetails(gs, entityMap)
	invItems, _ := p.discoverInventory(gs)
	healthItems, _ := p.discoverAdvancedHealth(gs, entityMap)

	discovered := append(ifaceItems, routingItems...)
	discovered = append(discovered, lldpItems...)
	discovered = append(discovered, sfpItems...)
	discovered = append(discovered, invItems...)
	discovered = append(discovered, healthItems...)
	discovered = append(discovered, bgpItems...)
	discovered = append(discovered, stpItems...)

	existingIDs := make(map[string]bool)
	for _, it := range items {
		existingIDs[it.ID] = true
	}
	for _, it := range discovered {
		if !existingIDs[it.ID] {
			items = append(items, it)
		}
	}

	return items, nil
}

func (p *SnmpProvider) processPDU(v gosnmp.SnmpPDU, oidNames map[string]string, items *[]Item, memSize, memFree *float64, gs *gosnmp.GoSNMP) {
	norm := normalizeOID(v.Name)
	name := oidNames[norm]
	if name == "" {
		name = norm
	}
	val := formatSnmpValue(v)

	// Smart-Probe for V200 Hardware metrics (find active board index)
	if (val == "N/A" || val == "" || val == "0") && isV200HardwareMetric(name) {
		root := norm
		if idx := strings.LastIndex(root, "."); idx != -1 {
			root = root[:idx]
		}
		discoveredVal := p.discoverV200Metric(gs, root)
		if discoveredVal != "" {
			val = discoveredVal
		}
	}

	if strings.Contains(name, "Memory Capacity") {
		*memSize = float64(gosnmp.ToBigInt(v.Value).Int64())
	}
	if strings.Contains(name, "Memory Available") {
		*memFree = float64(gosnmp.ToBigInt(v.Value).Int64())
	}

	*items = append(*items, Item{
		ID: norm, Name: name, Key: norm, Value: val, Units: determineUnits(name), Timestamp: time.Now().Unix(), Status: "0",
	})
}

func determineUnits(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "percentage") || strings.Contains(lower, "(%)") || strings.Contains(lower, "usage") {
		return "%"
	}
	if strings.Contains(lower, "temperature") {
		return "°C"
	}
	if strings.Contains(lower, "bitrate") || strings.Contains(lower, "speed") || strings.Contains(lower, "rate") {
		return "bps"
	}
	if strings.Contains(lower, "capacity") || strings.Contains(lower, "available") || strings.Contains(lower, "total") || strings.Contains(lower, "free") {
		return "B"
	}
	if strings.Contains(lower, "power") && strings.Contains(lower, "dbm") {
		return "dBm"
	}
	if strings.Contains(lower, "errors") || strings.Contains(lower, "discards") || strings.Contains(lower, "count") {
		return "pkts"
	}
	return ""
}

func isV200HardwareMetric(name string) bool {
	return strings.Contains(name, "CPU Usage") || strings.Contains(name, "Memory") ||
		strings.Contains(name, "Temperature") || strings.Contains(name, "Fan Status")
}

func (p *SnmpProvider) discoverV200Metric(gs *gosnmp.GoSNMP, baseOid string) string {
	val := ""
	_ = gs.Walk("."+baseOid, func(v gosnmp.SnmpPDU) error {
		walkVal := formatSnmpValue(v)
		if walkVal != "N/A" && walkVal != "" && walkVal != "0" {
			val = walkVal
			return fmt.Errorf("found")
		}
		return nil
	})
	return val
}

func (p *SnmpProvider) discoverInventory(gs *gosnmp.GoSNMP) ([]Item, error) {
	var items []Item
	_ = gs.Walk(".1.3.6.1.2.1.47.1.1.1.1.11", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		if val == "" || val == "N/A" {
			return nil
		}
		items = append(items, Item{ID: normalizeOID(v.Name), Name: "Hardware Serial Number", Key: normalizeOID(v.Name), Value: val, Status: "0", Timestamp: time.Now().Unix()})
		return fmt.Errorf("found")
	})
	_ = gs.Walk(".1.3.6.1.2.1.47.1.1.1.1.13", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		if val == "" || val == "N/A" {
			return nil
		}
		items = append(items, Item{ID: normalizeOID(v.Name), Name: "Hardware Model Identifier", Key: normalizeOID(v.Name), Value: val, Status: "0", Timestamp: time.Now().Unix()})
		return fmt.Errorf("found")
	})
	return items, nil
}

func (p *SnmpProvider) discoverAdvancedHealth(gs *gosnmp.GoSNMP, entityMap map[string]string) ([]Item, error) {
	var items []Item
	_ = gs.Walk(".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.13", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		idx := v.Name[strings.LastIndex(v.Name, ".")+1:]
		pretty := entityMap[idx]
		if pretty == "" {
			pretty = "Power Module " + idx
		}
		switch val {
		case "1":
			val = "Normal"
		case "2":
			val = "Abnormal"
		case "3":
			val = "Not Supplied"
		}
		items = append(items, Item{ID: normalizeOID(v.Name), Name: pretty + " Health State", Key: normalizeOID(v.Name), Value: val, Status: "0", Timestamp: time.Now().Unix()})
		return nil
	})
	// MAC table count
	_ = gs.Walk(".1.3.6.1.4.1.2011.5.25.42.2.1.1.1.1.1.1", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		items = append(items, Item{ID: normalizeOID(v.Name), Name: "Total MAC Address Count", Key: normalizeOID(v.Name), Value: val, Status: "0", Timestamp: time.Now().Unix()})
		return fmt.Errorf("found")
	})
	return items, nil
}

func (p *SnmpProvider) discoverBGPMetrics(gs *gosnmp.GoSNMP) ([]Item, error) {
	var items []Item
	_ = gs.Walk(".1.3.6.1.2.1.15.3.1.2", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		norm := normalizeOID(v.Name)
		parts := strings.Split(norm, ".")
		nbrIp := "Unknown"
		if len(parts) >= 4 {
			nbrIp = strings.Join(parts[len(parts)-4:], ".")
		}
		switch val {
		case "6":
			val = "Established"
		case "1":
			val = "Idle"
		default:
			val = "Connecting"
		}
		items = append(items, Item{ID: norm, Name: "BGP Neighbor [" + nbrIp + "] Session", Key: norm, Value: val, Status: "0", Timestamp: time.Now().Unix()})
		return nil
	})
	return items, nil
}

func (p *SnmpProvider) discoverSTPMetrics(gs *gosnmp.GoSNMP) ([]Item, error) {
	var items []Item
	_ = gs.Walk(".1.3.6.1.2.1.17.2.15.1.3", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		norm := normalizeOID(v.Name)
		idx := norm[strings.LastIndex(norm, ".")+1:]
		switch val {
		case "5":
			val = "Forwarding"
		case "2":
			val = "Blocking"
		case "1":
			val = "Disabled"
		default:
			val = "Learning"
		}
		items = append(items, Item{ID: norm, Name: "STP Port [" + idx + "] Operational State", Key: norm, Value: val, Status: "0", Timestamp: time.Now().Unix()})
		return nil
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
	if len(ifNames) == 0 {
		_ = gs.Walk(".1.3.6.1.2.1.2.2.1.2", handlePDU)
	}
	if len(ifNames) == 0 {
		return items, nil
	}

	var fetchOids []string
	oidToMeta := make(map[string]struct{ Name, Type string })
	for idx, name := range ifNames {
		lowerName := strings.ToLower(name)
		if !strings.Contains(lowerName, "gigabit") && !strings.Contains(lowerName, "ten-gigabit") &&
			!strings.Contains(lowerName, "eth-trunk") && !strings.Contains(lowerName, "ge") &&
			!strings.Contains(lowerName, "xge") && !strings.Contains(lowerName, "meth") &&
			!strings.Contains(lowerName, "ethernet") {
			continue
		}
		// Huawei V5 Instant Rate OIDs (match display interface exactly)
		inOid := ".1.3.6.1.4.1.2011.5.25.41.1.1.1.1.3." + idx
		outOid := ".1.3.6.1.4.1.2011.5.25.41.1.1.1.1.4." + idx
		statOid := ".1.3.6.1.2.1.2.2.1.8." + idx
		speedOid := ".1.3.6.1.2.1.31.1.1.1.15." + idx

		fetchOids = append(fetchOids, inOid, outOid, statOid, speedOid)
		oidToMeta[normalizeOID(inOid)] = struct{ Name, Type string }{name + " Inbound Rate", "traffic"}
		oidToMeta[normalizeOID(outOid)] = struct{ Name, Type string }{name + " Outbound Rate", "traffic"}
		oidToMeta[normalizeOID(statOid)] = struct{ Name, Type string }{name + " Link Status", "status"}
		oidToMeta[normalizeOID(speedOid)] = struct{ Name, Type string }{name + " Negotiated Speed", "speed"}
	}

	for i := 0; i < len(fetchOids); i += 20 {
		end := i + 20
		if end > len(fetchOids) {
			end = len(fetchOids)
		}
		result, err := gs.Get(fetchOids[i:end])
		if err != nil {
			continue
		}
		for _, v := range result.Variables {
			val := formatSnmpValue(v)
			normOid := normalizeOID(v.Name)
			meta := oidToMeta[normOid]
			if meta.Name == "" {
				continue
			}
			if meta.Type == "status" {
				switch val {
				case "1":
					val = "Up"
				case "2":
					val = "Down"
				default:
					val = "Unknown"
				}
			}
			items = append(items, Item{ID: normOid, Name: meta.Name, Key: normOid, Value: val, Units: determineUnits(meta.Name), Status: "0", Timestamp: time.Now().Unix()})
		}
	}
	return items, nil
}

func (p *SnmpProvider) discoverRoutingMetrics(gs *gosnmp.GoSNMP) ([]Item, error) {
	var items []Item
	_ = gs.Walk(".1.3.6.1.2.1.14.10.1.6", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		normOid := normalizeOID(v.Name)
		parts := strings.Split(normOid, ".")
		nbrIp := "Unknown"
		if len(parts) >= 4 {
			nbrIp = strings.Join(parts[len(parts)-4:], ".")
		}
		switch val {
		case "8":
			val = "Full"
		case "1":
			val = "Down"
		default:
			val = "Syncing"
		}
		items = append(items, Item{ID: normOid, Name: "OSPF Neighbor [" + nbrIp + "] Session", Key: normOid, Value: val, Status: "0", Timestamp: time.Now().Unix()})
		return nil
	})
	return items, nil
}

func (p *SnmpProvider) discoverLLDPNeighbors(gs *gosnmp.GoSNMP) ([]Item, error) {
	var items []Item
	_ = gs.Walk(".1.0.8802.1.1.2.1.4.1.1.9", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		normOid := normalizeOID(v.Name)
		if val == "" || val == "N/A" {
			return nil
		}
		items = append(items, Item{ID: normOid, Name: "LLDP Neighbor Device [" + val + "]", Key: normOid, Value: "Connected", Status: "0", Timestamp: time.Now().Unix()})
		return nil
	})
	return items, nil
}

func (p *SnmpProvider) discoverSFPDetails(gs *gosnmp.GoSNMP, entityMap map[string]string) ([]Item, error) {
	var items []Item
	_ = gs.Walk(".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.22", func(v gosnmp.SnmpPDU) error {
		val := formatSnmpValue(v)
		normOid := normalizeOID(v.Name)
		if val == "" || val == "N/A" {
			return nil
		}

		// Parse SFP power (Unit is 0.01 dBm)
		numVal := 0.0
		n, _ := fmt.Sscanf(val, "%f", &numVal)
		// If valid number parsed
		if n > 0 {
			// Convert to dBm
			val = fmt.Sprintf("%.2f", numVal/100.0)
		} else if val == "0" {
			// If exactly 0, it likely means no signal or unplugged, keep as 0.00
			val = "0.00"
		}

		idx := normOid[strings.LastIndex(normOid, ".")+1:]
		pretty := entityMap[idx]
		if pretty == "" {
			pretty = "SFP Port " + idx
		}
		items = append(items, Item{ID: normOid, Name: pretty + " RX Power Level", Key: normOid, Value: val, Units: "dBm", Status: "0", Timestamp: time.Now().Unix()})
		return nil
	})
	return items, nil
}

func (p *SnmpProvider) pollViaWalk(gs *gosnmp.GoSNMP, oidNames map[string]string) ([]Item, error) {
	var items []Item
	for rawOid, name := range oidNames {
		_ = gs.Walk("."+rawOid, func(v gosnmp.SnmpPDU) error {
			norm := normalizeOID(v.Name)
			items = append(items, Item{ID: norm, Name: name, Key: norm, Value: formatSnmpValue(v), Timestamp: time.Now().Unix(), Status: "0"})
			return fmt.Errorf("found")
		})
	}
	return items, nil
}

func formatSnmpValue(variable gosnmp.SnmpPDU) string {
	switch variable.Type {
	case gosnmp.OctetString:
		bytes, ok := variable.Value.([]byte)
		if !ok {
			return ""
		}
		// Check if it contains binary data (non-printable)
		isBinary := false
		for _, b := range bytes {
			if (b < 32 && b != 9 && b != 10 && b != 13) || b > 126 {
				isBinary = true
				break
			}
		}
		if isBinary {
			return fmt.Sprintf("%x", bytes)
		}
		return string(bytes)
	case gosnmp.Integer, gosnmp.Counter32, gosnmp.Gauge32, gosnmp.Counter64, gosnmp.TimeTicks:
		return fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value))
	case gosnmp.ObjectIdentifier:
		return fmt.Sprintf("%s", variable.Value)
	case gosnmp.IPAddress:
		return fmt.Sprintf("%s", variable.Value)
	case gosnmp.Null:
		return "N/A"
	default:
		return fmt.Sprintf("%v", variable.Value)
	}
}

func (p *SnmpProvider) GetItemByID(ctx context.Context, itemID string) (*Item, error) {
	return nil, nil
}
func (p *SnmpProvider) GetItemHistory(ctx context.Context, itemID string, from, to int64) ([]Item, error) {
	return []Item{}, nil
}
func (p *SnmpProvider) CreateItem(ctx context.Context, item Item) (Item, error) { return item, nil }
func (p *SnmpProvider) UpdateItem(ctx context.Context, item Item) (Item, error) { return item, nil }
func (p *SnmpProvider) DeleteItem(ctx context.Context, itemID string) error     { return nil }
func (p *SnmpProvider) GetAlerts(ctx context.Context) ([]Alert, error)          { return []Alert{}, nil }
func (p *SnmpProvider) GetAlertsByHost(ctx context.Context, hostID string) ([]Alert, error) {
	return []Alert{}, nil
}
func (p *SnmpProvider) GetTriggers(ctx context.Context) ([]Trigger, error) { return []Trigger{}, nil }
func (p *SnmpProvider) GetTriggersByHost(ctx context.Context, hostID string) ([]Trigger, error) {
	return []Trigger{}, nil
}
func (p *SnmpProvider) GetTemplateidByName(ctx context.Context, name string) ([]string, error) {
	return []string{}, nil
}
func (p *SnmpProvider) GetHostGroups(ctx context.Context) ([]string, error) { return []string{}, nil }
func (p *SnmpProvider) GetHostGroupsDetails(ctx context.Context) ([]struct{ ID, Name string }, error) {
	return nil, nil
}
func (p *SnmpProvider) GetHostGroupByName(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (p *SnmpProvider) CreateHostGroup(ctx context.Context, name string) (string, error) {
	return "", nil
}
func (p *SnmpProvider) UpdateHostGroup(ctx context.Context, id, name string) error { return nil }
func (p *SnmpProvider) DeleteHostGroup(ctx context.Context, id string) error       { return nil }
func (p *SnmpProvider) Name() string                                               { return p.config.Name }
func (p *SnmpProvider) Type() MonitorType                                          { return MonitorSNMP }
