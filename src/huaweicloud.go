package dnscli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

const (
	huaweicloudEndpoint string = "https://dns.myhuaweicloud.com"
)

// Huaweicloud Huaweicloud
type HuaweicloudProvider struct {
	ID     string
	Secret string
}

// HuaweicloudZonesResp zones response
type HuaweicloudZonesResp struct {
	Zones []struct {
		ID         string
		Name       string
		Recordsets []HuaweicloudRecordsets
	}
}

// HuaweicloudRecordsResp 记录返回结果
type HuaweicloudRecordsResp struct {
	Recordsets []HuaweicloudRecordsets
}

// HuaweicloudRecordsets 记录
type HuaweicloudRecordsets struct {
	ID      string
	Name    string `json:"name"`
	ZoneID  string `json:"zone_id"`
	Status  string
	Type    string   `json:"type"`
	Records []string `json:"records"`
}

func (hw *HuaweicloudProvider) request(method string, url string, data interface{}) (response *http.Response, err error) {
	jsonStr := make([]byte, 0)
	if data != nil {
		jsonStr, _ = json.Marshal(data)
	}
	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		return nil, err
	}
	s := Signer{
		Key:    hw.ID,
		Secret: hw.Secret,
	}
	s.Sign(req)
	req.Header.Add("content-type", "application/json")
	clt := http.Client{}
	clt.Timeout = 30 * time.Second
	resp, err := clt.Do(req)
	return resp, err
}

func NewHuaweiProvider(info map[string]string) DNSProvider {

	return nil
}
