import sys

def fix_globals():
    with open('d:/Nagare_Project/nagare/backend/internal/service/item.go', 'r', encoding='utf-8') as f:
        text = f.read()

    # Replacing all variations of ExternalItemID and ExternalHostID
    text = text.replace('.ExternalItemID', '.ExternalID')
    text = text.replace('.ExternalHostID', '.ExternalID')
    text = text.replace('ExternalHostID:', 'ExternalID:')
    text = text.replace('ExternalItemID:', 'ExternalID:')
    
    # Also fix item.HID -> item.HostID (but not req.HID)
    text = text.replace('item.HID', 'item.HostID')
    text = text.replace('localItem.HID', 'localItem.HostID')

    with open('d:/Nagare_Project/nagare/backend/internal/service/item.go', 'w', encoding='utf-8') as f:
        f.write(text)

if __name__ == "__main__":
    fix_globals()
