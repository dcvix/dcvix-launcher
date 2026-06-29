//  SPDX-FileCopyrightText: 2026 Diego Cortassa
//  SPDX-License-Identifier: MIT

// Package client provides the HTTP client logic to connect to the dcv broker,
// authenticate users, list available servers, and manage DCV sessions.
package client

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/tidwall/gjson"
)

const HTTPTimeout = 30 * time.Second

type APIClient struct {
	brokerURL       string
	httpClient      *http.Client
	userToken       string
	connectionToken string
	mu              sync.RWMutex
}

// Close closes idle HTTP connections held by the underlying transport.
func (c *APIClient) Close() {
	if tr, ok := c.httpClient.Transport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
}

// NewAPIClient creates a new APIClient with the given broker URL and TLS configuration.
func NewAPIClient(broker string, acceptUntrustedCert bool) (*APIClient, error) {
	u, err := url.Parse(broker)
	if err != nil {
		return nil, fmt.Errorf("invalid broker URL: %w", err)
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("broker URL must use https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("broker URL must have a host")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS13,
			InsecureSkipVerify: acceptUntrustedCert,
		},
		DialContext: (&net.Dialer{
			Timeout: HTTPTimeout,
		}).DialContext,
	}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   HTTPTimeout,
	}

	return &APIClient{
		brokerURL:  u.String(),
		httpClient: httpClient,
	}, nil
}

// ConnectionToken returns the current connection token, thread-safe.
func (c *APIClient) ConnectionToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connectionToken
}

func (c *APIClient) setConnectionToken(t string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connectionToken = t
}

// UserToken returns the current user session token, thread-safe.
func (c *APIClient) UserToken() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userToken
}

func (c *APIClient) setUserToken(t string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.userToken = t
}

type loginData struct {
	UserID   string `json:"userID"`
	Password string `json:"password"`
	OTP      string `json:"otp"`
}

type createSessionData struct {
	ServerID    string `json:"serverId"`
	UserID      string `json:"userId"`
	SessionType string `json:"sessionType"`
}

func (c *APIClient) getURL(reqURL string) (string, error) {
	tokenStr := c.UserToken()
	if tokenStr == "" {
		tokenStr = c.ConnectionToken()
	}

	log.Debugf("getUrl: Calling %s token: %s", sanitizeURL(reqURL), truncateToken(tokenStr))

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		log.Errorf("getUrl: Can't create request: %s", err)
		return "", err
	}

	if tokenStr != "" {
		sessionCookie := createCookie(tokenStr)
		req.AddCookie(sessionCookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Errorf("getUrl: Can't get url: %s", err)
		return "", err
	}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		log.Errorf("getUrl: Can't read response body: %s", readErr)
		return "", readErr
	}

	if resp.StatusCode != 200 {
		log.Errorf("getUrl: Status code: %d", resp.StatusCode)
		return "", fmt.Errorf("status code %d:\n%s", resp.StatusCode, string(body))
	}

	return string(body), nil
}

func (c *APIClient) postURL(reqURL, data string) (string, error) {
	resp, err := c.doPost(reqURL, data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("postUrl: Can't read response body: %s", err)
		return "", err
	}

	if resp.StatusCode != 200 {
		log.Errorf("postUrl: Status code: %d", resp.StatusCode)
		return "", fmt.Errorf("status code %d:\n%s", resp.StatusCode, string(body))
	}

	log.Debugf("postUrl: response: %s", string(body))
	return string(body), nil
}

func (c *APIClient) doPost(reqURL, data string) (*http.Response, error) {
	tokenStr := c.UserToken()
	if tokenStr == "" {
		tokenStr = c.ConnectionToken()
	}

	log.Debugf("doPost: Calling %s token: %s", sanitizeURL(reqURL), truncateToken(tokenStr))

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(data))
	if err != nil {
		log.Errorf("doPost: Can't create request: %s", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if tokenStr != "" {
		sessionCookie := createCookie(tokenStr)
		req.AddCookie(sessionCookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Errorf("doPost error: %s", err)
		return nil, err
	}

	return resp, nil
}

func getCookie(resp *http.Response, name string) *http.Cookie {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	return nil
}

func createCookie(tokenStr string) *http.Cookie {
	return &http.Cookie{
		Name:  "dcvix_session",
		Value: tokenStr,
		Path:  "/",
	}
}

// Login authenticates a user with the broker using password and optional OTP.
func (c *APIClient) Login(userID, password, otp string) error {
	reqURL, err := url.JoinPath(c.brokerURL, "/v1/user/login")
	if err != nil {
		return err
	}

	log.Debugf("Login: Calling %s", sanitizeURL(reqURL))

	data := loginData{
		UserID:   userID,
		Password: password,
		OTP:      otp,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Errorf("Login: Can't marshal json: %s", err)
		return err
	}

	resp, err := c.doPost(reqURL, string(jsonData))
	if err != nil {
		log.Errorf("Login error: %s", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Login: Can't read response body: %s", err)
		return err
	}

	if resp.StatusCode != 200 {
		log.Errorf("Login: Status code: %d", resp.StatusCode)
		if len(body) > 0 {
			return fmt.Errorf("login failed: %s", strings.TrimSpace(string(body)))
		}
		return fmt.Errorf("login failed: status code %d", resp.StatusCode)
	}

	if sessionCookie := getCookie(resp, "dcvix_session"); sessionCookie != nil {
		log.Infof("Session: %s", truncateToken(sessionCookie.Value))
		c.setUserToken(sessionCookie.Value)
	} else {
		return errors.New("login: server did not return a session cookie")
	}

	log.Debugf("Login: response: %s", string(body))
	return nil
}

// ConfigEntry represents a single configuration parameter the broker applies to a DCV server.
// Its shape is dictated by the broker's SetConfig JSON contract.
// The broker expects Section/Key/Value triplets.
type ConfigEntry struct {
	Section string `json:"section"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

// SetConfig sets configuration entries for the given server.
func (c *APIClient) SetConfig(server string, config []ConfigEntry) error {
	reqURL, err := url.JoinPath(c.brokerURL, "/v1/user/servers", server, "/config")
	if err != nil {
		return err
	}
	log.Debugf("SetConfig: Calling %s", sanitizeURL(reqURL))
	data := struct {
		Config []ConfigEntry `json:"config"`
	}{
		Config: config,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Errorf("SetConfig: Can't marshal json: %s", err)
		return err
	}
	var resp string
	resp, err = c.postURL(reqURL, string(jsonData))
	if err != nil {
		log.Errorf("SetConfig error: %s", err)
		return err
	}
	log.Debugf("SetConfig: response: %s", resp)
	return nil
}

// ListServers returns the list of server names available to the given user.
func (c *APIClient) ListServers() ([]string, error) {
	reqURL, err := url.JoinPath(c.brokerURL, "/v1/user/servers")
	if err != nil {
		return nil, err
	}
	log.Debugf("ListServers: Calling %s", sanitizeURL(reqURL))
	resp, err := c.getURL(reqURL)
	if err != nil {
		log.Errorf("ListServers: Can't get url: %s", err)
		return nil, err
	}
	log.Debugf("ListServers: response: %s", resp)
	servers := gjson.Get(resp, "servers").Array()
	serverList := make([]string, len(servers))
	for i, s := range servers {
		serverList[i] = s.String()
	}
	return serverList, nil
}

// CreateSession requests a new DCV session on the given server and returns the session ID.
func (c *APIClient) CreateSession(serverID, userID, sessionType string) (string, error) {
	data := createSessionData{
		ServerID:    serverID,
		UserID:      userID,
		SessionType: sessionType,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Errorf("CreateSession: Can't marshal json: %s", err)
		return "", err
	}
	reqURL, err := url.JoinPath(c.brokerURL, "/v1/user/sessions")
	if err != nil {
		return "", err
	}

	log.Debugf("CreateSession: Calling %s", sanitizeURL(reqURL))
	resp, err := c.postURL(reqURL, string(jsonData))
	if err != nil {
		return "", err
	}
	result := gjson.Parse(resp)
	failureReason := result.Get("unsuccessful_list.0.failure_reason").String()
	if failureReason != "" {
		log.Errorf("CreateSession: %s", failureReason)
		return "", errors.New(failureReason)
	}
	sessionID := result.Get("sessionID").String()
	if sessionID == "" {
		log.Errorf("CreateSession: no sessionID in response")
		return "", errors.New("no sessionID in response")
	}
	return sessionID, nil
}

// GetConnectionToken retrieves a connection token for a session and stores it in the client.
func (c *APIClient) GetConnectionToken(server, userID, sessionID string) error {
	vals := url.Values{}
	vals.Add("userId", userID)
	vals.Add("sessionId", sessionID)
	vals.Add("serverId", server)
	reqURL, err := url.JoinPath(c.brokerURL, "/v1/user/connectiontoken")
	if err != nil {
		return err
	}
	reqURL += "?" + vals.Encode()

	log.Debugf("GetConnectionToken: Calling %s", sanitizeURL(reqURL))
	resp, err := c.getURL(reqURL)
	if err != nil {
		log.Errorf("GetConnectionToken: Can't get url: %s", err)
		return err
	}
	log.Debugf("GetConnectionToken: response: %s", resp)
	c.setConnectionToken(gjson.Get(resp, "token").String())
	return nil
}
