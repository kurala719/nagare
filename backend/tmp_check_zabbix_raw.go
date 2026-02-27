package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"nagare/internal/database"
	"nagare/internal/repository"
)

// A manual Zabbix JSON-RPC script to pinpoint the issue
func main() {
	os.Setenv("NAGARE_CONFIG_PATH", "../configs/nagare_config.json")
	if err := database.InitDBFromConfig(); err != nil {
		fmt.Println("DB Init:", err)
		return
	}

	hosts, _ := repository.GetAllHostsDAO()
	if len(hosts) == 0 {
		return
	}
	target := hosts[0]
	fmt.Println("Trying to delete Host ID:", target.Hostid, "(Name:", target.Name, ")")

	// 1. Get monitor
	var monitor struct {
		APIURL   string `gorm:"column:api_url"`
		Username string `gorm:"column:username"`
		Password string `gorm:"column:password"`
	}
	database.DB.Table("monitors").Where("id = ?", target.MonitorID).First(&monitor)

	// 2. Auth to Zabbix
	authPayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "user.login",
		"params": map[string]string{
			"username": monitor.Username,
			"password": monitor.Password,
		},
		"id": 1,
	}
	authData, _ := json.Marshal(authPayload)
	req, _ := http.NewRequest("POST", monitor.APIURL, bytes.NewBuffer(authData))
	req.Header.Set("Content-Type", "application/json-rpc")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Auth Req failed", err)
		return
	}
	defer resp.Body.Close()
	authBody, _ := ioutil.ReadAll(resp.Body)

	var authResult struct {
		Result string `json:"result"`
	}
	json.Unmarshal(authBody, &authResult)
	fmt.Println("Auth Token:", authResult.Result)

	// 3. Delete Host
	delPayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "host.delete",
		"params":  []string{target.Hostid},
		"auth":    authResult.Result,
		"id":      2,
	}
	delData, _ := json.Marshal(delPayload)
	req2, _ := http.NewRequest("POST", monitor.APIURL, bytes.NewBuffer(delData))
	req2.Header.Set("Content-Type", "application/json-rpc")
	resp2, _ := client.Do(req2)
	defer resp2.Body.Close()
	delBody, _ := ioutil.ReadAll(resp2.Body)
	fmt.Println("Delete Response Data:", string(delBody))
}
