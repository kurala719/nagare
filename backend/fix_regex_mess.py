import os
import re

def fix_files():
    # Fix alert.go
    alert_path = 'd:/Nagare_Project/nagare/backend/internal/service/alert.go'
    with open(alert_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    content = content.replace('req/* .HostName */', 'req.HostName')
    content = content.replace('req/* .ItemName */', 'req.ItemName')
    
    with open(alert_path, 'w', encoding='utf-8') as f:
        f.write(content)

    # Fix chat.go
    chat_path = 'd:/Nagare_Project/nagare/backend/internal/service/chat.go'
    with open(chat_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    content = content.replace('sanitizeSensitiveText(metric/* .HostName */)', 'fmt.Sprintf("%d", metric.HostID)')
    content = content.replace('sanitizeSensitiveText(metric/* .ItemName */)', 'fmt.Sprintf("%d", metric.ItemID)')
    
    with open(chat_path, 'w', encoding='utf-8') as f:
        f.write(content)


    # Fix group.go
    group_path = 'd:/Nagare_Project/nagare/backend/internal/service/group.go'
    with open(group_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Remove updated.ExternalSource check
    content = re.sub(r'^[ \t]*if req\.ExternalSource != nil \{\r?\n[ \t]*updated\.ExternalSource = \*req\.ExternalSource\r?\n[ \t]*\}\r?\n?', '', content, flags=re.MULTILINE)

    with open(group_path, 'w', encoding='utf-8') as f:
        f.write(content)

if __name__ == '__main__':
    fix_files()
