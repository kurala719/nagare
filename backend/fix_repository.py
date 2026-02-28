import re

def fix():
    # 1. Fix repository/trigger.go
    with open('d:/Nagare_Project/nagare/backend/internal/repository/trigger.go', 'r', encoding='utf-8') as f:
        text = f.read()

    # Removes assignments in trigger DAO update
    text = re.sub(r'(trigger\.(Entity|AlertStatus|AlertGroupID|AlertMonitorID|AlertHostID|AlertItemID|AlertQuery|LogType|LogSeverity|LogQuery)\s*=[^\n]*\n)', '', text)
    
    # Removes check for filter fields
    text = re.sub(r'if filter\.(Entity|AlertStatus|AlertGroupID|AlertMonitorID|AlertHostID|AlertItemID|AlertQuery|LogType|LogSeverity|LogQuery) != nil[^{]*{[^}]*}', '', text)

    with open('d:/Nagare_Project/nagare/backend/internal/repository/trigger.go', 'w', encoding='utf-8') as f:
        f.write(text)

    # 2. Fix api/trigger.go
    with open('d:/Nagare_Project/nagare/backend/internal/api/trigger.go', 'r', encoding='utf-8') as f:
        text = f.read()

    # Remove field passing into filter
    text = re.sub(r'^\s*(Entity|AlertStatus|AlertGroupID|AlertMonitorID|AlertHostID|AlertItemID|AlertQuery|LogType|LogSeverity|LogQuery):[^\n]*\n', '', text, flags=re.MULTILINE)

    with open('d:/Nagare_Project/nagare/backend/internal/api/trigger.go', 'w', encoding='utf-8') as f:
        f.write(text)


    # 3. Fix log files
    with open('d:/Nagare_Project/nagare/backend/internal/repository/log.go', 'r', encoding='utf-8') as f:
        text = f.read()
    text = re.sub(r'if filter\.Type != ""[^{]*{[^}]*}', '', text)
    with open('d:/Nagare_Project/nagare/backend/internal/repository/log.go', 'w', encoding='utf-8') as f:
        f.write(text)

    with open('d:/Nagare_Project/nagare/backend/internal/service/log.go', 'r', encoding='utf-8') as f:
        text = f.read()
    text = re.sub(r'^\s*Type:\s*[^\n]*\n', '', text, flags=re.MULTILINE)
    text = re.sub(r'entry\.Type\b', '""', text)
    with open('d:/Nagare_Project/nagare/backend/internal/service/log.go', 'w', encoding='utf-8') as f:
        f.write(text)
        
    with open('d:/Nagare_Project/nagare/backend/internal/api/log.go', 'r', encoding='utf-8') as f:
        text = f.read()
    text = re.sub(r'^\s*Type:\s*[^\n]*\n', '', text, flags=re.MULTILINE)
    with open('d:/Nagare_Project/nagare/backend/internal/api/log.go', 'w', encoding='utf-8') as f:
        f.write(text)

if __name__ == "__main__":
    fix()
