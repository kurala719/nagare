import re

def fix_item():
    with open('d:/Nagare_Project/nagare/backend/internal/service/item.go', 'r', encoding='utf-8') as f:
        text = f.read()

    # duplicate ExternalID
    text = re.sub(r'(?m)^\s*ExternalID:\s*req\.ExternalID,\n\s*ExternalID:\s*req\.ExternalID,\n', '        ExternalID:        req.ExternalID,\n', text)

    # undefined updated.ExternalSource
    text = re.sub(r'(?m)^\s*if req\.ExternalSource != "" \{\n\s*updated\.ExternalSource = req\.ExternalSource\n\s*\}\n', '', text)

    # Replace specific function calls one by one safely, checking how they appear
    
    # PushItemToMonitorServ
    text = text.replace(
        'if err == nil && host.MonitorID > 0 {\n\t\t\tif err := PushItemToMonitorServ(host.MonitorID, host.ID, item.ID); err == nil {',
        'if err == nil {\n\t\t\thostGroup, _ := repository.GetGroupByIDDAO(host.GroupID)\n\t\t\tif hostGroup.MonitorID > 0 {\n\t\t\t\tif err := PushItemToMonitorServ(hostGroup.MonitorID, host.ID, item.ID); err == nil {'
    )
    
    # PushToMonitor auto-push
    text = text.replace(
        'if err == nil && host.MonitorID > 0 {\n\t\t\t_ = PushItemToMonitorServ(host.MonitorID, host.ID, id)\n\t\t}',
        'if err == nil {\n\t\t\thostGroup, _ := repository.GetGroupByIDDAO(host.GroupID)\n\t\t\tif hostGroup.MonitorID > 0 {\n\t\t\t\t_ = PushItemToMonitorServ(hostGroup.MonitorID, host.ID, id)\n\t\t\t}\n\t\t}'
    )
    
    # DeleteItemFromMonitorServ
    text = text.replace(
        'if err == nil && host.MonitorID > 0 {\n\t\t\t_ = DeleteItemFromMonitorServ(host.MonitorID, item.ExternalID)\n\t\t}',
        'if err == nil {\n\t\t\thostGroup, _ := repository.GetGroupByIDDAO(host.GroupID)\n\t\t\tif hostGroup.MonitorID > 0 {\n\t\t\t\t_ = DeleteItemFromMonitorServ(hostGroup.MonitorID, item.ExternalID)\n\t\t\t}\n\t\t}'
    )

    # All instances of `monitor, err := repository.GetMonitorByIDDAO(host.MonitorID)`
    text = re.sub(
        r'monitor, err := repository\.GetMonitorByIDDAO\(host\.MonitorID\)',
        r'hostGroup, gErr := repository.GetGroupByIDDAO(host.GroupID)\n\tif gErr != nil { return nil, fmt.Errorf("failed to get host group: %w", gErr) }\n\tmonitor, err := repository.GetMonitorByIDDAO(hostGroup.MonitorID)',
        text
    )
    
    # The setMonitorStatusError takes hostGroup.MonitorID instead.
    text = re.sub(
        r'setMonitorStatusError\(host\.MonitorID\)',
        r'setMonitorStatusError(hostGroup.MonitorID)',
        text
    )

    with open('d:/Nagare_Project/nagare/backend/internal/service/item.go', 'w', encoding='utf-8') as f:
        f.write(text)

if __name__ == "__main__":
    fix_item()
