# Gmail Integration - Quick Start Guide

⏱️ **Time to complete: 30 minutes**

## Before You Start

You need:
- ✅ Google Cloud Project with Gmail API enabled
- ✅ OAuth2 credentials JSON file downloaded
- ✅ Backend source code with Gmail enhancements
- ✅ Running Nagare backend server

---

## Step 1: Download Credentials (5 minutes)

### 1.1 Go to Google Cloud Console
```
https://console.cloud.google.com
```

### 1.2 Create Project (if needed)
- Click "Select a Project"
- Click "New Project"
- Name: "Nagare"
- Click "Create"

### 1.3 Enable Gmail API
- Search for "Gmail API"
- Click "Gmail API"
- Click "Enable"

### 1.4 Create OAuth2 Credentials
- Go to "Credentials"
- Click "Create Credentials"
- Choose "OAuth 2.0 Client ID"
- Application type: "Desktop application"
- Click "Create"
- A JSON file downloads automatically

### 1.5 Save Credentials File
```bash
# Copy downloaded file to backend/configs/
cp ~/Downloads/client_secret_*.json backend/configs/gmail_credentials.json
```

**Verify:**
```bash
ls -la backend/configs/gmail_credentials.json
# Should show file exists and is readable
```

---

## Step 2: Generate Token (5 minutes)

### 2.1 Create Authorization Script

Save this as `backend/generate_gmail_token.go`:

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

    // Generate auth URL
    authURL := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
    fmt.Printf("Go to this URL to authorize:\n%v\n\n", authURL)

    // Get authorization code
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
    fmt.Println("Token saved successfully to configs/gmail_token.json!")
}
```

### 2.2 Run Authorization Script

```bash
cd backend
go run generate_gmail_token.go
```

### 2.3 Follow the Prompts

**Output:**
```
Go to this URL to authorize:
https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=...

Enter the authorization code:
```

1. **Click the link** with your Google account
2. **Click "Allow"** to authorize the app
3. **Copy the code** from the browser
4. **Paste it** into terminal when prompted
5. **Wait** for "Token saved successfully"

**Verify Token Created:**
```bash
ls -la backend/configs/gmail_token.json
# Should show file exists
```

---

## Step 3: Update Configuration (2 minutes)

Edit `backend/configs/nagare_config.json`:

Find the Gmail section:
```json
"gmail": {
  "enabled": false,
  "credentials_file": "configs/gmail_credentials.json",
  "token_file": "configs/gmail_token.json",
  "from": "your-email@gmail.com"
}
```

Change to:
```json
"gmail": {
  "enabled": true,
  "credentials_file": "configs/gmail_credentials.json",
  "token_file": "configs/gmail_token.json",
  "from": "your-email@gmail.com"
}
```

**Key change:** `"enabled": true`

---

## Step 4: Rebuild Backend (3 minutes)

### 4.1 Navigate to Backend
```bash
cd backend
```

### 4.2 Build
```bash
go build -o bin/nagare-web-server ./cmd/server
```

**Expected output:**
```
# No errors shown = success
```

**If errors occur:**
- Check Go is installed: `go version`
- Verify you're in backend directory: `pwd`
- Clean and rebuild: `go clean && go build -o bin/nagare-web-server ./cmd/server`

---

## Step 5: Start Server (2 minutes)

### 5.1 Stop Old Server (if running)
```bash
# Kill the process or Ctrl+C if in terminal
pkill -f nagare-web-server
```

### 5.2 Start New Server
```bash
# From backend directory
./bin/nagare-web-server
# or
go run ./cmd/server/main.go
```

**Expected output:**
```
Server running on :8080
Connected to database
Gmail service initialized
```

**Verify:**
```bash
# In another terminal
curl http://localhost:8080/api/v1/health
# Should return 200 OK
```

---

## Step 6: Create Test Media (2 minutes)

### 6.1 Call Create API

```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Gmail",
    "type": "gmail",
    "target": "your-email@gmail.com",
    "enabled": 1
  }'
```

**Expected response:**
```json
{
  "id": 3,
  "name": "Test Gmail",
  "type": "gmail",
  "target": "your-email@gmail.com",
  "enabled": 1,
  "created_at": "2024-01-15T10:45:00Z"
}
```

**Important:** Save the ID (usually 3, but might be different)

---

## Step 7: Test Gmail (3 minutes)

### 7.1 Send Test Email

Replace `3` with your media ID from above:

```bash
curl -X POST http://localhost:8080/api/v1/media/3/test \
  -H "Content-Type: application/json" \
  -d '{"message": "This is a test email from Nagare"}'
```

**Expected response (200 OK):**
```json
{
  "status": "success",
  "message": "test message sent successfully"
}
```

### 7.2 Check Email Inbox

- Open your email client
- Look for email from Nagare
- From: `your-email@gmail.com`
- Subject: `Nagare - Testing media connection`
- Should arrive in 30 seconds

**If email doesn't arrive:**
- Check spam folder
- Wait 2-3 minutes
- Check logs: `tail logs/system.log | grep gmail`

---

## Step 8: Verify Logs (1 minute)

Check that operations were logged:

```bash
tail logs/system.log | grep "test media"
```

**Expected output:**
```
INFO    test media succeeded    media_id=3 media_type=gmail media_name="Test Gmail" target=your-email@gmail.com
```

---

## 🎉 Success!

If you reached this point with:
✅ Token saved successfully
✅ Server started without errors
✅ Media created with 201 response
✅ Test endpoint returned 200 OK
✅ Email received in inbox
✅ Logs show success

**Gmail integration is working!**

---

## Troubleshooting Quick Fixes

### Problem: 500 Error on Test

**Solution 1: Check if Gmail is enabled**
```bash
grep -A 1 '"gmail"' backend/configs/nagare_config.json
# Check for "enabled": true
```

**Solution 2: Check if files exist**
```bash
ls backend/configs/gmail_*.json
# Should show both files
```

**Solution 3: Check logs for details**
```bash
tail -20 logs/system.log | grep -i gmail
```

### Problem: "gmail is disabled"

**Solution:**
- Open `backend/configs/nagare_config.json`
- Find `"enabled": false`
- Change to `"enabled": true`
- Restart server

### Problem: "credentials file not found"

**Solution:**
- Verify file exists: `ls backend/configs/gmail_credentials.json`
- If missing, download again from Google Cloud
- Restart server

### Problem: "token not found"

**Solution:**
- Run authorization script: `go run generate_gmail_token.go`
- Follow prompts to generate token
- Verify file created: `ls backend/configs/gmail_token.json`
- Restart server

### Problem: Email not arriving

**Solution 1:**
- Check spam/junk folder
- Wait 2-3 minutes
- Try again

**Solution 2:**
- Verify config has correct email: `grep "from" backend/configs/nagare_config.json`
- Try different recipient email
- Check for Gmail API errors: `tail logs/system.log | grep -i "gmail\|error"`

---

## Next Steps

### Once Everything Works

1. **Test with Triggers** - Create alert triggers that send Gmail notifications
2. **Create More Media** - Create Gmail media for different recipient emails
3. **Error Scenarios** - Deliberately break config to test error handling
4. **Integration Test** - Test full trigger → alert → Gmail flow

### See Documentation

- **GMAIL_DEBUG_SETUP_GUIDE.md** - Detailed setup information
- **GMAIL_TROUBLESHOOTING.md** - Comprehensive troubleshooting guide
- **GMAIL_TESTING_GUIDE.md** - Extended testing procedures
- **GMAIL_IMPLEMENTATION_SUMMARY.md** - Full project overview

---

## Need Help?

Check the logs:
```bash
tail -50 logs/system.log | grep -i gmail
```

Error message will tell you exactly what's wrong and how to fix it.

For example:
- Missing credentials? Error tells you path
- Missing token? Error shows how to generate
- Disabled Gmail? Error shows how to enable

---

**🚀 Ready to test more? See GMAIL_TESTING_GUIDE.md for comprehensive test cases and scenarios.**
