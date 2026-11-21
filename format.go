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

	return marshalWaybar(waybarOutput{
		Text:       fmt.Sprintf("PVE %.0f%% %.0f%% %d/%d", s.CPU, s.Mem, s.Running, s.Total),
		Tooltip:    strings.Join(lines, "\n"),
		Class:      class,
		Percentage: int(math.Round(peak)),
	})
}

// FormatWaybarError returns Waybar JSON that shows an error state instead of empty output.
func FormatWaybarError(err error) (string, error) {
	return marshalWaybar(waybarOutput{
		Text:    "PVE !",
		Tooltip: err.Error(),
		Class:   "error",
	})
}

func marshalWaybar(out waybarOutput) (string, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return "", fmt.Errorf("encoding waybar output: %w", err)
	}

	return string(b), nil
}
