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
		return err
	}

	if token, ok := res.Result.(string); ok {
		ap.Token = token
		return nil
	}
	return fmt.Errorf("authentication failed")
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
	cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.radio0.channel='%d'; ", channel))

	stations := []string{"red1", "red2", "red3", "blue1", "blue2", "blue3"}
	vlans := []int{10, 20, 30, 40, 50, 60}

	for i, team := range teams {
		station := stations[i]
		vlan := vlans[i]
		if team == nil {
			cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.disabled='1'; ", station))
			continue
		}

		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.disabled='0'; ", station))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.network='vlan%d'; ", station, vlan))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.mode='ap'; ", station))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.encryption='psk2'; ", station))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.ssid='%d'; ", station, team.Id))
		cmdBuffer.WriteString(fmt.Sprintf("uci set wireless.%s.key='%s'; ", station, team.WpaKey))
	}

	cmdBuffer.WriteString("uci commit wireless; /sbin/wifi reload;")

	err := ap.runSysCommand(cmdBuffer.String())
	if err != nil {
		// Retry once with authentication
		ap.Authenticate()
		err = ap.runSysCommand(cmdBuffer.String())
		if err != nil {
			ap.Status = "ERROR"
			return err
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
