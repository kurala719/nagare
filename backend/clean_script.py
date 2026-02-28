import os
import re

def clean_file(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    original = content
    # Clean up dirty struct initializations
    content = re.sub(r'^[ \t]*MonitorID:\s*existing\.MonitorID,\r?\n?', '', content, flags=re.MULTILINE)
    content = re.sub(r'^[ \t]*ActiveAvailable:\s*existing\.ActiveAvailable,\r?\n?', '', content, flags=re.MULTILINE)
    content = re.sub(r'^[ \t]*ExternalSource:\s*existing\.ExternalSource,\r?\n?', '', content, flags=re.MULTILINE)

    # In host.go AddHost/UpdateHost, newHost.MonitorID / updated.MonitorID should be h.MID or groupMonitorID.
    # We will just replace it with `h.MID` in the context of these functions
    if 'host.go' in file_path:
        content = re.sub(r'newHost\.MonitorID', 'h.MID', content)
        content = re.sub(r'updated\.MonitorID', 'h.MID', content)
        content = re.sub(r'host\.MonitorID', 'h.MID', content) # Approximation
        content = re.sub(r'globalHost\.MonitorID', 'mid', content) # in loop approximations

    if 'item.go' in file_path:
        # req.ItemID -> req.ExternalID
        content = re.sub(r'req\.ItemID', 'req.ExternalID', content)
        # item.ExternalItemID -> item.ExternalID
        content = re.sub(r'item\.ExternalItemID', 'item.ExternalID', content)
        # item.HID -> item.HostID
        content = re.sub(r'item\.HID', 'item.HostID', content)

    if content != original:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"Cleaned {file_path}")

def main():
    dirs_to_check = [
        'd:/Nagare_Project/nagare/backend/internal/service',
    ]
    for d in dirs_to_check:
        for root, dirs, files in os.walk(d):
            for file in files:
                if file.endswith('.go'):
                    clean_file(os.path.join(root, file))

if __name__ == '__main__':
    main()
