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

func TestFormatJSON(t *testing.T) {
	t.Parallel()

	s := Summary{
		CPU: 23.456, Mem: 61.04, MemUsed: 2 << 30, MemTotal: 8 << 30, Running: 1, Total: 2,
		Guests: []Resource{
			{VMID: 100, Name: "k3s", Type: "qemu", Node: "pve", Status: "running", CPU: 0.1234, Mem: 1 << 30, MaxMem: 4 << 30},
			{VMID: 200, Name: "test", Type: "lxc", Node: "pve", Status: "stopped"},
		},
	}

	line, err := FormatJSON(s)
	if err != nil {
		t.Fatal(err)
	}

	var out jsonOutput
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, line)
	}

	if !out.OK || out.Version != jsonVersion {
		t.Errorf("ok/version = %v/%d, want true/%d", out.OK, out.Version, jsonVersion)
	}

	if out.CPU != 23.5 || out.Mem != 61 {
		t.Errorf("cpu/mem = %v/%v, want 23.5/61 (rounded to 0.1)", out.CPU, out.Mem)
	}

	if out.MemUsed != 2<<30 || out.MemTotal != 8<<30 || out.Running != 1 || out.Total != 2 {
		t.Errorf("unexpected totals: %+v", out)
	}

	if len(out.Guests) != 2 {
		t.Fatalf("got %d guests, want 2", len(out.Guests))
	}

	g := out.Guests[0]
	if g.VMID != 100 || g.Name != "k3s" || g.Type != "qemu" || g.Node != "pve" || g.Status != "running" {
		t.Errorf("unexpected guest: %+v", g)
	}

	// The API reports guest CPU as 0.0 - 1.0; JSON uses percent like the totals.
	if g.CPU != 12.3 || g.Mem != 1<<30 || g.MaxMem != 4<<30 {
		t.Errorf("guest cpu/mem/maxmem = %v/%d/%d", g.CPU, g.Mem, g.MaxMem)
	}
}

func TestFormatJSONNoGuestsIsEmptyArray(t *testing.T) {
	t.Parallel()

	line, err := FormatJSON(Summary{})
	if err != nil {
		t.Fatal(err)
	}

	// Widgets iterate over guests, so it must be [] and not null.
	if !strings.Contains(line, `"guests":[]`) {
		t.Errorf("guests should be an empty array: %s", line)
	}
}

func TestFormatJSONError(t *testing.T) {
	t.Parallel()

	line, err := FormatJSONError(errors.New("connection refused"))
	if err != nil {
		t.Fatal(err)
	}

	var out jsonError
	if err := json.Unmarshal([]byte(line), &out); err != nil {
		t.Fatal(err)
	}

	if out.OK || out.Error != "connection refused" || out.Version != jsonVersion {
		t.Errorf("unexpected error output: %+v", out)
	}
}
