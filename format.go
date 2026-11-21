package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
)

// FormatPlain returns a single line for tmux or any status bar that shows plain text.
func FormatPlain(s Summary) string {
	return fmt.Sprintf("PVE CPU %.0f%% RAM %.0f%% VM %d/%d", s.CPU, s.Mem, s.Running, s.Total)
}

// waybarOutput is the JSON format Waybar's custom module expects with "return-type": "json".
type waybarOutput struct {
	Text       string `json:"text"`
	Tooltip    string `json:"tooltip"`
	Class      string `json:"class"`
	Percentage int    `json:"percentage"`
}

// FormatWaybar returns one line of Waybar JSON. warn and crit are percent thresholds
// applied to the higher of CPU and RAM usage; they set the CSS class.
func FormatWaybar(s Summary, warn, crit float64) (string, error) {
	peak := max(s.CPU, s.Mem)

	class := "ok"
	switch {
	case peak >= crit:
		class = "critical"
	case peak >= warn:
		class = "warning"
	}

	lines := make([]string, 0, len(s.Guests)+1)
	lines = append(lines, fmt.Sprintf("CPU %.1f%%  RAM %.1f%%", s.CPU, s.Mem))
	for _, g := range s.Guests {
		lines = append(lines, fmt.Sprintf("%d %s (%s): %s", g.VMID, g.Name, g.Type, g.Status))
	}

	return marshalLine(waybarOutput{
		Text:       fmt.Sprintf("PVE %.0f%% %.0f%% %d/%d", s.CPU, s.Mem, s.Running, s.Total),
		Tooltip:    strings.Join(lines, "\n"),
		Class:      class,
		Percentage: int(math.Round(peak)),
	})
}

// FormatWaybarError returns Waybar JSON that shows an error state instead of empty output.
func FormatWaybarError(err error) (string, error) {
	return marshalLine(waybarOutput{
		Text:    "PVE !",
		Tooltip: err.Error(),
		Class:   "error",
	})
}

// jsonVersion is bumped when a field in the JSON output changes meaning or is removed.
// Adding a field does not change it.
const jsonVersion = 1

// jsonOutput is the machine-readable format for widgets and scripts.
// Percentages are 0 - 100; memory is in bytes.
type jsonOutput struct {
	Version  int         `json:"version"`
	OK       bool        `json:"ok"`
	CPU      float64     `json:"cpu"`
	Mem      float64     `json:"mem"`
	MemUsed  int64       `json:"mem_used"`
	MemTotal int64       `json:"mem_total"`
	Running  int         `json:"running"`
	Total    int         `json:"total"`
	Guests   []jsonGuest `json:"guests"`
}

type jsonGuest struct {
	VMID   int     `json:"vmid"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Node   string  `json:"node"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpu"` // percent of the guest's own vCPUs
	Mem    int64   `json:"mem"`
	MaxMem int64   `json:"maxmem"`
}

type jsonError struct {
	Version int    `json:"version"`
	OK      bool   `json:"ok"`
	Error   string `json:"error"`
}

// FormatJSON returns the summary as one line of JSON with the raw numbers.
func FormatJSON(s Summary) (string, error) {
	guests := make([]jsonGuest, 0, len(s.Guests))
	for _, g := range s.Guests {
		guests = append(guests, jsonGuest{
			VMID:   g.VMID,
			Name:   g.Name,
			Type:   g.Type,
			Node:   g.Node,
			Status: g.Status,
			CPU:    round1(g.CPU * 100),
			Mem:    g.Mem,
			MaxMem: g.MaxMem,
		})
	}

	return marshalLine(jsonOutput{
		Version:  jsonVersion,
		OK:       true,
		CPU:      round1(s.CPU),
		Mem:      round1(s.Mem),
		MemUsed:  s.MemUsed,
		MemTotal: s.MemTotal,
		Running:  s.Running,
		Total:    s.Total,
		Guests:   guests,
	})
}

// FormatJSONError returns JSON with ok=false so a widget can show the error.
func FormatJSONError(err error) (string, error) {
	return marshalLine(jsonError{Version: jsonVersion, OK: false, Error: err.Error()})
}

func marshalLine(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encoding json output: %w", err)
	}

	return string(b), nil
}

// round1 rounds to one decimal place to keep the output short.
func round1(x float64) float64 {
	return math.Round(x*10) / 10
}
