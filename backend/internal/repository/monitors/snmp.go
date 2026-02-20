package monitors

import (
	"context"
	"fmt"
	"time"

	"github.com/gosnmp/gosnmp"
)

type SnmpProvider struct {
	config Config
}

func NewSnmpProvider(cfg Config) (Provider, error) {
	return &SnmpProvider{config: cfg}, nil
}

func (p *SnmpProvider) Authenticate(ctx context.Context) error {
	// For SNMP, authentication is per-host. 
	// As a monitor-level check, we could try to ping the "URL" if it's an IP,
	// but usually the Monitor URL for SNMP might just be "localhost" or a seed device.
	return nil
}

func (p *SnmpProvider) GetAuthToken() string { return "" }
func (p *SnmpProvider) SetAuthToken(token string) {}

func (p *SnmpProvider) GetHosts(ctx context.Context) ([]Host, error) {
	// SNMP provider doesn't "discover" hosts from a central server like Zabbix.
	// It relies on hosts defined in Nagare.
	return []Host{}, nil
}

func (p *SnmpProvider) GetHostsByGroupID(ctx context.Context, groupID string) ([]Host, error) {
	return []Host{}, nil
}

func (p *SnmpProvider) GetHostByName(ctx context.Context, name string) (*Host, error) {
	return nil, nil
}

func (p *SnmpProvider) GetHostByID(ctx context.Context, hostID string) (*Host, error) {
	return nil, nil
}

func (p *SnmpProvider) CreateHost(ctx context.Context, host Host) (Host, error) {
	return host, nil
}

func (p *SnmpProvider) UpdateHost(ctx context.Context, host Host) (Host, error) {
	return host, nil
}

func (p *SnmpProvider) DeleteHost(ctx context.Context, hostID string) error {
	return nil
}

// GetItems performs SNMP GET for a specific host.
// In SNMP context, hostID is actually the IP address or a reference we can use.
// We expect SNMP metadata (version, community) to be passed or available.
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
	if len(oid) > 0 && oid[0] == '.' {
		return oid[1:]
	}
	return oid
}

func (p *SnmpProvider) GetItems(ctx context.Context, hostID string) ([]Item, error) {
	config, ok := ctx.Value("snmp_config").(SnmpConfig)
	if !ok {
		config = SnmpConfig{
			Community: "public",
			Version:   "v2c",
			Port:      161,
		}
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

	fmt.Printf("SNMP Debug: Starting poll for %s (Version: %s, Community: %s, Port: %d)\n", hostID, config.Version, config.Community, config.Port)

	// OID Map: Normalizing keys to avoid dot mismatches
	oidNames := make(map[string]string)
	addOID := func(oid, name string) {
		oidNames[normalizeOID(oid)] = name
	}

	addOID(".1.3.6.1.2.1.1.3.0", "sysUpTime")
	addOID(".1.3.6.1.2.1.1.5.0", "sysName")
	
	if len(config.CustomOIDs) == 0 {
		// Huawei entities
		addOID(".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.5", "hwEntityCpuUsage")
		addOID(".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.7", "hwEntityMemUsage")
		addOID(".1.3.6.1.4.1.2011.5.25.31.1.1.1.1.11", "hwEntityTemperature")
	} else {
		for k, v := range config.CustomOIDs {
			addOID(k, v)
		}
	}

	gs := &gosnmp.GoSNMP{
		Target:    hostID,
		Port:      uint16(config.Port),
		Timeout:   time.Duration(10) * time.Second, // Increased timeout for reliability
		Retries:   1,
		MaxOids:   gosnmp.MaxOids,
	}

	switch config.Version {
	case "v1":
		gs.Version = gosnmp.Version1
		gs.Community = config.Community
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

		sp := &gosnmp.UsmSecurityParameters{
			UserName: config.V3User,
		}

		if level != "NoAuthNoPriv" {
			sp.AuthenticationPassphrase = config.V3AuthPass
			switch config.V3AuthProtocol {
			case "MD5": sp.AuthenticationProtocol = gosnmp.MD5
			case "SHA": sp.AuthenticationProtocol = gosnmp.SHA
			case "SHA224": sp.AuthenticationProtocol = gosnmp.SHA224
			case "SHA256": sp.AuthenticationProtocol = gosnmp.SHA256
			case "SHA384": sp.AuthenticationProtocol = gosnmp.SHA384
			case "SHA512": sp.AuthenticationProtocol = gosnmp.SHA512
			default: sp.AuthenticationProtocol = gosnmp.SHA
			}
		}

		if level == "AuthPriv" {
			sp.PrivacyPassphrase = config.V3PrivPass
			switch config.V3PrivProtocol {
			case "DES": sp.PrivacyProtocol = gosnmp.DES
			case "AES", "AES128": sp.PrivacyProtocol = gosnmp.AES
			case "AES192": sp.PrivacyProtocol = gosnmp.AES192
			case "AES256": sp.PrivacyProtocol = gosnmp.AES256
			default: sp.PrivacyProtocol = gosnmp.AES
			}
		}
		gs.SecurityParameters = sp
	default: // v2c
		gs.Version = gosnmp.Version2c
		gs.Community = config.Community
	}

	err := gs.Connect()
	if err != nil {
		fmt.Printf("SNMP Debug: Connect failed: %v\n", err)
		return nil, fmt.Errorf("SNMP connect failed to %s:%d: %w", hostID, config.Port, err)
	}
	defer gs.Conn.Close()

	var targetOids []string
	for rawOid := range oidNames {
		// GoSNMP expects OIDs starting with dot for Get/Walk
		targetOids = append(targetOids, "."+rawOid)
	}

	fmt.Printf("SNMP Debug: Sending GET request for %d OIDs...\n", len(targetOids))
	result, err := gs.Get(targetOids)
	
	// Fallback to Walk if GET fails (some devices don't like bulk GET)
	if err != nil {
		fmt.Printf("SNMP Debug: GET failed (%v), attempting WALK fallback...\n", err)
		var items []Item
		for _, oid := range targetOids {
			fmt.Printf("SNMP Debug: Walking %s...\n", oid)
			err := gs.Walk(oid, func(variable gosnmp.SnmpPDU) error {
				normalized := normalizeOID(variable.Name)
				name := oidNames[normalized]
				if name == "" {
					name = normalized
				}
				
				items = append(items, Item{
					ID:        normalized,
					Name:      name,
					Key:       normalized,
					Value:     formatSnmpValue(variable),
					Timestamp: time.Now().Unix(),
					Status:    "1",
				})
				return nil
			})
			if err != nil {
				fmt.Printf("SNMP Debug: WALK failed for %s: %v\n", oid, err)
			}
		}
		if len(items) > 0 {
			fmt.Printf("SNMP Debug: WALK success, retrieved %d items.\n", len(items))
			return items, nil
		}
		return nil, fmt.Errorf("SNMP poll failed for %s: GET and WALK both failed", hostID)
	}

	fmt.Printf("SNMP Debug: GET success, processing %d results.\n", len(result.Variables))
	var items []Item
	for _, variable := range result.Variables {
		normalized := normalizeOID(variable.Name)
		name := oidNames[normalized]
		if name == "" {
			name = normalized
		}

		val := formatSnmpValue(variable)
		fmt.Printf("SNMP Debug: Raw Value for %s (%s) is [%s], Type: %v\n", variable.Name, name, val, variable.Type)
		
		// Huawei Smart-Fix: If CPU/Mem usage is N/A, try to walk the branch to find the instance
		if (val == "N/A" || val == "" || val == "0") && (name == "hwEntityCpuUsage" || name == "hwEntityMemUsage" || name == "hwEntityTemperature") {
			fmt.Printf("SNMP Debug: Huawei OID %s is [%s], triggering board discovery Smart-Walk on .%s...\n", name, val, normalized)
			_ = gs.Walk("."+normalized, func(v gosnmp.SnmpPDU) error {
				walkVal := formatSnmpValue(v)
				if walkVal != "N/A" && walkVal != "0" && (val == "N/A" || val == "" || val == "0") {
					val = walkVal
					fmt.Printf("SNMP Debug: Found valid instance for %s at %s: %s\n", name, v.Name, val)
				}
				return nil
			})
		}

		fmt.Printf("SNMP Debug: Final Result for %s (%s) = %s\n", variable.Name, name, val)

		items = append(items, Item{
			ID:        normalized,
			Name:      name,
			Key:       normalized,
			Value:     val,
			Timestamp: time.Now().Unix(),
			Status:    "1",
		})
	}

	return items, nil
}

func formatSnmpValue(variable gosnmp.SnmpPDU) string {
	switch variable.Type {
	case gosnmp.OctetString:
		return string(variable.Value.([]byte))
	case gosnmp.Integer, gosnmp.Counter32, gosnmp.Gauge32, gosnmp.Counter64, gosnmp.TimeTicks:
		return fmt.Sprintf("%v", gosnmp.ToBigInt(variable.Value))
	case gosnmp.NoSuchObject, gosnmp.NoSuchInstance:
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

func (p *SnmpProvider) CreateItem(ctx context.Context, item Item) (Item, error) {
	return item, nil
}

func (p *SnmpProvider) UpdateItem(ctx context.Context, item Item) (Item, error) {
	return item, nil
}

func (p *SnmpProvider) DeleteItem(ctx context.Context, itemID string) error {
	return nil
}

func (p *SnmpProvider) GetAlerts(ctx context.Context) ([]Alert, error) {
	return []Alert{}, nil
}

func (p *SnmpProvider) GetAlertsByHost(ctx context.Context, hostID string) ([]Alert, error) {
	return []Alert{}, nil
}

func (p *SnmpProvider) GetTriggers(ctx context.Context) ([]Trigger, error) {
	return []Trigger{}, nil
}

func (p *SnmpProvider) GetTriggersByHost(ctx context.Context, hostID string) ([]Trigger, error) {
	return []Trigger{}, nil
}

func (p *SnmpProvider) GetTemplateidByName(ctx context.Context, name string) ([]string, error) {
	return []string{}, nil
}

func (p *SnmpProvider) GetHostGroups(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (p *SnmpProvider) GetHostGroupsDetails(ctx context.Context) ([]struct{ ID, Name string }, error) {
	return nil, nil
}

func (p *SnmpProvider) GetHostGroupByName(ctx context.Context, name string) (string, error) {
	return "", nil
}

func (p *SnmpProvider) CreateHostGroup(ctx context.Context, name string) (string, error) {
	return "", nil
}

func (p *SnmpProvider) UpdateHostGroup(ctx context.Context, id, name string) error {
	return nil
}

func (p *SnmpProvider) DeleteHostGroup(ctx context.Context, id string) error {
	return nil
}

func (p *SnmpProvider) CreateMediaType(ctx context.Context, name string, script string, params map[string]string) error {
	return nil
}

func (p *SnmpProvider) GetMediaTypeIDByName(ctx context.Context, name string) (string, error) {
	return "", nil
}

func (p *SnmpProvider) GetUserIDByUsername(ctx context.Context, username string) (string, error) {
	return "", nil
}

func (p *SnmpProvider) EnsureUserMedia(ctx context.Context, userID string, mediaTypeID string, sendTo string) error {
	return nil
}

func (p *SnmpProvider) EnsureActionWithMedia(ctx context.Context, name string, userID string, mediaTypeID string) error {
	return nil
}

func (p *SnmpProvider) Name() string {
	return p.config.Name
}

func (p *SnmpProvider) Type() MonitorType {
	return MonitorSNMP
}
