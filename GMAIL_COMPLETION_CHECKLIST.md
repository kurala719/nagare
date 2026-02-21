# Gmail Integration - Completion Checklist

## Phase 1: Code Implementation ✅

### Core Components

- [x] **GmailProvider.SendMessage()** - Main provider method
  - [x] Validates target email not empty
  - [x] Validates message not empty
  - [x] Checks if Gmail is enabled in config
  - [x] Returns clear error messages
  - [x] Proper error propagation

- [x] **getGmailClient()** - OAuth2 client retrieval
  - [x] Reads credentials file with error handling
  - [x] Parses JSON credentials
  - [x] Validates token file with error handling
  - [x] Returns informative error messages
  - [x] Includes file path in errors for debugging

- [x] **SendGmailServ()** - Email sending service
  - [x] Validates recipient email
  - [x] Validates subject not empty
  - [x] Validates body not empty
  - [x] Creates proper MIME message
  - [x] Handles encoding correctly
  - [x] Returns detailed errors

- [x] **SendGmailHTMLServ()** - HTML email support
  - [x] Similar validation as SendGmailServ
  - [x] Supports HTML content
  - [x] Proper MIME formatting

- [x] **tokenFromFile()** - Token loading utility
  - [x] Reads token file with error handling
  - [x] Validates JSON format
  - [x] Checks for access_token field
  - [x] Returns contextual errors

### Service Layer Integration

- [x] **TestMediaServ()** - Media testing function
  - [x] Loads media from database
  - [x] Validates media exists
  - [x] Logs operation start
  - [x] Calls SendIMReply with proper parameters
  - [x] Logs success or error
  - [x] Error wrapping with context

- [x] **SendIMReply()** - IM command dispatcher
  - [x] Validates mediaType parameter
  - [x] Validates target parameter
  - [x] Validates message parameter
  - [x] Provides logging for debug
  - [x] Error propagation from provider
  - [x] Proper error format

### API Layer

- [x] **TestMediaCtrl()** - API endpoint controller
  - [x] Already properly implemented
  - [x] Calls TestMediaServ correctly
  - [x] Returns appropriate responses

- [x] **Media configuration** in nagare_config.json
  - [x] `enabled` flag (false by default)
  - [x] `credentials_file` path
  - [x] `token_file` path
  - [x] `from` email address

---

## Phase 2: File Setup & Configuration

### Configuration Files

- [ ] **backend/configs/nagare_config.json**
  - [ ] File exists and is readable
  - [ ] Gmail section has `enabled: true`
  - [ ] Credentials file path is correct
  - [ ] Token file path is correct
  - [ ] From email is set

  ```json
  "gmail": {
    "enabled": true,
    "credentials_file": "configs/gmail_credentials.json",
    "token_file": "configs/gmail_token.json",
    "from": "your-email@gmail.com"
  }
  ```

### Required Files

- [ ] **backend/configs/gmail_credentials.json**
  - [ ] File exists at correct path
  - [ ] Contains valid Google OAuth2 credentials
  - [ ] Valid JSON format
  - [ ] Readable permissions (chmod 600)
  - [ ] Contains `client_id` field
  - [ ] Contains `client_secret` field

  Location: `backend/configs/gmail_credentials.json`

  Download from: [Google Cloud Console](https://console.cloud.google.com) → Credentials → OAuth 2.0 Desktop

- [ ] **backend/configs/gmail_token.json**
  - [ ] File exists at correct path
  - [ ] Contains valid OAuth2 token
  - [ ] Valid JSON format
  - [ ] Readable permissions (chmod 600)
  - [ ] Contains `access_token` field
  - [ ] Non-empty access_token value
  - [ ] Contains `refresh_token` field
  - [ ] Contains `token_type` field
  - [ ] Contains `expiry` field

  Location: `backend/configs/gmail_token.json`

  Format:
  ```json
  {
    "access_token": "ya29.a0AfH6SMB...",
    "token_type": "Bearer",
    "expiry": "2026-12-31T23:59:59Z",
    "refresh_token": "1//0gU..."
  }
  ```

  Generated via: Run Go authorization script provided in setup guide

### Google Cloud Project Setup

- [ ] **Google Cloud Project Created**
  - [ ] Project exists for Nagare
  - [ ] Project ID recorded

- [ ] **Gmail API Enabled**
  - [ ] Navigate to Gmail API page
  - [ ] Click "Enable" if not already enabled
  - [ ] Verify shows "API enabled" status

- [ ] **OAuth2 Credentials Created**
  - [ ] Go to Credentials page
  - [ ] Click "Create Credentials"
  - [ ] Selected "OAuth 2.0 Client ID"
  - [ ] Application type: Desktop application
  - [ ] JSON file downloaded

- [ ] **Application Authorized**
  - [ ] Ran authorization script
  - [ ] Clicked link to authorize
  - [ ] Grant permission to app
  - [ ] Token file generated and saved

---

## Phase 3: Code Compilation & Deployment

- [ ] **Go Backend Built**
  ```bash
  cd backend
  go build -o bin/nagare-web-server ./cmd/server
  ```
  - [ ] No compilation errors
  - [ ] Binary created at `backend/bin/nagare-web-server`
  - [ ] Binary is executable

- [ ] **Server Restarted**
  - [ ] Old server process killed
  - [ ] New binary started
  - [ ] Server listening on port 8080
  - [ ] No startup errors in logs
  - [ ] Gmail service initialized successfully

- [ ] **Database Connection**
  - [ ] Connected to database
  - [ ] Media table accessible
  - [ ] Can read/write media records

---

## Phase 4: Testing

### Pre-Test Verification

- [ ] **Configuration Validated**
  ```bash
  grep -A 5 '"gmail"' backend/configs/nagare_config.json
  ```
  - [ ] `enabled: true` confirmed
  - [ ] File paths correct

- [ ] **Files Verified**
  ```bash
  ls -la backend/configs/gmail_*.json
  ```
  - [ ] Both files exist and readable
  - [ ] File permissions set (chmod 600)

- [ ] **JSON Valid**
  ```bash
  python3 -m json.tool backend/configs/gmail_credentials.json
  python3 -m json.tool backend/configs/gmail_token.json
  ```
  - [ ] No JSON parsing errors
  - [ ] Files contain expected fields

### Unit Tests

- [ ] **Connect to Gmail API**
  - [ ] Token can authenticate
  - [ ] OAuth2 client created successfully
  - [ ] Credentials parsed correctly

- [ ] **Email Sending**
  - [ ] Test email sent successfully
  - [ ] Email received at recipient
  - [ ] Subject line correct
  - [ ] Message body preserved
  - [ ] From address correct

### Integration Tests

- [ ] **Create Media**
  ```bash
  POST /api/v1/media
  {
    "name": "Test Gmail",
    "type": "gmail",
    "target": "test@gmail.com",
    "enabled": 1
  }
  ```
  - [ ] Status 201 Created
  - [ ] Media ID returned
  - [ ] Record saved to database

- [ ] **Test Media Success**
  ```bash
  POST /api/v1/media/3/test
  ```
  - [ ] Status 200 OK
  - [ ] Response: "test message sent successfully"
  - [ ] Email received in inbox
  - [ ] Log shows success entry

- [ ] **Test Media Error Handling**
  - [ ] Test with disabled Gmail: Returns error about disabled
  - [ ] Test with missing credentials: Returns error about file
  - [ ] Test with invalid token: Returns error about token
  - [ ] Test with empty email: Returns error about empty target
  - [ ] All error messages clear and actionable

- [ ] **Logging**
  - [ ] Test success logged with media details
  - [ ] Test failure logged with error details
  - [ ] Logs include: media_id, media_type, media_name, target
  - [ ] Can tail logs and see operations

### End-to-End Tests

- [ ] **Multiple Recipients**
  - [ ] Create multiple media with different emails
  - [ ] Test each media
  - [ ] All emails received

- [ ] **Repeated Tests**
  - [ ] Send multiple test emails
  - [ ] All sent successfully
  - [ ] No rate limiting issues

- [ ] **Long-Running Tests**
  - [ ] Leave running overnight
  - [ ] Email token auto-refreshed
  - [ ] No memory leaks
  - [ ] Server stable

---

## Phase 5: Documentation

- [x] **GMAIL_DEBUG_SETUP_GUIDE.md Created**
  - [x] Configuration requirements documented
  - [x] Step-by-step setup instructions
  - [x] Common issues and solutions
  - [x] Google OAuth2 setup explained
  - [x] File format examples provided

- [x] **GMAIL_TROUBLESHOOTING.md Created**
  - [x] Error flow diagram explained
  - [x] All error points documented
  - [x] Error resolution guide
  - [x] Debugging commands provided
  - [x] Checklist included

- [x] **GMAIL_TESTING_GUIDE.md Created**
  - [x] Prerequisites listed
  - [x] Step-by-step testing procedure
  - [x] Expected responses documented
  - [x] Test cases provided
  - [x] Postman/REST client examples
  - [x] Success criteria defined

- [ ] **Add to Main Documentation**
  - [ ] Reference Gmail guide in DEVELOPER_GUIDE.md
  - [ ] Add Gmail media to API_REFERENCE.md
  - [ ] Include Gmail config in DEPLOYMENT_GUIDE.md

---

## Phase 6: Integration with Other Features

- [ ] **Trigger-Alert-Action System**
  - [ ] Trigger can generate alert
  - [ ] Alert can be sent via Gmail media
  - [ ] Test end-to-end trigger → alert → email

- [ ] **Webhook Media**
  - [ ] Similar error handling implemented
  - [ ] Consistent logging approach
  - [ ] Test webhook integration

- [ ] **QQ Media**
  - [ ] Similar error handling implemented
  - [ ] Consistent logging approach
  - [ ] Test QQ integration

- [ ] **WeChat Media**
  - [ ] Similar error handling implemented
  - [ ] Consistent logging approach
  - [ ] Test WeChat integration

---

## Phase 7: Security

- [ ] **Credentials Protection**
  - [ ] Config not committed to git
  - [ ] Credentials file not committed
  - [ ] Token file not committed
  - [ ] .gitignore includes `configs/gmail_*.json`

- [ ] **Environment Secrets**
  - [ ] Consider environment variables for sensitive config
  - [ ] Credentials can be passed at runtime
  - [ ] No secrets in code

- [ ] **Scoping**
  - [ ] App only has `gmail.send` scope
  - [ ] Cannot read other emails
  - [ ] Cannot modify email folders
  - [ ] Cannot delete emails

- [ ] **Token Management**
  - [ ] Refresh token stored securely
  - [ ] Token auto-refreshes when expired
  - [ ] Old tokens can be revoked

---

## Phase 8: Production Deployment

- [ ] **Build Production Binary**
  ```bash
  cd backend
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -o bin/nagare-web-server-linux ./cmd/server
  ```
  - [ ] Binary builds without errors
  - [ ] Binary optimized for target OS

- [ ] **Production Configuration**
  - [ ] `nagare_config.json` updated for production
  - [ ] Credentials uploaded to production server
  - [ ] Token uploaded to production server
  - [ ] Permissions set to 600 (chmod 600)

- [ ] **Database Migrations**
  - [ ] Media table created if needed
  - [ ] Indexes created
  - [ ] Permissions set

- [ ] **Monitoring & Logging**
  - [ ] Log rotation configured
  - [ ] Mail service logs monitored
  - [ ] Failed emails logged and alerted
  - [ ] Daily email summary sent

- [ ] **Backup & Recovery**
  - [ ] Credentials backed up securely
  - [ ] Token backed up (non-production backup)
  - [ ] Can restore quickly if needed

---

## Sign-Off Checklist

### Development Team

- [ ] Code reviewed and approved
- [ ] All tests passed
- [ ] Documentation complete
- [ ] No known bugs or issues
- [ ] Performance acceptable
- [ ] Error handling comprehensive

### QA Team

- [ ] All test cases passed
- [ ] Error scenarios validated
- [ ] Edge cases tested
- [ ] Documentation verified
- [ ] No regressions found
- [ ] Ready for production

### DevOps Team

- [ ] Production environment ready
- [ ] Deployment procedure documented
- [ ] Rollback plan in place
- [ ] Monitoring configured
- [ ] Alerts configured
- [ ] Backup/recovery tested

### Product Team

- [ ] Feature meets requirements
- [ ] User documentation prepared
- [ ] Support team trained
- [ ] Deployment date set
- [ ] Communication plan in place
- [ ] Ready for release

---

## Rollout Plan

### Phase 1: Internal Testing (Day 1-2)
- [ ] Deploy to staging environment
- [ ] Full integration testing
- [ ] Performance baseline established
- [ ] Documentation reviewed

### Phase 2: Beta Testing (Day 3-7)
- [ ] Deploy to beta users
- [ ] Collect feedback
- [ ] Monitor for issues
- [ ] Prepare for production

### Phase 3: Production Rollout (Day 8)
- [ ] Deploy to production
- [ ] Monitor closely first 24 hours
- [ ] Respond to issues quickly
- [ ] Announce to users

### Phase 4: Post-Launch (Day 9+)
- [ ] Monitor for issues
- [ ] Collect user feedback
- [ ] Plan improvements
- [ ] Document lessons learned

---

## Success Metrics

- [ ] **Uptime:** 99.9% or higher
- [ ] **Email Delivery:** 99.5% or higher
- [ ] **Response Time:** < 1 second for test endpoint
- [ ] **Error Rate:** < 0.5% of requests
- [ ] **User Satisfaction:** >= 4/5 stars
- [ ] **Support Tickets:** < 2 per week

---

## Next Steps

1. **Complete All Checkboxes Above**
   - [ ] Verify each item is complete
   - [ ] Document any deviations
   - [ ] Get team sign-off

2. **Complete Other Media Types**
   - [ ] Implement webhook media type
   - [ ] Implement QQ media type
   - [ ] Implement WeChat media type
   - [ ] Create similar testing guides

3. **Integrate with Triggers**
   - [ ] Test trigger → alert → Gmail
   - [ ] Test multiple media types in cascade
   - [ ] Test failure handling and retries

4. **Production Deployment**
   - [ ] Execute rollout plan
   - [ ] Monitor closely
   - [ ] Respond to issues

5. **Post-Launch Improvements**
   - [ ] Collect user feedback
   - [ ] Plan enhancements
   - [ ] Schedule follow-up release

---

## Contact & Support

For questions or issues:

1. **Check Documentation**
   - GMAIL_DEBUG_SETUP_GUIDE.md - Setup help
   - GMAIL_TROUBLESHOOTING.md - Error resolution
   - GMAIL_TESTING_GUIDE.md - Testing help

2. **Review Code**
   - internal/repository/media/gmail.go
   - internal/service/media.go
   - internal/service/im_command.go

3. **Check Logs**
   - logs/system.log
   - Error messages include file paths and suggestions

4. **Ask Development Team**
   - Create issue with error details
   - Include relevant logs
   - Describe reproduction steps
