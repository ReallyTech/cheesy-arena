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

const (
	// FRC Network Settings
	switchTeamGatewaySuffix = 4 // OpenWrt 習慣使用 .1 作為 Gateway (Cisco code是用 .4)
)

type WrtSwitch struct {
	User     string
	Password string
	Token    string // 儲存 RPC Session Token
	client   *http.Client
	mutex    sync.Mutex
	Status   string
	Address  string
	baseURL  string
}

func (sw *WrtSwitch) GetStatus() string {
	return sw.Status
}

// JSON-RPC 請求結構
type JsonRpcRequest struct {
	Id     int           `json:"id"`
	Method string        `json:"method"`
	Params []interface{} `json:"params"`
}

// JSON-RPC 回應結構
type JsonRpcResponse struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
}

func NewWrtSwitch(address, rpcURL, password string) *WrtSwitch {
	return &WrtSwitch{
		User:     "root",
		Password: password,
		client:   &http.Client{Timeout: 10 * time.Second},
		Status:   "UNKNOWN",
		Address:  address,
		baseURL:  fmt.Sprintf("http://%s%s", address, rpcURL),
	}
}

// 驗證並取得 Token
func (sw *WrtSwitch) Authenticate() error {
	reqBody := JsonRpcRequest{
		Id:     1,
		Method: "login",
		Params: []interface{}{sw.User, sw.Password},
	}
	res, err := sw.sendRequest(sw.baseURL+"/auth", reqBody)
	if err != nil {
		return err
	}

	if token, ok := res.Result.(string); ok {
		sw.Token = token
		return nil
	}
	return fmt.Errorf("authentication failed")
}

// 核心功能：設定隊伍 VLAN IP 並啟用 DHCP
func (sw *WrtSwitch) ConfigureTeamEthernet(teams [6]*model.Team) error {
	sw.mutex.Lock()
	defer sw.mutex.Unlock()
	sw.Status = "CONFIGURING"

	// 確保有 Token，若無則登入
	if sw.Token == "" {
		if err := sw.Authenticate(); err != nil {
			sw.Status = "AUTH_ERROR"
			return err
		}
	}

	var cmdBuffer bytes.Buffer
	vlanMap := []int{10, 20, 30, 40, 50, 60} // Red1-3, Blue1-3

	for i, team := range teams {
		vlanId := vlanMap[i]
		if team == nil {
			// 若該位置無隊伍，將 protocol 設為 none 並停用 DHCP
			cmdBuffer.WriteString(fmt.Sprintf("uci set network.vlan%d.proto='none'; ", vlanId))
			cmdBuffer.WriteString(fmt.Sprintf("uci set dhcp.vlan%d.ignore='1'; ", vlanId))
			continue
		}

		// 計算 FRC IP: 10.TE.AM.xx
		teamIpPart := fmt.Sprintf("%d.%d", team.Id/100, team.Id%100)
		gatewayIp := fmt.Sprintf("10.%s.%d", teamIpPart, switchTeamGatewaySuffix)

		// 產生 UCI 指令
		// 1. 設定 Interface 為 static 並指定 IP
		cmdBuffer.WriteString(fmt.Sprintf("uci set network.vlan%d.proto='static'; ", vlanId))
		cmdBuffer.WriteString(fmt.Sprintf("uci set network.vlan%d.ipaddr='%s'; ", vlanId, gatewayIp))
		cmdBuffer.WriteString(fmt.Sprintf("uci set network.vlan%d.netmask='255.255.255.0'; ", vlanId))

		// 2. 啟用該介面上的 DHCP (確保 ignore 為 0 並設定範圍)
		cmdBuffer.WriteString(fmt.Sprintf("uci set dhcp.vlan%d='dhcp'; ", vlanId))
		cmdBuffer.WriteString(fmt.Sprintf("uci set dhcp.vlan%d.interface='vlan%d'; ", vlanId, vlanId))
		cmdBuffer.WriteString(fmt.Sprintf("uci set dhcp.vlan%d.ignore='0'; ", vlanId))
		cmdBuffer.WriteString(fmt.Sprintf("uci set dhcp.vlan%d.start='20'; ", vlanId))
		cmdBuffer.WriteString(fmt.Sprintf("uci set dhcp.vlan%d.limit='180'; ", vlanId))
		cmdBuffer.WriteString(fmt.Sprintf("uci set dhcp.vlan%d.leasetime='1h'; ", vlanId))
	}

	// 3. Commit 與 Reload
	cmdBuffer.WriteString("uci commit network; uci commit dhcp; ")
	for _, vlanId := range vlanMap {
		cmdBuffer.WriteString(fmt.Sprintf("ifup vlan%d; ", vlanId))
	}
	cmdBuffer.WriteString("/etc/init.d/dnsmasq reload;")

	// 執行指令
	err := sw.runSysCommand(cmdBuffer.String())
	if err != nil {
		// Token 可能過期，重試一次
		sw.Authenticate()
		err = sw.runSysCommand(cmdBuffer.String())
		if err != nil {
			sw.Status = "ERROR"
			return err
		}
	}

	sw.Status = "ACTIVE"
	return nil
}

// 透過 RPC sys 模組執行 Shell 指令
func (sw *WrtSwitch) runSysCommand(command string) error {
	reqBody := JsonRpcRequest{
		Id:     2,
		Method: "exec", // 呼叫 sys.exec
		Params: []interface{}{command},
	}

	// 需將 Token 帶入 URL Query
	url := fmt.Sprintf("%s/sys?auth=%s", sw.baseURL, sw.Token)
	_, err := sw.sendRequest(url, reqBody)
	return err
}

// HTTP Request Helper
func (sw *WrtSwitch) sendRequest(url string, reqBody JsonRpcRequest) (*JsonRpcResponse, error) {
	jsonBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := sw.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	var rpcResp JsonRpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %v (check if RPC URL is correct)", err)
	}

	if rpcResp.Error != nil {
		return &rpcResp, fmt.Errorf("rpc error: %v", rpcResp.Error)
	}

	return &rpcResp, nil
}
