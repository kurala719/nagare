# Gmail Integration Debug & Setup Guide

## Overview

The 500 error when testing Gmail media indicates configuration issues or missing files. This guide will help you set up Gmail integration properly and debug any issues.

## Common Issues & Solutions

### Issue 1: Error 500 - Credentials File Not Found

**Symptoms:**
```
POST :5173/api/v1/media/3/test 500 Internal Server Error
```

**Root Cause:**
The `gmail_credentials.json` file is missing or the path is incorrect.

**Solution:**

1. **Get Your Google OAuth2 Credentials:**
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Create a new project
   - Enable Gmail API
   - Create OAuth 2.0 credentials (Desktop application)
   - Download the JSON file
   - Rename it to `gmail_credentials.json`
   - Place it in `backend/configs/` directory

2. **Update Configuration:**
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

3. **Verify File Exists:**
   ```bash
   # Windows
   dir backend\configs\gmail_credentials.json
   
   # Linux/Mac
   ls -la backend/configs/gmail_credentials.json
   ```

### Issue 2: Error 500 - Token File Not Found

**Symptoms:**
```
Error: gmail token not found or invalid at configs/gmail_token.json
```

**Root Cause:**
The OAuth2 token file doesn't exist or hasn't been initialized.

**Solution:**

1. **Create a Token File:**
   - You need to authorize the application first
   - Run this Go code to generate token:

   ```go
   package main
   
   import (
       "context"
       "encoding/json"
       "fmt"
       "log"
       "os"
       "github.com/spf13/viper"
       "golang.org/x/oauth2"
       "golang.org/x/oauth2/google"
       "google.golang.org/api/gmail/v1"
   )
   
   func main() {
       // Load config
       viper.SetConfigFile("configs/nagare_config.json")
       viper.ReadInConfig()
       
       // Read credentials
       b, err := os.ReadFile("configs/gmail_credentials.json")
       if err != nil {
           log.Fatal("Unable to read credentials file:", err)
       }
       
       config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
       if err != nil {
           log.Fatal("Unable to parse credentials:", err)
       }
       
       // Create auth URL
       authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
       fmt.Printf("Go to this URL to authorize:\n%v\n", authURL)
       
       // Get authorization code
       var code string
       fmt.Println("Enter the authorization code:")
       fmt.Scan(&code)
       
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

   Or use a helper script:
   ```bash
   # Create token with curl (after getting auth code)
   curl -X POST https://oauth2.googleapis.com/token \
     -d "client_id=YOUR_CLIENT_ID" \
     -d "client_secret=YOUR_CLIENT_SECRET" \
     -d "code=AUTHORIZATION_CODE" \
     -d "grant_type=authorization_code" \
     -d "redirect_uri=http://localhost"
   ```

2. **Token File Format:**
   ```json
   {
     "access_token": "ya29.a0AfH6SMB...",
     "token_type": "Bearer",
     "expiry": "2026-02-21T14:30:00.000000000Z",
     "refresh_token": "1//0gU..."
   }
   ```

### Issue 3: Error 500 - Invalid Configuration

**Symptoms:**
```
Error: gmail is disabled in configuration
```

**Solution:**

Make sure `gmail.enabled` is set to `true` in config:

```json
{
  "gmail": {
    "enabled": true,  // ← Must be true
    "credentials_file": "configs/gmail_credentials.json",
    "token_file": "configs/gmail_token.json",
    "from": "your-email@gmail.com"
  }
}
```

## Configuration Reference

### Required Configuration

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

### Configuration Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `enabled` | boolean | Yes | Enable Gmail integration |
| `credentials_file` | string | Yes | Path to OAuth2 credentials JSON |
| `token_file` | string | Yes | Path to OAuth2 token JSON |
| `from` | string | No | Sender email (defaults to config value) |

## Step-by-Step Setup Guide

### Step 1: Get Google OAuth2 Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create a new project:
   - Click "Select a Project"
   - Click "New Project"
   - Name: "Nagare Monitoring"
   - Click "Create"

3. Enable Gmail API:
   - Search for "Gmail API"
   - Click "Gmail API"
   - Click "Enable"

4. Create OAuth2 Credentials:
   - Click "Create Credentials"
   - Application type: "Desktop application"
   - Click "Create"
   - A JSON file will download
   - Save as `backend/configs/gmail_credentials.json`

### Step 2: Authorize the Application

Run the Go authorization script (see Issue 2 section) or use a web-based tool.

### Step 3: Update Configuration

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

### Step 4: Verify Files Exist

```bash
# Windows
dir backend\configs\gmail_credentials.json
dir backend\configs\gmail_token.json

# Linux/Mac
ls -la backend/configs/gmail_*.json
```

### Step 5: Test the Connection

1. Create a Gmail media via API:
```bash
POST /api/v1/media
{
  "name": "Gmail Alerts",
  "type": "gmail",
  "target": "recipient@example.com",
  "enabled": 1
}
```

2. Test the media:
```bash
POST /api/v1/media/3/test
```

3. Check for success response:
```json
{
  "status": "success",
  "message": "test message sent successfully"
}
```

## Debugging Steps

### Step 1: Check Configuration

```bash
# Linux/Mac
grep -A 5 '"gmail"' backend/configs/nagare_config.json

# Windows
findstr /A:2 "gmail" backend\configs\nagare_config.json
```

Expected output:
```json
"gmail": {
  "enabled": true,
  "credentials_file": "configs/gmail_credentials.json",
  "token_file": "configs/gmail_token.json",
  "from": "your-email@gmail.com"
}
```

### Step 2: Verify Files and Permissions

```bash
# Check files exist
ls -la backend/configs/gmail_*.json

# Verify readable
cat backend/configs/gmail_credentials.json | head -20
cat backend/configs/gmail_token.json | head -20
```

### Step 3: Check Logs

Look for error messages in logs:

```bash
# Check application logs
tail -f logs/system.log | grep -i "gmail\|test media"
```

Expected log line on success:
```
INFO    test media succeeded    media_id=3 media_type=gmail media_name="Gmail Alerts" target=recipient@example.com
```

Expected error lines to debug:
```
ERROR   test media failed    media_id=3 media_type=gmail error="unable to read credentials file: no such file or directory"
ERROR   test media failed    media_id=3 media_type=gmail error="gmail token not found or invalid at configs/gmail_token.json"
ERROR   test media failed    media_id=3 media_type=gmail error="gmail is disabled in configuration"
```

### Step 4: Validate Token File

Token must have valid access token:

```bash
# View token (Linux/Mac)
cat backend/configs/gmail_token.json | jq .

# Expected output:
# {
#   "access_token": "ya29.a0AfH6SMB...",
#   "token_type": "Bearer",
#   "expiry": "2026-02-21T14:30:00Z",
#   "refresh_token": "1//0gU..."
# }
```

### Step 5: Test Gmail API Directly

```bash
# Test with curl
curl -X POST https://www.googleapis.com/gmail/v1/users/me/messages/send \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "raw": "BASE64_ENCODED_EMAIL"
  }'
```

## Error Messages & Solutions

| Error | Cause | Solution |
|-------|-------|----------|
| `unable to read client secret file` | Credentials file missing | Place `gmail_credentials.json` in `configs/` |
| `gmail token not found or invalid` | Token file missing | Authorize app and generate token |
| `unable to parse credentials` | Invalid JSON format | Download fresh credentials from Google |
| `unable to send message` | Token expired | Refresh token (automatic via Google) |
| `gmail is disabled` | Config setting false | Set `enabled: true` in config |
| `target email address cannot be empty` | Missing recipient | Provide valid email in media target |

## Advanced Configuration

### Multiple Gmail Accounts

Create multiple media entries with different targets:

```bash
POST /api/v1/media
{
  "name": "Gmail Alerts - Admin",
  "type": "gmail",
  "target": "admin@company.com",
  "enabled": 1
}

POST /api/v1/media
{
  "name": "Gmail Alerts - Support",
  "type": "gmail",
  "target": "support@company.com",
  "enabled": 1
}
```

### Custom Sender Email

Set in configuration:

```json
{
  "gmail": {
    "from": "monitoring@company.com"
  }
}
```

## Security Notes

### Protect Your Files

```bash
# Set restrictive permissions (Linux/Mac)
chmod 600 backend/configs/gmail_credentials.json
chmod 600 backend/configs/gmail_token.json

# Don't commit to git
echo "configs/gmail_*.json" >> .gitignore
```

### Token Rotation

- Tokens expire after the time in `expiry` field
- Google SDK automatically refreshes with `refresh_token`
- No action needed - automatic refresh happens

### Scope Limitations

The app only requests `Gmail.GmailSendScope`:
- Can send emails only
- Cannot read emails
- Cannot delete emails
- Cannot modify labels

## Testing Checklist

- [ ] Gmail enabled in config (`enabled: true`)
- [ ] Credentials file exists at configured path
- [ ] Token file exists at configured path
- [ ] Token has valid `access_token` field
- [ ] Sender email configured (`from` field)
- [ ] Recipient email valid
- [ ] Network connectivity to Gmail API
- [ ] Firewall allows outbound HTTPS (port 443)

## Quick Fixes

If you get a 500 error, try these in order:

1. Check `enabled: true` in config
2. Verify files exist: `ls backend/configs/gmail_*.json`
3. Check file permissions: `chmod 600 backend/configs/gmail_*.json`
4. Verify token is valid: `cat backend/configs/gmail_token.json | grep access_token`
5. Check logs: `tail logs/system.log | grep gmail`
6. Restart server if config changed
7. Regenerate token if expired

## Support Resources

- [Google Gmail API Docs](https://developers.google.com/gmail/api)
- [OAuth2 Authorization](https://developers.google.com/identity/protocols/oauth2)
- [Gmail API Scopes](https://developers.google.com/gmail/api/auth/scopes)

## Example: Complete Setup

```bash
# 1. Create configs directory
mkdir -p backend/configs

# 2. Download credentials from Google Cloud Console
# (Place gmail_credentials.json in backend/configs/)

# 3. Authorize and create token
# (Use Go script or Google CLI tool)

# 4. Update config
cat > backend/configs/nagare_config.json << EOF
{
  "gmail": {
    "enabled": true,
    "credentials_file": "configs/gmail_credentials.json",
    "token_file": "configs/gmail_token.json",
    "from": "nagare@company.com"
  }
}
EOF

# 5. Test
curl -X POST http://localhost:8080/api/v1/media/3/test

# 6. Check logs
tail logs/system.log | grep -i gmail
```

---

**Need more help?** Check the logs with: `tail logs/system.log | grep "test media"` for detailed error messages.
