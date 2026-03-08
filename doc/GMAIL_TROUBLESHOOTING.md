# Gmail Media Testing - Troubleshooting Guide

## Error Flow Diagram

When you test Gmail media, here's the path the request takes:

```
1. Browser/API Client
   ↓ POST /api/v1/media/:id/test
2. TestMediaCtrl (API Controller)
   ↓ Calls TestMediaServ() 
3. TestMediaServ (Service Layer)
   ↓ Loads media from database
   ↓ Validates media exists
   ↓ Calls SendIMReply()
4. SendIMReply (IM Command)
   ↓ Validates inputs
   ↓ Gets media service
   ↓ Calls mediaSvc.SendMessage()
5. Medium Service Registry
   ↓ Looks up "gmail" provider
6. GmailProvider.SendMessage()
   ↓ Validates email/message
   ↓ Calls getGmailClient()
7. getGmailClient()
   ↓ Reads credentials file
   ↓ Parses JSON
   ↓ Reads token file
   ↓ Returns OAuth2 client
8. SendGmailServ()
   ↓ Creates Gmail service
   ↓ Sends message via Gmail API
   ↓ Returns result
9. Response back to client
```

## Every Possible Error Point

### Error Point 1: Media Not Found

**What Happens:**
```
HTTP 404 Not Found
{"error": "media not found"}
```

**Why:**
- The media ID you're testing doesn't exist in database
- Media was deleted
- Wrong ID in URL

**Fix:**
- Verify ID exists: `GET /api/v1/media/:id`
- Create new media if needed
- Double-check media ID in URL

---

### Error Point 2: Gmail Disabled in Configuration

**What Happens:**
```
HTTP 500 Internal Server Error
```

**Logs Show:**
```
ERROR test media failed media_type=gmail error="gmail is disabled in configuration (set gmail.enabled=true)"
```

**Why:**
- Configuration has `"enabled": false`
- Default configuration disables Gmail for security

**Fix:**
Edit `backend/configs/nagare_config.json`:

```json
{
  "gmail": {
    "enabled": true,
    "credentials_file": "configs/gmail_credentials.json",
    "token_file": "configs/gmail_token.json",
    "from": "your-email@gmail.com"
  }
}
```

Then **restart the server** for config changes to take effect.

---

### Error Point 3: Credentials File Not Found

**What Happens:**
```
HTTP 500 Internal Server Error
```

**Logs Show:**
```
ERROR test media failed media_type=gmail error="unable to read credentials file at configs/gmail_credentials.json: open configs/gmail_credentials.json: no such file or directory"
```

**Why:**
- File doesn't exist at path specified in config
- Wrong path in configuration
- File deleted or moved

**Fix:**

1. Download Google OAuth2 credentials:
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Select your project
   - Go to "Credentials"
   - Click "Create Credentials" → "OAuth 2.0 Client ID"
   - Choose "Desktop application"
   - Download JSON

2. Save file to exact path:
   ```bash
   # Copy to backend/configs/
   cp ~/Downloads/client_secret_*.json backend/configs/gmail_credentials.json
   ```

3. Verify it's readable:
   ```bash
   # Linux/Mac
   cat backend/configs/gmail_credentials.json | head -5

   # Windows PowerShell
   Get-Content backend\configs\gmail_credentials.json -Head 5
   ```

4. Restart server

---

### Error Point 4: Credentials File Invalid JSON

**What Happens:**
```
HTTP 500 Internal Server Error
```

**Logs Show:**
```
ERROR test media failed media_type=gmail error="unable to parse credentials file at configs/gmail_credentials.json: invalid character '}' looking for beginning of value"
```

**Why:**
- File is corrupted or not valid JSON
- File contains invalid characters
- Partially downloaded file

**Fix:**

1. Validate JSON syntax:
   ```bash
   # Linux/Mac
   python3 -m json.tool backend/configs/gmail_credentials.json

   # Windows PowerShell
   Get-Content backend\configs\gmail_credentials.json | ConvertFrom-Json
   ```

2. If validation fails:
   - Delete the file
   - Download fresh credentials from Google Cloud Console
   - Save to `backend/configs/gmail_credentials.json`

3. Restart server

---

### Error Point 5: Token File Not Found

**What Happens:**
```
HTTP 500 Internal Server Error
```

**Logs Show:**
```
ERROR test media failed media_type=gmail error="gmail token not found or invalid at configs/gmail_token.json: open configs/gmail_token.json: no such file or directory"
```

**Why:**
- OAuth2 token hasn't been generated
- File was deleted
- Wrong path in configuration

**Fix:**

1. Generate token using Go script:

Create `backend/generate_gmail_token.go`:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/gmail/v1"
)

func main() {
    // Read credentials
    b, err := os.ReadFile("configs/gmail_credentials.json")
    if err != nil {
        log.Fatal("Unable to read credentials file:", err)
    }

    config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
    if err != nil {
        log.Fatal("Unable to parse credentials:", err)
    }

    // Generate auth code
    authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
    fmt.Printf("Go to this URL to authorize:\n%v\n\n", authURL)

    var code string
    fmt.Print("Enter the authorization code: ")
    fmt.Scanln(&code)

    // Exchange for token
    tok, err := config.Exchange(context.Background(), code)
    if err != nil {
        log.Fatal("Unable to retrieve token:", err)
    }

    // Save token
    f, err := os.Create("configs/gmail_token.json")
    if err != nil {
        log.Fatal("Unable to create token file:", err)
    }
    defer f.Close()

    json.NewEncoder(f).Encode(tok)
    fmt.Println("Token saved successfully!")
}
```

2. Run it:
   ```bash
   cd backend
   go run generate_gmail_token.go
   # Follow the prompts to authorize
   ```

3. Verify token created:
   ```bash
   ls -la configs/gmail_token.json
   cat configs/gmail_token.json
   ```

4. Restart server

---

### Error Point 6: Token File Invalid JSON

**What Happens:**
```
HTTP 500 Internal Server Error
```

**Logs Show:**
```
ERROR test media failed media_type=gmail error="unable to decode token file at configs/gmail_token.json: unexpected end of JSON input"
```

**Why:**
- Token file is corrupted
- File contains invalid JSON
- Incomplete write

**Fix:**

1. Delete and regenerate token:
   ```bash
   rm backend/configs/gmail_token.json
   cd backend
   go run generate_gmail_token.go
   ```

2. Or manually create valid token:
   ```bash
   cat > backend/configs/gmail_token.json << 'EOF'
   {
     "access_token": "your_access_token_here",
     "token_type": "Bearer",
     "expiry": "2026-12-31T23:59:59Z",
     "refresh_token": "your_refresh_token_here"
   }
   EOF
   ```

3. Restart server

---

### Error Point 7: Token Has No Access Token

**What Happens:**
```
HTTP 500 Internal Server Error
```

**Logs Show:**
```
ERROR test media failed media_type=gmail error="token file at configs/gmail_token.json contains no access token"
```

**Why:**
- Token file is empty or missing fields
- Token file was manually edited and corrupted
- Authorization failed

**Fix:**

1. Regenerate token:
   ```bash
   rm backend/configs/gmail_token.json
   cd backend
   go run generate_gmail_token.go
   # Complete the authorization flow
   ```

2. Verify token format:
   ```bash
   # Linux/Mac
   cat backend/configs/gmail_token.json | jq .

   # Windows PowerShell
   (Get-Content backend\configs\gmail_token.json) | ConvertFrom-Json | Select access_token
   ```

3. File must have `access_token` field with value
4. Restart server

---

### Error Point 8: Target Email is Empty

**What Happens:**
```
HTTP 500 Internal Server Error
```

**Logs Show:**
```
ERROR test media failed media_type=gmail error="target email address cannot be empty"
```

**Why:**
- Media target field is not set
- Target is blank string
- Media created without target

**Fix:**

1. Update media with target:
   ```bash
   PUT /api/v1/media/:id
   {
     "target": "recipient@example.com"
   }
   ```

2. Or delete and recreate:
   ```bash
   POST /api/v1/media
   {
     "name": "Gmail Test",
     "type": "gmail",
     "target": "your-email@example.com",
     "enabled": 1
   }
   ```

3. Try test again

---

### Error Point 9: Gmail API Unreachable

**What Happens:**
```
HTTP 500 Internal Server Error
```

**Logs Show:**
```
ERROR test media failed media_type=gmail error="unable to send message: context deadline exceeded / network unreachable"
```

**Why:**
- No internet connectivity
- Firewall blocking port 443 to Google
- Proxy configuration issue
- Gmail API temporarily down

**Fix:**

1. Test connectivity:
   ```bash
   # Linux/Mac
   curl -I https://www.googleapis.com/gmail/v1/users/me/messages/send

   # Windows PowerShell
   Invoke-WebRequest -Uri "https://www.googleapis.com" -Method Head
   ```

2. Check firewall:
   - Allow outbound HTTPS (port 443)
   - Check proxy settings
   - Try from different network if possible

3. Check Gmail API status:
   - Visit [Google Cloud Status Dashboard](https://status.cloud.google.com)
   - Verify Gmail API service is operational

---

### Error Point 10: Token Expired

**What Happens:**
```
HTTP 500 Internal Server Error
```

**Logs Show:**
```
ERROR test media failed media_type=gmail error="unable to send message: invalid_grant"
```

**Why:**
- Token expired (happens after ~1 hour)
- Refresh token invalid
- User revoked authorization

**Fix:**

1. Automatic refresh (should happen automatically):
   - System will automatically refresh using refresh_token
   - May take a few seconds

2. If refresh fails, regenerate token:
   ```bash
   rm backend/configs/gmail_token.json
   cd backend
   go run generate_gmail_token.go
   ```

3. Check that token file has `refresh_token` field

---

### Error Point 11: User Revoked Authorization

**What Happens:**
```
HTTP 500 Internal Server Error
```

**Logs Show:**
```
ERROR test media failed media_type=gmail error="unable to send message: invalid_grant"
```

**Why:**
- User removed app from authorized apps in Google account
- Happened at [https://myaccount.google.com/permissions](https://myaccount.google.com/permissions)
- Token refresh failed

**Fix:**

1. Remove app authorization:
   - Go to [https://myaccount.google.com/permissions](https://myaccount.google.com/permissions)
   - Find "Nagare" or Google Cloud app
   - Click "Remove access"

2. Regenerate authorization:
   ```bash
   rm backend/configs/gmail_token.json
   cd backend
   go run generate_gmail_token.go
   # This will re-authorize the app
   ```

3. Restart server

---

### Error Point 12: Invalid Recipient Email

**What Happens:**
```
HTTP 500 Internal Server Error
```

**Logs Show:**
```
ERROR test media failed media_type=gmail error="unable to send message: invalid email format"
```

**Why:**
- Target email has invalid format (e.g., "test@" or "@example.com")
- Email contains invalid characters
- Typo in email address

**Fix:**

1. Verify email is valid:
   - Should match pattern: `user@domain.com`
   - No spaces in email
   - Valid domain

2. Update media:
   ```bash
   PUT /api/v1/media/:id
   {
     "target": "valid-email@example.com"
   }
   ```

3. Try test again

---

## Error Resolution Flowchart

```
GET 500 Error on /api/v1/media/:id/test
    ├─ Check logs: tail logs/system.log | grep gmail
    │
    ├─ Error: "gmail is disabled"
    │  └─ Set enabled: true in config → Restart → Retry
    │
    ├─ Error: "credentials file" or "no such file"
    │  └─ Download credentials → Save to backend/configs/ → Restart → Retry
    │
    ├─ Error: "parse credentials" or "invalid character"
    │  └─ Validate JSON → Re-download if invalid → Restart → Retry
    │
    ├─ Error: "token not found"
    │  └─ Run generate_gmail_token.go → Authorize → Restart → Retry
    │
    ├─ Error: "token file invalid" or "no access_token"
    │  └─ Delete token file → Run generate_gmail_token.go → Restart → Retry
    │
    ├─ Error: "email address cannot be empty"
    │  └─ Set media target → PUT /api/v1/media/:id → Retry
    │
    ├─ Error: "network unreachable"
    │  └─ Check internet → Check firewall → Retry
    │
    ├─ Error: "invalid_grant"
    │  └─ Regenerate token → Restart → Retry
    │
    └─ Success: 200 OK
       └─ Email sent! Check recipient mailbox
```

---

## Testing Checklist

Before testing Gmail media, verify:

- [ ] `backend/configs/nagare_config.json` exists
- [ ] `"gmail": {"enabled": true}` in config
- [ ] `backend/configs/gmail_credentials.json` exists and is readable
- [ ] Credentials file contains valid JSON
- [ ] `backend/configs/gmail_token.json` exists and is readable
- [ ] Token file contains valid JSON with `access_token` field
- [ ] Media has valid target email address
- [ ] Server restarted after config changes
- [ ] Internet connectivity available
- [ ] Gmail API enabled in Google Cloud project

## Quick Debug Commands

```bash
# Check config
cat backend/configs/nagare_config.json | grep -A 5 gmail

# Validate credentials JSON
python3 -m json.tool backend/configs/gmail_credentials.json

# Validate token JSON
python3 -m json.tool backend/configs/gmail_token.json

# Check token has access_token
cat backend/configs/gmail_token.json | grep access_token

# View recent logs
tail -50 logs/system.log | grep -i "gmail\|test media"

# Check file permissions
ls -la backend/configs/gmail_*.json

# Test connectivity to Gmail API
curl -I https://www.googleapis.com/gmail/v1/users/me/messages/send
```

---

**When all else fails:**
1. Delete both token and credentials files
2. Download fresh credentials from Google Cloud
3. Regenerate token using the Go script  
4. Restart server
5. Try test again
