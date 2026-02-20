package monitors

import (
	"context"
	"encoding/json"
	"fmt"
)

// SetupWebhookMedia creates the necessary Media Type, User, and Action in Zabbix for Nagare Webhook
func (p *ZabbixProvider) SetupWebhookMedia(ctx context.Context, webhookURL, eventToken string) error {
	// 1. Create Media Type
	script := `var req = new CurlHttpRequest();
req.AddHeader('Content-Type: application/json');
if (value.event_token !== "") {
    req.AddHeader('Authorization: Bearer ' + value.event_token);
}
var payload = {
    message: value.message,
    severity: value.severity,
    host: value.host,
    item: value.item,
    eventid: value.eventid
};
var resp = req.Post(value.url, JSON.stringify(payload));
return resp;`

	mediaTypeParams := map[string]interface{}{
		"name":   "Nagare Webhook",
		"type":   "4", // Webhook
		"script": script,
		"parameters": []map[string]string{
			{"name": "url", "value": webhookURL},
			{"name": "event_token", "value": eventToken},
			{"name": "message", "value": "{ALERT.MESSAGE}"},
			{"name": "severity", "value": "{EVENT.SEVERITY}"},
			{"name": "host", "value": "{HOST.HOST}"},
			{"name": "item", "value": "{ITEM.NAME}"},
			{"name": "eventid", "value": "{EVENT.ID}"},
		},
	}

	// check if exists
	checkMedia, err := p.sendRequest(ctx, "mediatype.get", map[string]interface{}{
		"filter": map[string]interface{}{
			"name": "Nagare Webhook",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to check media type: %w", err)
	}

	var mediaTypeID string
	var mediaTypes []map[string]interface{}
	// Unmarshal result to get the media types
	// Note: We need to use json.Unmarshal since sendRequest returns raw result
	if err := json.Unmarshal(checkMedia.Result, &mediaTypes); err == nil && len(mediaTypes) > 0 {
		mediaTypeID = fmt.Sprintf("%v", mediaTypes[0]["mediatypeid"])
		// Update existing
		mediaTypeParams["mediatypeid"] = mediaTypeID
		_, err = p.sendRequest(ctx, "mediatype.update", mediaTypeParams)
		if err != nil {
			return fmt.Errorf("failed to update media type: %w", err)
		}
	} else {
		// Create new
		createResp, err := p.sendRequest(ctx, "mediatype.create", mediaTypeParams)
		if err != nil {
			return fmt.Errorf("failed to create media type: %w", err)
		}
		var createResult map[string]interface{}
		if err := json.Unmarshal(createResp.Result, &createResult); err == nil {
			if ids, ok := createResult["mediatypeids"].([]interface{}); ok && len(ids) > 0 {
				mediaTypeID = fmt.Sprintf("%v", ids[0])
			}
		}
	}

	if mediaTypeID == "" {
		return fmt.Errorf("failed to get or create media type ID")
	}

	// 2. Create User
	// Check user group (Administrators usually ID 7)
	userParams := map[string]interface{}{
		"username": "nagare_webhook_user",
		"passwd":   "NagareUser123!",
		"usrgrps": []map[string]interface{}{
			{"usrgrpid": "7"},
		},
		"usrmedias": []map[string]interface{}{
			{
				"mediatypeid": mediaTypeID,
				"sendto":      "nagare",
				"active":      0,
				"severity":    63,
				"period":      "1-7,00:00-24:00",
			},
		},
	}

	checkUser, err := p.sendRequest(ctx, "user.get", map[string]interface{}{
		"filter": map[string]interface{}{
			"username": "nagare_webhook_user",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to check user: %w", err)
	}

	var userID string
	var users []map[string]interface{}
	if err := json.Unmarshal(checkUser.Result, &users); err == nil && len(users) > 0 {
		userID = fmt.Sprintf("%v", users[0]["userid"])
		userParams["userid"] = userID
		_, err = p.sendRequest(ctx, "user.update", userParams)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
	} else {
		createResp, err := p.sendRequest(ctx, "user.create", userParams)
		if err != nil {
			// fallback without usrgrpid 7 if it fails (maybe different ID)
			return fmt.Errorf("failed to create user (make sure user group 7 exists): %w", err)
		}
		var createResult map[string]interface{}
		if err := json.Unmarshal(createResp.Result, &createResult); err == nil {
			if ids, ok := createResult["userids"].([]interface{}); ok && len(ids) > 0 {
				userID = fmt.Sprintf("%v", ids[0])
			}
		}
	}

	if userID == "" {
		return fmt.Errorf("failed to get or create user ID")
	}

	// 3. Create Action
	actionParams := map[string]interface{}{
		"name":        "Nagare Webhook Action",
		"eventsource": 0, // Trigger
		"status":      0, // Enabled
		"esc_period":  "1m",
		"filter": map[string]interface{}{
			"evaltype": 0, // And/Or
			"conditions": []map[string]interface{}{
				{
					"conditiontype": 14, // Event acknowledged
					"operator":      0,
					"value":         "0", // Not acked
				},
			},
		},
		"operations": []map[string]interface{}{
			{
				"operationtype": 0, // Send message
				"esc_period":    "0",
				"esc_step_from": 1,
				"esc_step_to":   1,
				"opmessage_usr": []map[string]interface{}{
					{"userid": userID},
				},
				"opmessage": map[string]interface{}{
					"default_msg": 1,
					"mediatypeid": mediaTypeID,
				},
			},
		},
	}

	checkAction, err := p.sendRequest(ctx, "action.get", map[string]interface{}{
		"filter": map[string]interface{}{
			"name": "Nagare Webhook Action",
		},
	})
	if err != nil {
		return fmt.Errorf("failed to check action: %w", err)
	}

	var actions []map[string]interface{}
	if err := json.Unmarshal(checkAction.Result, &actions); err == nil && len(actions) > 0 {
		actionParams["actionid"] = fmt.Sprintf("%v", actions[0]["actionid"])
		_, err = p.sendRequest(ctx, "action.update", actionParams)
		if err != nil {
			return fmt.Errorf("failed to update action: %w", err)
		}
	} else {
		_, err = p.sendRequest(ctx, "action.create", actionParams)
		if err != nil {
			return fmt.Errorf("failed to create action: %w", err)
		}
	}

	return nil
}
