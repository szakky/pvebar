package main

import (
	"math"
	"testing"
)

func TestSummarize(t *testing.T) {
	t.Parallel()

	resources := []Resource{
		{Type: "node", Node: "pve1", CPU: 0.5, MaxCPU: 4, Mem: 2 << 30, MaxMem: 8 << 30},
		{Type: "node", Node: "pve2", CPU: 0.0, MaxCPU: 4, Mem: 2 << 30, MaxMem: 8 << 30},
		{Type: "qemu", VMID: 101, Name: "postgres", Status: "running"},
		{Type: "qemu", VMID: 100, Name: "k3s", Status: "running"},
		{Type: "lxc", VMID: 200, Name: "test", Status: "stopped"},
		{Type: "qemu", VMID: 9000, Name: "template", Status: "stopped", Template: 1},
		{Type: "storage", Node: "pve1"},
	}

	s := Summarize(resources)

	// (0.5*4 + 0*4) / 8 = 25%
	if math.Abs(s.CPU-25) > 0.01 {
		t.Errorf("CPU = %.2f, want 25", s.CPU)
	}

	// 4 GiB / 16 GiB = 25%
	if math.Abs(s.Mem-25) > 0.01 {
		t.Errorf("Mem = %.2f, want 25", s.Mem)
	}

	if s.Running != 2 || s.Total != 3 {
		t.Errorf("Running/Total = %d/%d, want 2/3 (template excluded)", s.Running, s.Total)
	}

	wantOrder := []int{100, 101, 200}
	for i, g := range s.Guests {
		if g.VMID != wantOrder[i] {
			t.Errorf("Guests[%d].VMID = %d, want %d", i, g.VMID, wantOrder[i])
		}
	}
}

func TestSummarizeNoNodes(t *testing.T) {
	t.Parallel()

	s := Summarize(nil)

	if s.CPU != 0 || s.Mem != 0 || s.Total != 0 {
		t.Errorf("empty input should give zero summary, got %+v", s)
	}
}
