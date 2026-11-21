// Command pvebar prints a one-line Proxmox VE status for tmux, Waybar and other status bars.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

var version = "dev"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run holds the program logic so tests can call it without exiting the process.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("pvebar", flag.ContinueOnError)
	fs.SetOutput(stderr)

	format := fs.String("format", "plain", `output format: "plain" or "waybar"`)
	configPath := fs.String("config", defaultConfigPath(), "path to the config file")
	timeout := fs.Duration("timeout", 5*time.Second, "timeout for the API request")
	warn := fs.Float64("warn", 80, "warning threshold in percent (waybar class)")
	crit := fs.Float64("crit", 90, "critical threshold in percent (waybar class)")
	showVersion := fs.Bool("version", false, "print the version and exit")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *showVersion {
		fmt.Fprintln(stdout, "pvebar", version)
		return 0
	}

	if *format != "plain" && *format != "waybar" {
		fmt.Fprintf(stderr, "pvebar: unknown format %q (use plain or waybar)\n", *format)
		return 2
	}

	summary, err := fetchSummary(*configPath, *timeout)
	if err != nil {
		return reportError(*format, err, stdout, stderr)
	}

	if *format == "waybar" {
		out, err := FormatWaybar(summary, *warn, *crit)
		if err != nil {
			return reportError(*format, err, stdout, stderr)
		}

		fmt.Fprintln(stdout, out)

		return 0
	}

	fmt.Fprintln(stdout, FormatPlain(summary))

	return 0
}

func fetchSummary(configPath string, timeout time.Duration) (Summary, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return Summary{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := NewClient(cfg.Endpoint, cfg.Token, cfg.Insecure, timeout)

	resources, err := client.Resources(ctx)
	if err != nil {
		return Summary{}, err
	}

	return Summarize(resources), nil
}

// reportError prints a short marker for the bar and the details to stderr.
func reportError(format string, err error, stdout, stderr io.Writer) int {
	fmt.Fprintln(stderr, "pvebar:", err)

	if format == "waybar" {
		out, jsonErr := FormatWaybarError(err)
		if jsonErr == nil {
			fmt.Fprintln(stdout, out)
		}

		// Exit 0 so Waybar keeps showing the error state.
		return 0
	}

	fmt.Fprintln(stdout, "PVE !")

	return 1
}
