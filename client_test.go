package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// These tests use a local fake server, so they check pvebar's own HTTP and JSON
// handling only. Try pvebar against a real Proxmox VE before a release.

func TestClientResources(t *testing.T) {
	t.Parallel()

	const token = "pvebar@pve!bar=secret"

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api2/json/cluster/resources" {
			http.NotFound(w, r)
			return
		}

		if got := r.Header.Get("Authorization"); got != "PVEAPIToken="+token {
			http.Error(w, "bad token", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[
			{"type":"node","node":"pve","cpu":0.1,"maxcpu":8,"mem":1024,"maxmem":4096,"status":"online"},
			{"type":"qemu","vmid":100,"name":"k3s","status":"running","node":"pve"}
		]}`))
	}))
	defer srv.Close()

	// insecure=true because httptest uses a self-signed certificate, like Proxmox VE.
	client := NewClient(srv.URL+"/", token, true, 2*time.Second)

	resources, err := client.Resources(context.Background())
	if err != nil {
		t.Fatalf("Resources() error = %v", err)
	}

	if len(resources) != 2 {
		t.Fatalf("got %d resources, want 2", len(resources))
	}

	if resources[1].Name != "k3s" || resources[1].VMID != 100 {
		t.Errorf("unexpected guest: %+v", resources[1])
	}
}

func TestClientResourcesHTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "permission denied", http.StatusForbidden)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "x@pve!y=z", true, 2*time.Second)

	if _, err := client.Resources(context.Background()); err == nil {
		t.Fatal("expected an error for HTTP 403")
	}
}

func TestClientRejectsSelfSignedWithoutInsecure(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "x@pve!y=z", false, 2*time.Second)

	if _, err := client.Resources(context.Background()); err == nil {
		t.Fatal("expected a TLS error when insecure is false")
	}
}
