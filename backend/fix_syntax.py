import os
import re

def fix_file(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    original = content
    # Revert all comment-out of MonitorID
    content = re.sub(r'/\*\s*(.*?)\.MonitorID\s*\*/', r'\g<1>.MonitorID', content)
    content = re.sub(r'// MonitorID removed,', r'MonitorID: existing.MonitorID,', content)

    # Revert all comment-out of ExternalSource
    content = re.sub(r'/\*\s*(.*?)\.ExternalSource\s*\*/', r'\g<1>.ExternalSource', content)
    content = re.sub(r'// ExternalSource removed,', r'ExternalSource: existing.ExternalSource,', content)

    # Revert all comment-out of ActiveAvailable
    content = re.sub(r'/\*\s*(.*?)\.ActiveAvailable\s*\*/', r'\g<1>.ActiveAvailable', content)
    content = re.sub(r'// ActiveAvailable removed,', r'ActiveAvailable: existing.ActiveAvailable,', content)

    if content != original:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"Fixed {file_path}")

def main():
    dirs_to_check = [
        'd:/Nagare_Project/nagare/backend/internal/service',
        'd:/Nagare_Project/nagare/backend/cmd/debug'
    ]
    for d in dirs_to_check:
        for root, dirs, files in os.walk(d):
            for file in files:
                if file.endswith('.go'):
                    fix_file(os.path.join(root, file))

if __name__ == '__main__':
    main()
