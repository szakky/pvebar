package main

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestFormatPlain(t *testing.T) {
	t.Parallel()

	got := FormatPlain(Summary{CPU: 23.4, Mem: 61.2, Running: 2, Total: 3})
	want := "PVE CPU 23% RAM 61% VM 2/3"

	if got != want {
		t.Errorf("FormatPlain() = %q, want %q", got, want)
	}
}

func TestFormatWaybarClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cpu  float64
		mem  float64
		want string
	}{
		{"ok", 10, 50, "ok"},
		{"warning by memory", 10, 85, "warning"},
		{"critical by cpu", 95, 10, "critical"},
		{"exactly at warn", 80, 0, "warning"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			line, err := FormatWaybar(Summary{CPU: tt.cpu, Mem: tt.mem}, 80, 90)
			if err != nil {
				t.Fatal(err)
			}

			var out waybarOutput
			if err := json.Unmarshal([]byte(line), &out); err != nil {
				t.Fatalf("output is not valid JSON: %v\n%s", err, line)
			}

			if out.Class != tt.want {
				t.Errorf("class = %q, want %q", out.Class, tt.want)
			}
		})
	}
}

func TestFormatWaybarTooltipListsGuests(t *testing.T) {
	t.Parallel()

	s := Summary{
		Guests: []Resource{
			{VMID: 100, Name: "k3s", Type: "qemu", Status: "running"},
			{VMID: 101, Name: "postgres", Type: "qemu", Status: "stopped"},
		},
	}

	line, err := FormatWaybar(s, 80, 90)
	if err != nil {
		t.Fatal(err)
	}

	var out waybarOutput
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{"100 k3s (qemu): running", "101 postgres (qemu): stopped"} {
		if !strings.Contains(out.Tooltip, want) {
			t.Errorf("tooltip %q does not contain %q", out.Tooltip, want)
		}
	}
}

func TestFormatWaybarError(t *testing.T) {
	t.Parallel()

	line, err := FormatWaybarError(errors.New("connection refused"))
	if err != nil {
		t.Fatal(err)
	}

	var out waybarOutput
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		t.Fatal(err)
	}

	if out.Class != "error" || out.Tooltip != "connection refused" {
		t.Errorf("unexpected error output: %+v", out)
	}
}
