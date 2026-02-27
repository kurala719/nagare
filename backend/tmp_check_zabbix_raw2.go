package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// Let's manually plug the default Zabbix credentials if we know them
	apiURL := "http://localhost:8080/api_jsonrpc.php"
	username := "Admin"
	password := "zabbix"

	// Assuming hostid 10084 or similar exists
	testHostID := "10084"

	// Auth
	authPayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "user.login",
		"params": map[string]string{
			"username": username,
			"password": password,
		},
		"id": 1,
	}
	authData, _ := json.Marshal(authPayload)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(authData))
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

	if authResult.Result == "" {
		fmt.Println("Auth failed or Zabbix not running")
		return
	}

	// Try to get hosts first to find a valid hostid
	getPayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "host.get",
		"params": map[string]string{
			"output": "extend",
		},
		"auth": authResult.Result,
		"id":   2,
	}
	getData, _ := json.Marshal(getPayload)
	reqGet, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(getData))
	reqGet.Header.Set("Content-Type", "application/json-rpc")
	respGet, _ := client.Do(reqGet)
	getBody, _ := ioutil.ReadAll(respGet.Body)

	var getResult struct {
		Result []struct {
			Hostid string `json:"hostid"`
			Host   string `json:"host"`
		} `json:"result"`
	}
	json.Unmarshal(getBody, &getResult)

	if len(getResult.Result) > 0 {
		testHostID = getResult.Result[0].Hostid
		fmt.Printf("Found a valid host to try deleting: %s (ID %s)\n", getResult.Result[0].Host, testHostID)
	} else {
		fmt.Println("No hosts found on Zabbix to test delete.")
		return
	}

	// Try to delete Host
	delPayload := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "host.delete",
		"params":  []string{testHostID},
		"auth":    authResult.Result,
		"id":      3,
	}
	delData, _ := json.Marshal(delPayload)
	req2, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(delData))
	req2.Header.Set("Content-Type", "application/json-rpc")
	resp2, _ := client.Do(req2)
	defer resp2.Body.Close()
	delBody, _ := ioutil.ReadAll(resp2.Body)
	fmt.Println("Delete Response Data:\n", string(delBody))
}
