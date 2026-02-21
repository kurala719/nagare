# Gmail Media Testing Guide

## Prerequisites

Before testing, ensure you have:

1. ✅ Gmail enabled in config
2. ✅ Credentials file downloaded
3. ✅ Token file generated
4. ✅ Media created with Gmail type
5. ✅ Server rebuilt and restarted

## Step 1: Rebuild and Start Server

```bash
# Navigate to backend
cd backend

# Rebuild the project
go build -o bin/nagare-web-server ./cmd/server

# Start the server
./bin/nagare-web-server
# or
go run ./cmd/server/main.go
```

**Expected Output:**
```
Server running on :8080
Connected to database
Gmail service initialized
```

---

## Step 2: Check Configuration

Before testing, verify Gmail is enabled:

```bash
# View Gmail configuration
cat configs/nagare_config.json | grep -A 5 '"gmail"'
```

**Expected Output:**
```json
"gmail": {
  "enabled": true,
  "credentials_file": "configs/gmail_credentials.json",
  "token_file": "configs/gmail_token.json",
  "from": "your-email@gmail.com"
}
```

If `enabled` is `false`, update it:

```bash
# Edit the file
nano configs/nagare_config.json
# Change "enabled": false to "enabled": true
# Save and restart server
```

---

## Step 3: Verify Configuration Files

Check that both files exist and are readable:

```bash
# Linux/Mac
ls -la configs/gmail_credentials.json configs/gmail_token.json

# Windows PowerShell
Get-Item configs\gmail_credentials.json, configs\gmail_token.json
```

**Expected Output:**
```
-rw------x 1 user group  1234 Jan 15 10:30 configs/gmail_credentials.json
-rw------x 1 user group   567 Jan 15 10:35 configs/gmail_token.json
```

Both files should exist and be readable.

---

## Step 4: Create a Gmail Media

Create a new Gmail media entry if you don't have one:

**Endpoint:** `POST /api/v1/media`

**HTTP Request:**
```bash
curl -X POST http://localhost:8080/api/v1/media \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Gmail",
    "type": "gmail",
    "target": "your-email@gmail.com",
    "enabled": 1,
    "description": "Test media for Gmail"
  }'
```

**Expected Response:**
```json
{
  "id": 3,
  "name": "Test Gmail",
  "type": "gmail",
  "target": "your-email@gmail.com",
  "enabled": 1,
  "description": "Test media for Gmail",
  "created_at": "2024-01-15T10:45:00Z"
}
```

**Save the ID** (e.g., `3`) for next step.

---

## Step 5: Test the Media

Test sending a message through the Gmail media:

**Endpoint:** `POST /api/v1/media/:id/test`

**HTTP Request:**
```bash
curl -X POST http://localhost:8080/api/v1/media/3/test \
  -H "Content-Type: application/json" \
  -d '{
    "message": "This is a test message from Nagare"
  }'
```

**Expected Success Response (200 OK):**
```json
{
  "status": "success",
  "message": "test message sent successfully"
}
```

**Possible Error Responses:**

1. **Gmail Disabled:**
   ```json
   {
     "status": "error",
     "error": "gmail is disabled in configuration (set gmail.enabled=true)"
   }
   ```
   → Enable in config, restart server

2. **Credentials Missing:**
   ```json
   {
     "status": "error",
     "error": "unable to read credentials file at configs/gmail_credentials.json: no such file or directory"
   }
   ```
   → Download credentials from Google Cloud

3. **Token Missing:**
   ```json
   {
     "status": "error",
     "error": "gmail token not found or invalid at configs/gmail_token.json"
   }
   ```
   → Generate token using provided Go script

4. **Invalid Email:**
   ```json
   {
     "status": "error",
     "error": "target email address cannot be empty"
   }
   ```
   → Update media with valid email target

---

## Step 6: Check Logs

If test fails, check logs for detailed error information:

```bash
# View last 50 lines of logs
tail -50 logs/system.log

# Search for Gmail errors
grep -i gmail logs/system.log

# Search for test media operations
grep "test media" logs/system.log
```

**Expected Log Line on Success:**
```
INFO    test media succeeded    media_id=3 media_type=gmail media_name="Test Gmail" target=your-email@gmail.com
```

**Expected Log Line on Error:**
```
ERROR   test media failed    media_id=3 media_type=gmail error="<specific error message>"
```

---

## Step 7: Verify Email Received

Check your email inbox for the test message:

1. Open email client (Gmail, Outlook, etc.)
2. Look for email from: `nagare-system@example.com` (or configured `from` address)
3. Subject line: `Nagare - Testing media connection`
4. Message body: Contains "This is a test message from Nagare"

**If email received:**
✅ Gmail integration is working correctly!

**If email NOT received after 5 minutes:**
- Check spam/junk folder
- Try test again
- Check logs for delivery errors

---

## Complete Testing Flow

### Test Case 1: Basic Functionality

```bash
# 1. Create media
curl -X POST http://localhost:8080/api/v1/media \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gmail Test",
    "type": "gmail",
    "target": "test@gmail.com",
    "enabled": 1
  }'
# Expected: 201 Created with media ID

# 2. Test media
curl -X POST http://localhost:8080/api/v1/media/3/test \
  -H "Content-Type: application/json" \
  -d '{"message": "Test from Nagare"}'
# Expected: 200 OK with success message

# 3. Check logs
tail logs/system.log | grep -i gmail
# Expected: SUCCESS log entry

# 4. Check email
# Expected: Email received in inbox
```

### Test Case 2: Error Handling - Disabled Gmail

```bash
# 1. Disable Gmail in config
# Edit configs/nagare_config.json
# Set "enabled": false

# 2. Restart server
# Kill current server and restart

# 3. Try to test
curl -X POST http://localhost:8080/api/v1/media/3/test
# Expected: 500 with "gmail is disabled" message

# 4. Check logs
tail logs/system.log | grep "gmail is disabled"
```

### Test Case 3: Error Handling - Missing Credentials

```bash
# 1. Rename credentials file temporarily
mv backend/configs/gmail_credentials.json backend/configs/gmail_credentials.json.bak

# 2. Try to test
curl -X POST http://localhost:8080/api/v1/media/3/test
# Expected: 500 with "unable to read credentials file" message

# 3. Restore file
mv backend/configs/gmail_credentials.json.bak backend/configs/gmail_credentials.json
```

### Test Case 4: Error Handling - Invalid Email

```bash
# 1. Create media with empty target
curl -X POST http://localhost:8080/api/v1/media \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Invalid Email Test",
    "type": "gmail",
    "target": "",
    "enabled": 1
  }'

# 2. Try to test
curl -X POST http://localhost:8080/api/v1/media/4/test
# Expected: 500 with "email address cannot be empty" message
```

---

## Using REST Client (VS Code)

Create a file `test_gmail.http`:

```http
### Create Gmail Media
POST http://localhost:8080/api/v1/media
Content-Type: application/json

{
  "name": "Test Gmail",
  "type": "gmail",
  "target": "your-email@gmail.com",
  "enabled": 1
}

### Test Gmail Media
POST http://localhost:8080/api/v1/media/3/test
Content-Type: application/json

{
  "message": "Test message from Nagare"
}

### Get Media Details
GET http://localhost:8080/api/v1/media/3

### List All Media
GET http://localhost:8080/api/v1/media

### Update Media Target
PUT http://localhost:8080/api/v1/media/3
Content-Type: application/json

{
  "target": "new-email@gmail.com"
}
```

Click "Send Request" to test each endpoint.

---

## Using Postman

1. **Create Collection:** "Gmail Testing"

2. **Request 1: Create Media**
   - Method: POST
   - URL: `http://localhost:8080/api/v1/media`
   - Body (JSON):
     ```json
     {
       "name": "Test Gmail",
       "type": "gmail",
       "target": "your-email@gmail.com",
       "enabled": 1
     }
     ```

3. **Request 2: Test Media**
   - Method: POST
   - URL: `http://localhost:8080/api/v1/media/3/test`
   - Body (JSON):
     ```json
     {
       "message": "Test from Postman"
     }
     ```

4. **Run Tests:**
   - Click "Send" for each request
   - Verify responses
   - Check email inbox

---

## Troubleshooting During Testing

### Problem: Connection Refused

```
Error: Connection refused on localhost:8080
```

**Solution:**
1. Check server is running: `ps aux | grep nagare`
2. Look for startup errors in terminal
3. Check port 8080 is available: `netstat -an | grep 8080`
4. Restart server if necessary

### Problem: JSON Parsing Error

```
Error: invalid character in JSON
```

**Solution:**
1. Validate JSON syntax: `python3 -m json.tool request.json`
2. Check for unescaped quotes
3. Verify Content-Type header is `application/json`

### Problem: Media Not Found

```json
{
  "error": "media not found"
}
```

**Solution:**
1. Verify media ID exists
2. Create new media if needed
3. List all media: `GET /api/v1/media`

### Problem: Authentication Failed

```json
{
  "error": "authentication failed"
}
```

**Solution:**
1. Check bearer token if required
2. Verify API keys are set
3. Check authorization headers

---

## Performance Testing

### Test 1: Multiple Recipients

```bash
# Send to multiple different emails
for i in {1..5}; do
  curl -X POST http://localhost:8080/api/v1/media/$i/test \
    -H "Content-Type: application/json" \
    -d '{"message": "Test message '$i'"}'
done
```

### Test 2: Rapid Requests

```bash
# Send 10 rapid requests
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/media/3/test &
done
wait
```

### Test 3: Large Message

```bash
# Send large message
MESSAGE=$(head -c 10000 /dev/urandom | base64)
curl -X POST http://localhost:8080/api/v1/media/3/test \
  -H "Content-Type: application/json" \
  -d "{\"message\": \"$MESSAGE\"}"
```

---

## Success Criteria

✅ **Gmail Integration is Working if:**
- [ ] API returns 200 OK on test
- [ ] Response shows "test message sent successfully"
- [ ] Email received in inbox within 1 minute
- [ ] Logs show SUCCESS entry for test media
- [ ] Error responses are clear and actionable
- [ ] Sender email is correct
- [ ] Message content is preserved
- [ ] Can send multiple emails without errors
- [ ] Different recipient emails work

---

## Next Steps After Success

Once Gmail testing works:

1. **Test Other Media Types:**
   - Webhook
   - QQ
   - WeChat

2. **Integrate with Triggers:**
   - Create alert triggers
   - Configure actions to send via Gmail
   - Test end-to-end flow

3. **Deploy to Production:**
   - Update production config
   - Ensure credentials secured
   - Set up logging and monitoring
   - Test daily digest emails

4. **Document in Post-Implementation Checklist**

---

## Support

If tests fail consistently:

1. Check [GMAIL_DEBUG_SETUP_GUIDE.md](GMAIL_DEBUG_SETUP_GUIDE.md) for configuration issues
2. Check [GMAIL_TROUBLESHOOTING.md](GMAIL_TROUBLESHOOTING.md) for error resolution
3. Review logs: `tail logs/system.log | grep -i gmail`
4. Verify Google Cloud project and credentials
5. Test with simple curl command first before using tools
