import re

def fix():
    with open('d:/Nagare_Project/nagare/backend/internal/service/trigger.go', 'r', encoding='utf-8') as f:
        text = f.read()

    # Remove trigger struct assignments in service layer
    text = re.sub(r'^\s*(Entity|AlertStatus|AlertGroupID|AlertMonitorID|AlertHostID|AlertItemID|AlertQuery|LogType|LogSeverity|LogQuery):[^\n]*\n', '', text, flags=re.MULTILINE)

    # Remove trigger properties being read
    text = re.sub(r'trigger\.(Entity|AlertStatus|AlertGroupID|AlertMonitorID|AlertHostID|AlertItemID|AlertQuery|LogType|LogSeverity|LogQuery)', '""', text)

    # Also clean up filter usages
    text = re.sub(r'filter\.(Entity|AlertStatus|AlertGroupID|AlertMonitorID|AlertHostID|AlertItemID|AlertQuery|LogType|LogSeverity|LogQuery)', 'None', text)

    with open('d:/Nagare_Project/nagare/backend/internal/service/trigger.go', 'w', encoding='utf-8') as f:
        f.write(text)

if __name__ == "__main__":
    fix()
