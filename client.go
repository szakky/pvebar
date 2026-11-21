package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Resource is one entry returned by GET /api2/json/cluster/resources.
// The same endpoint returns nodes, guests (qemu, lxc) and storage,
// so a single request is enough for pvebar.
type Resource struct {
	Type     string  `json:"type"` // "node", "qemu", "lxc", "storage", ...
	Node     string  `json:"node"`
	Name     string  `json:"name"`
	VMID     int     `json:"vmid"`
	Status   string  `json:"status"`
	CPU      float64 `json:"cpu"` // usage, 0.0 - 1.0
	MaxCPU   float64 `json:"maxcpu"`
	Mem      int64   `json:"mem"`
	MaxMem   int64   `json:"maxmem"`
	Template int     `json:"template"` // 1 for templates
}

// Client talks to the Proxmox VE API.
type Client struct {
	endpoint string
	token    string
	http     *http.Client
}

// NewClient returns a Client for the given endpoint and API token.
func NewClient(endpoint, token string, insecure bool, timeout time.Duration) *Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		// Proxmox VE uses a self-signed certificate by default.
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // opt-in via config
	}

	return &Client{
		endpoint: strings.TrimRight(endpoint, "/"),
		token:    token,
		http:     &http.Client{Transport: transport, Timeout: timeout},
	}
}

// Resources returns all cluster resources visible to the token.
func (c *Client) Resources(ctx context.Context) ([]Resource, error) {
	url := c.endpoint + "/api2/json/cluster/resources"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	req.Header.Set("Authorization", "PVEAPIToken="+c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling Proxmox API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("proxmox API returned %s", resp.Status)
	}

	var body struct {
		Data []Resource `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return body.Data, nil
}
