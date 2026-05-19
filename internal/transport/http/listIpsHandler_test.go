package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"proxy/configGolang"
)

type mockListUseCase struct {
	ips       map[string][]string
	contains  map[string]map[string]bool
	addErr    error
	deleteErr error
}

func (m *mockListUseCase) GetAllIps(listName string) ([]string, error) {
	return m.ips[listName], nil
}
func (m *mockListUseCase) AddIp(ip, listName string) error {
	if m.addErr != nil {
		return m.addErr
	}
	m.ips[listName] = append(m.ips[listName], ip)
	return nil
}
func (m *mockListUseCase) Contains(ip, listName string) (bool, error) {
	if m.contains[listName] == nil {
		return false, nil
	}
	return m.contains[listName][ip], nil
}
func (m *mockListUseCase) DeleteIp(ip, listName string) error {
	return m.deleteErr
}
func (m *mockListUseCase) GetAllCIDRIps(listName string) ([]string, error)  { return nil, nil }
func (m *mockListUseCase) GetAllRangeIps(listName string) ([]string, error) { return nil, nil }

func TestIpHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockListUseCase{
		ips: map[string][]string{
			"white": {"1.1.1.1", "2.2.2.2"},
			"black": {"9.9.9.9"},
			"gray":  {},
		},
		contains: map[string]map[string]bool{
			"white": {"1.1.1.1": true},
			"black": {"9.9.9.9": true},
		},
	}

	cfg := &configGolang.Config{SecretKey: "supersecret"}
	handler := CreateIpHandler(mockUC, nil, cfg)

	tests := []struct {
		name      string
		method    string
		path      string
		body      interface{}
		wantCode  int
		wantItems int
	}{
		{"Get White List", "GET", "/api/white", nil, http.StatusOK, 2},
		{"Get Black List", "GET", "/api/black", nil, http.StatusOK, 1},
		{"Get Gray List", "GET", "/api/gray", nil, http.StatusOK, 0},

		{"Add IP to White", "POST", "/api/white", IpConf{Ip: "8.8.8.8"}, http.StatusOK, 0},

		{"Check IP exists", "POST", "/api/white/check", IpConf{Ip: "1.1.1.1"}, http.StatusOK, 0},
		{"Check IP not exists", "POST", "/api/white/check", IpConf{Ip: "10.0.0.1"}, http.StatusOK, 0},

		{"Delete IP", "DELETE", "/api/white", IpConf{Ip: "1.1.1.1"}, http.StatusOK, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{{Key: "list", Value: tt.path[5:]}}

			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				c.Request = httptest.NewRequest(tt.method, tt.path, bytes.NewBuffer(bodyBytes))
				c.Request.Header.Set("Content-Type", "application/json")
			} else {
				c.Request = httptest.NewRequest(tt.method, tt.path, nil)
			}

			switch tt.method {
			case "GET":
				handler.GetWhiteIps(c)
			case "POST":
				if tt.path == "/api/white/check" {
					handler.CheckIp(c)
				} else {
					handler.NewIpInWhiteList(c)
				}
			case "DELETE":
				handler.DeleteIpFromWhiteList(c)
			}

			if w.Code != tt.wantCode {
				t.Errorf("expected status %d, got %d", tt.wantCode, w.Code)
			}
		})
	}
}

func TestIpHandler_Health(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &IpHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	handler.Health(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestIpHandler_NewIpInWhiteList_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockListUseCase{
		addErr: fmt.Errorf("duplicate ip"),
		ips:    make(map[string][]string),
	}
	handler := CreateIpHandler(mockUC, nil, &configGolang.Config{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "list", Value: "white"}}

	body, _ := json.Marshal(IpConf{Ip: "1.1.1.1"})
	c.Request = httptest.NewRequest("POST", "/api/white", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.NewIpInWhiteList(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on error, got %d", w.Code)
	}
}
