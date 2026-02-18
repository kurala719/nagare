package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"nagare/internal/model"
	"nagare/internal/repository"
	"nagare/internal/repository/monitors"

	"github.com/spf13/viper"
)

// AlarmReq represents an alarm request
type AlarmReq struct {
	Name        string `json:"name" binding:"required"`
	URL         string `json:"url" binding:"required"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	AuthToken   string `json:"auth_token"`
	EventToken  string `json:"event_token"`
	Description string `json:"description"`
	Type        int    `json:"type" binding:"required,oneof=1 2 3"`
	Enabled     int    `json:"enabled"`
}

// AlarmResp represents an alarm response
type AlarmResp struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	AuthToken   string `json:"auth_token"`
	EventToken  string `json:"event_token"`
	Description string `json:"description"`
	Type        int    `json:"type"`
	Enabled     int    `json:"enabled"`
	Status      int    `json:"status"`
	StatusDesc  string `json:"status_description"`
}

func generateAlarmEventToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetAllAlarmsServ retrieves all alarms
func GetAllAlarmsServ() ([]AlarmResp, error) {
	alarms, err := repository.GetAllAlarmsDAO()
	if err != nil {
		return nil, fmt.Errorf("failed to get alarms: %w", err)
	}

	result := make([]AlarmResp, 0, len(alarms))
	for _, a := range alarms {
		result = append(result, alarmToResp(a))
	}
	return result, nil
}

// SearchAlarmsServ retrieves alarms by filter
func SearchAlarmsServ(filter model.AlarmFilter) ([]AlarmResp, error) {
	alarms, err := repository.SearchAlarmsDAO(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search alarms: %w", err)
	}
	result := make([]AlarmResp, 0, len(alarms))
	for _, a := range alarms {
		result = append(result, alarmToResp(a))
	}
	return result, nil
}

// CountAlarmsServ returns total count for alarms by filter
func CountAlarmsServ(filter model.AlarmFilter) (int64, error) {
	return repository.CountAlarmsDAO(filter)
}

// GetAlarmByIDServ retrieves an alarm by ID
func GetAlarmByIDServ(id uint) (AlarmResp, error) {
	alarm, err := repository.GetAlarmByIDDAO(id)
	if err != nil {
		return AlarmResp{}, fmt.Errorf("failed to get alarm: %w", err)
	}
	return alarmToResp(alarm), nil
}

// GetAlarmByEventTokenServ retrieves an alarm by event token
func GetAlarmByEventTokenServ(eventToken string) (model.Alarm, error) {
	if strings.TrimSpace(eventToken) == "" {
		return model.Alarm{}, model.ErrUnauthorized
	}
	return repository.GetAlarmByEventTokenDAO(eventToken)
}

// AddAlarmServ creates a new alarm and optionally attempts to login
func AddAlarmServ(a AlarmReq) (AlarmResp, error) {
	eventToken := strings.TrimSpace(a.EventToken)
	if eventToken == "" {
		generated, err := generateAlarmEventToken()
		if err != nil {
			return AlarmResp{}, fmt.Errorf("failed to generate event token: %w", err)
		}
		eventToken = generated
	}

	alarm := model.Alarm{
		Name:        a.Name,
		URL:         a.URL,
		Username:    a.Username,
		Password:    a.Password,
		AuthToken:   a.AuthToken,
		EventToken:  eventToken,
		Description: a.Description,
		Type:        a.Type,
		Enabled:     a.Enabled,
		Status:      determineAlarmStatus(model.Alarm{Enabled: a.Enabled, AuthToken: a.AuthToken, Username: a.Username, Password: a.Password}),
	}

	if err := repository.AddAlarmDAO(alarm); err != nil {
		return AlarmResp{}, fmt.Errorf("failed to create alarm: %w", err)
	}

	alarms, err := repository.SearchAlarmsDAO(model.AlarmFilter{Query: a.Name})
	if err != nil || len(alarms) == 0 {
		return AlarmResp{}, fmt.Errorf("failed to retrieve created alarm")
	}
	createdAlarm := alarms[len(alarms)-1]
	result := alarmToResp(createdAlarm)

	if a.Username != "" && a.Password != "" {
		loggedInAlarm, err := LoginAlarmServ(createdAlarm.ID)
		if err != nil {
			LogService("warn", "auto-login failed for alarm", map[string]interface{}{"alarm_id": createdAlarm.ID, "error": err.Error()}, nil, "")
		} else {
			result = loggedInAlarm
		}
	}

	return result, nil
}

// DeleteAlarmServByID deletes an alarm by ID
func DeleteAlarmServByID(id int) error {
	return repository.DeleteAlarmByIDDAO(id)
}

// UpdateAlarmServ updates an existing alarm
func UpdateAlarmServ(id int, a AlarmReq) error {
	eventToken := strings.TrimSpace(a.EventToken)
	if eventToken == "" {
		existing, err := GetAlarmByIDServ(uint(id))
		if err != nil {
			return err
		}
		eventToken = existing.EventToken
	}
	updated := model.Alarm{
		Name:        a.Name,
		URL:         a.URL,
		Username:    a.Username,
		Password:    a.Password,
		AuthToken:   a.AuthToken,
		EventToken:  eventToken,
		Description: a.Description,
		Type:        a.Type,
		Enabled:     a.Enabled,
		Status:      determineAlarmStatus(model.Alarm{Enabled: a.Enabled, AuthToken: a.AuthToken, Username: a.Username, Password: a.Password}),
	}
	if err := repository.UpdateAlarmDAO(id, updated); err != nil {
		return err
	}
	return nil
}

// RegenerateAlarmEventTokenServ creates a new event token for an alarm
func RegenerateAlarmEventTokenServ(id uint) (AlarmResp, error) {
	newToken, err := generateAlarmEventToken()
	if err != nil {
		return AlarmResp{}, fmt.Errorf("failed to generate event token: %w", err)
	}
	if err := repository.UpdateAlarmEventTokenDAO(id, newToken); err != nil {
		return AlarmResp{}, fmt.Errorf("failed to save event token: %w", err)
	}
	updatedAlarm, err := GetAlarmByIDServ(id)
	if err != nil {
		return AlarmResp{}, fmt.Errorf("failed to retrieve updated alarm: %w", err)
	}
	return updatedAlarm, nil
}

// ValidateAlarmEventTokenServ validates inbound event tokens for webhook use
func ValidateAlarmEventTokenServ(eventToken string) error {
	if strings.TrimSpace(eventToken) == "" {
		return model.ErrUnauthorized
	}
	_, err := repository.GetAlarmByEventTokenDAO(eventToken)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.ErrUnauthorized
		}
		return err
	}
	return nil
}

// LoginAlarmServ authenticates with an alarm source and stores the auth token
func LoginAlarmServ(id uint) (AlarmResp, error) {
	alarm, err := GetAlarmByIDServ(id)
	if err != nil {
		return AlarmResp{}, err
	}

	client, err := monitors.NewClient(monitors.Config{
		Name: alarm.Name,
		Type: monitors.ParseMonitorType(alarm.Type),
		Auth: monitors.AuthConfig{
			URL:      alarm.URL,
			Username: alarm.Username,
			Password: alarm.Password,
			Token:    alarm.AuthToken,
		},
		Timeout: 30,
	})
	if err != nil {
		return AlarmResp{}, fmt.Errorf("failed to create alarm client: %w", err)
	}

	ctx := context.Background()
	if err := client.Authenticate(ctx); err != nil {
		return AlarmResp{}, fmt.Errorf("authentication failed: %w", err)
	}

	authToken := client.GetAuthToken()
	if authToken == "" {
		return AlarmResp{}, fmt.Errorf("authentication succeeded but no token received")
	}

	if err := repository.UpdateAlarmAuthTokenDAO(id, authToken); err != nil {
		return AlarmResp{}, fmt.Errorf("failed to save auth token: %w", err)
	}
	_ = repository.UpdateAlarmStatusDAO(id, 1)

	updatedAlarm, err := GetAlarmByIDServ(id)
	if err != nil {
		return AlarmResp{}, fmt.Errorf("failed to retrieve updated alarm: %w", err)
	}

	return updatedAlarm, nil
}

// SetupAlarmMediaTypeServ configures the Nagare media type in the external alarm system
func SetupAlarmMediaTypeServ(id uint) error {
	LogSystem("debug", "starting media type setup", map[string]interface{}{"alarm_id": id}, nil, "")

	alarm, err := repository.GetAlarmByIDDAO(id)
	if err != nil {
		LogSystem("error", "failed to get alarm", map[string]interface{}{"alarm_id": id, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to get alarm: %w", err)
	}

	LogSystem("debug", "alarm retrieved", map[string]interface{}{"alarm_id": id, "alarm_type": alarm.Type, "alarm_name": alarm.Name}, nil, "")

	if alarm.Type != 1 { // Only Zabbix (1) for now
		LogSystem("error", "unsupported alarm type", map[string]interface{}{"alarm_id": id, "alarm_type": alarm.Type}, nil, "")
		return fmt.Errorf("alarm type %d does not support automatic media type setup", alarm.Type)
	}

	ip := viper.GetString("system.ip_address")
	port := viper.GetInt("system.port")
	if ip == "" {
		ip = "localhost"
	}
	if port == 0 {
		port = 8080
	}

	webhookURL := fmt.Sprintf("http://%s:%d/api/v1/alerts/webhook", ip, port)
	LogSystem("info", "webhook URL configured for media type", map[string]interface{}{
		"url":      webhookURL,
		"alarm_id": id,
		"note":     "Ensure Zabbix can reach this URL. Test with: curl " + webhookURL + "/health",
	}, nil, "")

	cfg := monitors.Config{
		Name:    alarm.Name,
		Type:    monitors.ParseMonitorType(alarm.Type),
		Timeout: 30,
		Auth: monitors.AuthConfig{
			URL:      alarm.URL,
			Username: alarm.Username,
			Password: alarm.Password,
			Token:    alarm.AuthToken,
		},
	}

	client, err := monitors.NewClient(cfg)
	if err != nil {
		LogSystem("error", "failed to create monitor client", map[string]interface{}{"alarm_id": id, "error": err.Error()}, nil, "")
		return fmt.Errorf("failed to create monitor client: %w", err)
	}

	ctx := context.Background()
	if alarm.AuthToken == "" {
		LogSystem("debug", "authenticating with alarm source", map[string]interface{}{"alarm_id": id}, nil, "")
		if err := client.Authenticate(ctx); err != nil {
			LogSystem("error", "authentication failed", map[string]interface{}{"alarm_id": id, "error": err.Error()}, nil, "")
			return fmt.Errorf("failed to authenticate with alarm source: %w", err)
		}
		// Update token in DB
		alarm.AuthToken = client.GetAuthToken()
		_ = repository.UpdateAlarmAuthTokenDAO(alarm.ID, alarm.AuthToken)
		LogSystem("debug", "authentication successful, token updated", map[string]interface{}{"alarm_id": id}, nil, "")
	}

	script := `try {
	var params = JSON.parse(value),
		req = new HttpRequest(),
		data = {},
		resp;

	if (!params.url) {
		throw 'Missing required parameter: url';
	}

	if (typeof params.HTTPProxy === 'string' && params.HTTPProxy !== '') {
		req.setProxy(params.HTTPProxy);
	}

	req.addHeader('Content-Type: application/json');
	req.addHeader('X-Alarm-Token: ' + params.token);

	data.subject = params.subject;
	data.message = params.message;
	data.alert_subject = params.alert_subject;
	data.alert_message = params.alert_message;
	data.event_id = params.eventid;
	data.event_name = params.event_name;
	data.event_date = params.event_date;
	data.event_time = params.event_time;
	data.event_source = params.event_source;
	data.event_object = params.event_object;
	data.event_value = params.event_value;
	data.event_update_status = params.event_update_status;
	data.event_nseverity = params.event_nseverity;
	data.event_severity = params.event_severity;
	data.event_tags = params.event_tags;
	data.event_opdata = params.event_opdata;
	data.host_id = params.host_id;
	data.host = params.host;
	data.host_ip = params.host_ip;
	data.host_dns = params.host_dns;
	data.trigger_id = params.trigger_id;
	data.trigger = params.trigger;
	data.trigger_status = params.trigger_status;
	data.trigger_severity = params.trigger_severity;
	data.trigger_expression = params.trigger_expression;
	data.severity = params.severity;
	data.token = params.token;

	Zabbix.Log(4, '[ Nagare Webhook ] Sending request to ' + params.url);

	resp = req.post(params.url, JSON.stringify(data));

	if (req.getStatus() !== 200 && req.getStatus() !== 202) {
		throw 'Response code: ' + req.getStatus();
	}

	return resp;
} catch (error) {
	Zabbix.Log(3, '[ Nagare Webhook ] Notification failed : ' + error);
	throw 'Notification failed : ' + error;
}`

	params := map[string]string{
		"url":                 webhookURL,
		"token":               alarm.EventToken,
		"subject":             "{ALERT.SUBJECT}",
		"message":             "{ALERT.MESSAGE}",
		"alert_subject":       "{ALERT.SUBJECT}",
		"alert_message":       "{ALERT.MESSAGE}",
		"eventid":             "{EVENT.ID}",
		"event_name":          "{EVENT.NAME}",
		"event_date":          "{EVENT.DATE}",
		"event_time":          "{EVENT.TIME}",
		"event_source":        "{EVENT.SOURCE}",
		"event_object":        "{EVENT.OBJECT}",
		"event_value":         "{EVENT.VALUE}",
		"event_update_status": "{EVENT.UPDATE.STATUS}",
		"event_nseverity":     "{EVENT.NSEVERITY}",
		"event_severity":      "{EVENT.SEVERITY}",
		"event_tags":          "{EVENT.TAGSJSON}",
		"event_opdata":        "{EVENT.OPDATA}",
		"host_id":             "{HOST.ID}",
		"host":                "{HOST.NAME}",
		"host_ip":             "{HOST.IP}",
		"host_dns":            "{HOST.DNS}",
		"trigger_id":          "{TRIGGER.ID}",
		"trigger":             "{TRIGGER.NAME}",
		"trigger_status":      "{TRIGGER.STATUS}",
		"trigger_severity":    "{TRIGGER.SEVERITY}",
		"trigger_expression":  "{TRIGGER.EXPRESSION}",
		"severity":            "{TRIGGER.SEVERITY}",
	}

	mediaTypeName := "Nagare Alert Webhook"
	actionName := "Nagare Alert Webhook"
	sendTo := "nagare-webhook"

	LogSystem("debug", "creating media type", map[string]interface{}{"alarm_id": id, "webhook_url": webhookURL}, nil, "")
	if err := client.CreateMediaType(ctx, mediaTypeName, script, params); err != nil {
		LogSystem("error", "failed to create media type", map[string]interface{}{
			"alarm_id":    id,
			"error":       err.Error(),
			"webhook_url": webhookURL,
			"suggestion":  "Check if Zabbix can reach the webhook URL. Test manually: curl " + webhookURL + "/health",
		}, nil, "")
		return fmt.Errorf("failed to create media type: %w", err)
	}

	mediaTypeID, err := client.GetMediaTypeIDByName(ctx, mediaTypeName)
	if err != nil {
		LogSystem("error", "failed to resolve media type id", map[string]interface{}{
			"alarm_id":   id,
			"error":      err.Error(),
			"media_type": mediaTypeName,
		}, nil, "")
		return fmt.Errorf("failed to resolve media type id: %w", err)
	}

	userID, err := client.GetUserIDByUsername(ctx, alarm.Username)
	if err != nil {
		LogSystem("error", "failed to resolve zabbix user", map[string]interface{}{
			"alarm_id": id,
			"error":    err.Error(),
			"username": alarm.Username,
		}, nil, "")
		return fmt.Errorf("failed to resolve zabbix user: %w", err)
	}

	if err := client.EnsureUserMedia(ctx, userID, mediaTypeID, sendTo); err != nil {
		LogSystem("error", "failed to bind media type to user", map[string]interface{}{
			"alarm_id":      id,
			"error":         err.Error(),
			"user_id":       userID,
			"media_type_id": mediaTypeID,
		}, nil, "")
		return fmt.Errorf("failed to bind media type to user: %w", err)
	}

	if err := client.EnsureActionWithMedia(ctx, actionName, userID, mediaTypeID); err != nil {
		LogSystem("error", "failed to bind media type to action", map[string]interface{}{
			"alarm_id":      id,
			"error":         err.Error(),
			"action_name":   actionName,
			"user_id":       userID,
			"media_type_id": mediaTypeID,
		}, nil, "")
		return fmt.Errorf("failed to bind media type to action: %w", err)
	}

	LogSystem("info", "media type setup completed successfully", map[string]interface{}{
		"alarm_id":    id,
		"webhook_url": webhookURL,
		"test_url":    webhookURL + "/health",
	}, nil, "")
	return nil
}

// alarmToResp converts a domain Alarm to AlarmResp
func alarmToResp(a model.Alarm) AlarmResp {
	return AlarmResp{
		ID:          int(a.ID),
		Name:        a.Name,
		URL:         a.URL,
		Username:    a.Username,
		Password:    a.Password,
		AuthToken:   a.AuthToken,
		EventToken:  a.EventToken,
		Description: a.Description,
		Type:        a.Type,
		Enabled:     a.Enabled,
		Status:      a.Status,
		StatusDesc:  a.StatusDescription,
	}
}
