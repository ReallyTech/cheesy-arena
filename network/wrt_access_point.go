// Copyright 2024 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Methods for configuring a standard OpenWRT access point via JSON-RPC.

package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Team254/cheesy-arena/model"
)

type WrtAccessPoint struct {
	User     string
	Password string
	Token    string
	client   *http.Client
	mutex    sync.Mutex
	Status   string
	Address  string
	baseURL  string
}

func NewWrtAccessPoint(address, rpcURL, password string) *WrtAccessPoint {
	return &WrtAccessPoint{
		User:     "root",
		Password: password,
		client:   &http.Client{Timeout: 10 * time.Second},
		Status:   "UNKNOWN",
		Address:  address,
		baseURL:  fmt.Sprintf("http://%s%s", address, rpcURL),
	}
}

func (ap *WrtAccessPoint) Authenticate() error {
	reqBody := JsonRpcRequest{
		Id:     1,
		Method: "login",
		Params: []interface{}{ap.User, ap.Password},
	}
	res, err := ap.sendRequest(ap.baseURL+"/auth", reqBody)
	if err != nil {
		ap.Token = ""
		return err
	}

	if token, ok := res.Result.(string); ok {
		ap.Token = token
		return nil
	}
	ap.Token = ""
	return fmt.Errorf("authentication failed: result is not a string")
}

func (ap *WrtAccessPoint) ConfigureTeamWifi(teams [6]*model.Team, channel int) error {
	ap.mutex.Lock()
	defer ap.mutex.Unlock()
	ap.Status = "CONFIGURING"

	if ap.Token == "" {
		if err := ap.Authenticate(); err != nil {
			ap.Status = "AUTH_ERROR"
			return err
		}
	}

	var cmdBuffer bytes.Buffer
	// Set channel
	cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.radio1.channel='%d'; ", channel))

	stations := []string{"wifinet1", "wifinet2", "wifinet3", "wifinet4", "wifinet5", "wifinet6"}
	vlans := []int{10, 20, 30, 40, 50, 60}

	for i, team := range teams {
		station := stations[i]
		vlan := vlans[i]
		if team == nil {
			cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.disabled='1'; ", station))
			continue
		}

		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s='wifi-iface'; ", station))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.device='radio1'; ", station))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.disabled='0'; ", station))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.network='vlan%d'; ", station, vlan))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.mode='ap'; ", station))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.encryption='psk2'; ", station))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.ssid='%d-ds'; ", station, team.Id))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.key='%s'; ", station, team.WpaKey))
	}

	cmdBuffer.WriteString("uci commit wireless; /sbin/wifi reload;")

	err := ap.runSysCommand(cmdBuffer.String())
	if err != nil {
		// Retry once with authentication in case the token expired
		if authErr := ap.Authenticate(); authErr != nil {
			ap.Status = "ERROR"
			return fmt.Errorf("initial command failed (%v), and re-authentication failed (%v)", err, authErr)
		}
		err = ap.runSysCommand(cmdBuffer.String())
		if err != nil {
			ap.Status = "ERROR"
			return fmt.Errorf("command failed after re-authentication: %v", err)
		}
	}

	ap.Status = "ACTIVE"
	return nil
}

func (ap *WrtAccessPoint) runSysCommand(command string) error {
	reqBody := JsonRpcRequest{
		Id:     2,
		Method: "exec",
		Params: []interface{}{command},
	}

	url := fmt.Sprintf("%s/sys?auth=%s", ap.baseURL, ap.Token)
	_, err := ap.sendRequest(url, reqBody)
	return err
}

func (ap *WrtAccessPoint) sendRequest(url string, reqBody JsonRpcRequest) (*JsonRpcResponse, error) {
	jsonBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := ap.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	var rpcResp JsonRpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v", err)
	}

	if rpcResp.Error != nil {
		return &rpcResp, fmt.Errorf("rpc error: %v", rpcResp.Error)
	}

	return &rpcResp, nil
}
