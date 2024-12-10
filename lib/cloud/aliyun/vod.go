package aliyun

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	VodURL         = "https://vod.%s.aliyuncs.com"
	DefaultRegion  = "cn-shanghai"
	DefaultVersion = "2017-03-21"
)

type PlayInfoRequest struct {
	AccessKeyId   string   `json:"AccessKeyId"`
	Action        string   `json:"Action"`
	AuthInfo      string   `json:"AuthInfo"`
	AuthTimeout   int      `json:"AuthTimeout"`
	Channel       string   `json:"Channel"`
	Definition    string   `json:"Definition"`
	ResultType    string   `json:"ResultType"`
	Formats       string   `json:"Formats"`
	PlayConfig    struct{} `json:"PlayConfig"`
	PlayerVersion string   `json:"PlayerVersion"`
	Rand          string   `json:"Rand"`
	ReAuthInfo    struct{} `json:"ReAuthInfo"`
	SecurityToken string   `json:"SecurityToken"`
	StreamType    string   `json:"StreamType"`
	VideoId       string   `json:"VideoId"`
}

type PlayInfoResponse struct {
	VideoBase    VideoBase    `json:"VideoBase,omitempty"`
	RequestID    string       `json:"RequestId,omitempty"`
	PlayInfoList PlayInfoList `json:"PlayInfoList,omitempty"`
}

type VideoBase struct {
	Status        string    `json:"Status,omitempty"`
	VideoID       string    `json:"VideoId,omitempty"`
	StorageClass  string    `json:"StorageClass,omitempty"`
	TranscodeMode string    `json:"TranscodeMode,omitempty"`
	CreationTime  time.Time `json:"CreationTime,omitempty"`
	Title         string    `json:"Title,omitempty"`
	MediaType     string    `json:"MediaType,omitempty"`
	Duration      string    `json:"Duration,omitempty"`
	CoverURL      string    `json:"CoverURL,omitempty"`
	OutputType    string    `json:"OutputType,omitempty"`
}

type PlayInfo struct {
	Status           string    `json:"Status,omitempty"`
	StreamType       string    `json:"StreamType,omitempty"`
	Rand             string    `json:"Rand,omitempty"`
	Size             int       `json:"Size,omitempty"`
	Definition       string    `json:"Definition,omitempty"`
	Fps              string    `json:"Fps,omitempty"`
	Specification    string    `json:"Specification,omitempty"`
	ModificationTime time.Time `json:"ModificationTime,omitempty"`
	Duration         string    `json:"Duration,omitempty"`
	Bitrate          string    `json:"Bitrate,omitempty"`
	Encrypt          int       `json:"Encrypt,omitempty"`
	PreprocessStatus string    `json:"PreprocessStatus,omitempty"`
	Format           string    `json:"Format,omitempty"`
	EncryptType      string    `json:"EncryptType,omitempty"`
	NarrowBandType   string    `json:"NarrowBandType,omitempty"`
	PlayURL          string    `json:"PlayURL,omitempty"`
	CreationTime     time.Time `json:"CreationTime,omitempty"`
	Plaintext        string    `json:"Plaintext,omitempty"`
	Height           int       `json:"Height,omitempty"`
	Width            int       `json:"Width,omitempty"`
	JobID            string    `json:"JobId,omitempty"`
}

type PlayInfoList struct {
	PlayInfo []PlayInfo `json:"PlayInfo,omitempty"`
}

func GetPlayInfo(req PlayInfoRequest, accessKeySecret, regionId string) (*PlayInfoResponse, error) {
	if len(regionId) == 0 {
		regionId = DefaultRegion
	}
	baseURI := fmt.Sprintf(VodURL, regionId)
	req.Action = "GetPlayInfo"
	request, err := WrapRequest(http.MethodGet, baseURI,
		req, DefaultVersion, regionId, req.AccessKeyId, accessKeySecret, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("status code: %d, %s", response.StatusCode, string(raw))
	}
	defer func() { _ = response.Body.Close() }()
	var resp PlayInfoResponse
	if err = json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}
