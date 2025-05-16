package ocr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestBaiduOCRProvider_getAccessToken(t *testing.T) {
	tests := []struct {
		name        string
		APIEndpoint string
		TokenURL    string
		HTTPClient  *http.Client
		APIKey      string
		SecretKey   string
		AccessToken string
		ExpiresAt   time.Time
		want        string
		wantErr     bool
	}{
		// TODO: Add test cases.
		{
			name:        "test",
			APIEndpoint: "https://aip.baidubce.com/rest/2.0/ocr/v1/general_basic",
			TokenURL:    "https://aip.baidubce.com/oauth/2.0/token",
			HTTPClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &BaiduOCRProvider{
				HTTPClient:  tt.HTTPClient,
				APIKey:      tt.APIKey,
				SecretKey:   tt.SecretKey,
				APIEndpoint: tt.APIEndpoint,
				TokenURL:    tt.TokenURL,
				AccessToken: tt.AccessToken,
				ExpiresAt:   tt.ExpiresAt,
			}
			got, err := p.getAccessToken()
			if (err != nil) != tt.wantErr {
				t.Errorf("getAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getAccessToken() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test(t *testing.T) {
	// 构建请求参数
	u := fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?client_id=%s&client_secret=%s&grant_type=client_credentials", "PB8ldlCbraEHjV6zhtKI0tqg", "TVTwbJCrDYWXoaNt36ykLYWWXSi5FxTU")
	payload := strings.NewReader(``)
	client := &http.Client{}
	req, err := http.NewRequest("POST", u, payload)
	if err != nil {
		panic(err)
	}

	// 设置请求头
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	// 解析响应
	var tokenResp BaiduTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		panic(err)
	}

	// 检查错误
	if tokenResp.Error != "" {
		panic(err)
	}
	fmt.Println(tokenResp.AccessToken)
}
