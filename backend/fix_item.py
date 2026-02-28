import re

def fix_item():
    with open('d:/Nagare_Project/nagare/backend/internal/service/item.go', 'r', encoding='utf-8') as f:
        text = f.read()

    # duplicate ExternalID -> already fixed in another file or wait:
    # "internal\service\item.go:161:3: duplicate field name ExternalID in struct literal"
    # let's fix it:
    text = re.sub(r'^\s*ExternalID:\s*.*?,\n\s*ExternalID:\s*.*?,\n', '        ExternalID:        req.ExternalID,\n', text, flags=re.MULTILINE)

    # "internal\service\item.go:175:11: updated.ExternalSource undefined"
    text = re.sub(r'^\s*if req\.ExternalSource != "" \{\n\s*updated\.ExternalSource = req\.ExternalSource\n\s*\}\n', '', text, flags=re.MULTILINE)

    # replace host.MonitorID
    # For lines like `host, err := repository.GetHostByIDDAO(...)` followed by `host.MonitorID > 0`
    text = re.sub(
        r'if err == nil && host.MonitorID > 0 \{\n\s*if err := PushItemToMonitorServ\(host.MonitorID, host.ID, item.ID\); err == nil \{',
        r'if err == nil {\n\t\t\thostGroup, _ := repository.GetGroupByIDDAO(host.GroupID)\n\t\t\tif hostGroup.MonitorID > 0 {\n\t\t\t\tif err := PushItemToMonitorServ(hostGroup.MonitorID, host.ID, item.ID); err == nil {',
        text
    )

    text = re.sub(
        r'if err == nil && host.MonitorID > 0 \{\n\s*_ = PushItemToMonitorServ\(host.MonitorID, host.ID, id\)\n\s*\}',
        r'if err == nil {\n\t\t\thostGroup, _ := repository.GetGroupByIDDAO(host.GroupID)\n\t\t\tif hostGroup.MonitorID > 0 {\n\t\t\t\t_ = PushItemToMonitorServ(hostGroup.MonitorID, host.ID, id)\n\t\t\t}\n\t\t}',
        text
    )

    text = re.sub(
        r'if err == nil && host.MonitorID > 0 \{\n\s*_ = DeleteItemFromMonitorServ\(host.MonitorID, item.ExternalID\)\n\s*\}',
        r'if err == nil {\n\t\t\thostGroup, _ := repository.GetGroupByIDDAO(host.GroupID)\n\t\t\tif hostGroup.MonitorID > 0 {\n\t\t\t\t_ = DeleteItemFromMonitorServ(hostGroup.MonitorID, item.ExternalID)\n\t\t\t}\n\t\t}',
        text
    )

    # For functions getting MonitorID: GetItemsByHostIDFromMonitorServ, AddItemsByHostIDFromMonitorServ and others:
    text = re.sub(
        r'monitor, err := repository.GetMonitorByIDDAO\(host.MonitorID\)',
        r'hostGroup, gErr := repository.GetGroupByIDDAO(host.GroupID)\n\tif gErr != nil { return nil, gErr }\n\tmonitor, err := repository.GetMonitorByIDDAO(hostGroup.MonitorID)',
        text
    )
    
    text = re.sub(
        r'setMonitorStatusError\(host.MonitorID\)',
        r'setMonitorStatusError(hostGroup.MonitorID)',
        text
    )

    with open('d:/Nagare_Project/nagare/backend/internal/service/item.go', 'w', encoding='utf-8') as f:
        f.write(text)

if __name__ == "__main__":
    fix_item()
