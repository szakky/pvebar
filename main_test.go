package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// writeConfig writes a config file pointing at endpoint and returns its path.
func writeConfig(t *testing.T, endpoint string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")

	data, err := json.Marshal(Config{Endpoint: endpoint, Token: "pvebar@pve!bar=secret", Insecure: true})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestRunJSON(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":[
			{"type":"node","node":"pve","cpu":0.25,"maxcpu":4,"mem":1024,"maxmem":4096},
			{"type":"qemu","vmid":100,"name":"k3s","status":"running","node":"pve"}
		]}`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer

	code := run([]string{"--format", "json", "--config", writeConfig(t, srv.URL)}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %s", code, stderr.String())
	}

	var out jsonOutput
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}

	if !out.OK || out.CPU != 25 || out.Mem != 25 || out.Running != 1 || out.Total != 1 {
		t.Errorf("unexpected output: %+v", out)
	}
}

func TestRunJSONErrorExitsZero(t *testing.T) {
	t.Parallel()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "authentication failed", http.StatusUnauthorized)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer

	// Exit 0 so the widget reads stdout and shows the error instead of giving up.
	code := run([]string{"--format", "json", "--config", writeConfig(t, srv.URL)}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	var out jsonError
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout.String())
	}

	if out.OK || out.Error == "" {
		t.Errorf("want ok=false with an error message, got %+v", out)
	}

	if stderr.Len() == 0 {
		t.Error("details should also go to stderr")
	}
}

func TestRunUnknownFormat(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	if code := run([]string{"--format", "xml"}, &stdout, &stderr); code != 2 {
		t.Errorf("exit code = %d, want 2", code)
	}
}
