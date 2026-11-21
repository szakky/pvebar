package main

import "sort"

// Summary is what pvebar shows: overall usage and the guest list.
type Summary struct {
	CPU     float64    // percent, 0 - 100, weighted by core count across nodes
	Mem     float64    // percent, 0 - 100, across nodes
	Running int        // guests with status "running"
	Total   int        // all guests (templates excluded)
	Guests  []Resource // VMs and containers, sorted by VMID
}

// Summarize turns raw cluster resources into a Summary.
func Summarize(resources []Resource) Summary {
	var (
		s                 Summary
		cpuUsed, cpuTotal float64
		memUsed, memTotal int64
	)

	for _, r := range resources {
		switch r.Type {
		case "node":
			cpuUsed += r.CPU * r.MaxCPU
			cpuTotal += r.MaxCPU
			memUsed += r.Mem
			memTotal += r.MaxMem
		case "qemu", "lxc":
			if r.Template == 1 {
				continue
			}

			s.Total++
			if r.Status == "running" {
				s.Running++
			}

			s.Guests = append(s.Guests, r)
		}
	}

	if cpuTotal > 0 {
		s.CPU = cpuUsed / cpuTotal * 100
	}

	if memTotal > 0 {
		s.Mem = float64(memUsed) / float64(memTotal) * 100
	}

	sort.Slice(s.Guests, func(i, j int) bool { return s.Guests[i].VMID < s.Guests[j].VMID })

	return s
}
