import os

filepath = r'd:\\Nagare_Project\\nagare\\backend\\internal\\service\\host.go'
with open(filepath, 'r', encoding='utf-8') as f:
    content = f.read()

# Fix import
if "nagare/internal/database" not in content:
    content = content.replace('"nagare/internal/model"', '"nagare/internal/database"\n\t"nagare/internal/model"')

# Fix GetHostsFromMonitorServ
old_get_hosts = """func GetHostsFromMonitorServ(mid uint) ([]HostResp, error) {
	hostsRecord, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
	if err != nil {
		return nil, err
	}
	var hosts []HostResp
	for _, hr := range hostsRecord {
		hosts = append(hosts, hostToResp(hr))
	}
	return hosts, nil
}"""
new_get_hosts = """func GetHostsFromMonitorServ(mid uint) ([]HostResp, error) {
	hostsRecord, err := repository.SearchHostsDAO(model.HostFilter{MID: &mid})
	if err != nil {
		return nil, err
	}
	var hosts []HostResp
	for _, hr := range hostsRecord {
		hosts = append(hosts, hostToResp(hr))
	}
	return enrichHostResps(hosts), nil
}"""
content = content.replace(old_get_hosts, new_get_hosts)

# Fix GetFailedHostsServ
old_get_failed = """func GetFailedHostsServ() ([]HostResp, error) {
	status := 2 // Error state
	filter := model.HostFilter{Status: &status}
	hostsRecord, err := repository.SearchHostsDAO(filter)
	if err != nil {
		return nil, err
	}
	var hosts []HostResp
	for _, hr := range hostsRecord {
		hosts = append(hosts, hostToResp(hr))
	}
	return hosts, nil
}"""
new_get_failed = """func GetFailedHostsServ() ([]HostResp, error) {
	status := 2 // Error state
	filter := model.HostFilter{Status: &status}
	hostsRecord, err := repository.SearchHostsDAO(filter)
	if err != nil {
		return nil, err
	}
	var hosts []HostResp
	for _, hr := range hostsRecord {
		hosts = append(hosts, hostToResp(hr))
	}
	return enrichHostResps(hosts), nil
}"""
content = content.replace(old_get_failed, new_get_failed)

with open(filepath, 'w', encoding='utf-8') as f:
    f.write(content)

print("Python script executed.")
