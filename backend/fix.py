import os
import re

def process_file(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    original_content = content
    # Common struct field renames
    content = re.sub(r'\.ExternalHostID', '.ExternalID', content)
    content = re.sub(r'ExternalHostID:', 'ExternalID:', content)
    content = re.sub(r'\.HID\b', '.HostID', content)
    content = re.sub(r'\bHID:', 'HostID:', content)
    content = re.sub(r'\.ExternalItemID', '.ExternalID', content)
    content = re.sub(r'ExternalItemID:', 'ExternalID:', content)
    
    # Alert struct virtual fields removal (comment out or remove)
    content = re.sub(r'\.HostName\b', '/* .HostName */', content)
    content = re.sub(r'HostName:', '// HostName:', content)
    content = re.sub(r'\.ItemName\b', '/* .ItemName */', content)
    content = re.sub(r'ItemName:', '// ItemName:', content)
    content = re.sub(r'\.AlarmName\b', '/* .AlarmName */', content)
    content = re.sub(r'AlarmName:', '// AlarmName:', content)
    
    # ActiveAvailable format
    content = re.sub(r'(existing|updated|h|group)\.ActiveAvailable', '/* \g<1>.ActiveAvailable */', content)
    content = re.sub(r'\bActiveAvailable:\s*[^,\n]*,', '// ActiveAvailable removed,', content)

    # ExternalSource format
    content = re.sub(r'(existing|updated|h|group|item)\.ExternalSource', '/* \g<1>.ExternalSource */', content)
    content = re.sub(r'\bExternalSource:\s*[^,\n]*,', '// ExternalSource removed,', content)

    # MonitorID on host, Item, group is a bit tricky, but mostly on host.
    content = re.sub(r'(existing|updated|newHost|host|h|globalHost|localHost)\.MonitorID', '/* \g<1>.MonitorID */', content)
    content = re.sub(r'\bMonitorID:\s*[^,\n]*,', '// MonitorID removed,', content)

    # GroupName, MonitorName on Host
    content = re.sub(r'h\.GroupName', '/* h.GroupName */', content)
    content = re.sub(r'h\.MonitorName', '/* h.MonitorName */', content)

    # Action TriggerID, Users
    content = re.sub(r'\.TriggerID\b', '.UserID', content) # Approximation
    content = re.sub(r'\bTriggerID:', 'UserID:', content)
    
    if content != original_content:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"Updated {file_path}")

def main():
    dirs_to_check = [
        'd:/Nagare_Project/nagare/backend/internal/service',
        'd:/Nagare_Project/nagare/backend/cmd/debug'
    ]
    for d in dirs_to_check:
        for root, dirs, files in os.walk(d):
            for file in files:
                if file.endswith('.go'):
                    process_file(os.path.join(root, file))

if __name__ == '__main__':
    main()
