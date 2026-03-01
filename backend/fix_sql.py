import os

def replace_in_file(filepath, replacements):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    for old, new in replacements:
        content = content.replace(old, new)
        
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

# Fix items queries in repository
item_go = r'd:\\Nagare_Project\\nagare\\backend\\internal\\repository\\item.go'
replace_in_file(item_go, [
    ('items.itemid', 'items.external_id'),
    ('items.hostid', 'items.host_id'),  # Map to HostID DB col
    ('itemid LIKE', 'external_id LIKE'),
    ('itemid =', 'external_id ='),
    ('hostid LIKE', 'host_id LIKE'),
    ('hostid =', 'host_id ='),
    ('case "itemid":', 'case "external_id":'),
    ('case "hostid":', 'case "host_id":'),
    ('"hid = ?', '"host_id = ?'),
    (' hid,', ' hid,'),
])

# Fix hosts queries in repository
host_go = r'd:\\Nagare_Project\\nagare\\backend\\internal\\repository\\host.go'
replace_in_file(host_go, [
    ('hosts.hostid', 'hosts.external_id'),
    ('hostid LIKE', 'external_id LIKE'),
    ('case "hostid":', 'case "external_id":'),
])

# Fix alerts queries in repository
alert_go = r'd:\\Nagare_Project\\nagare\\backend\\internal\\repository\\alert.go'
replace_in_file(alert_go, [
    ('hosts.hostid', 'hosts.external_id'),
    ('items.itemid', 'items.external_id'),
])

print("SQL literals replaced successfully.")
