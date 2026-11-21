# pvebar

A tiny status line for [Proxmox VE](https://www.proxmox.com/en/proxmox-virtual-environment) in your tmux or Waybar.

```
PVE CPU 23% RAM 63% VM 2/3
```

pvebar makes one read-only API call (`GET /cluster/resources`), prints one line and exits.
Your status bar runs it on an interval.

## Install

```bash
go install github.com/szakky/pvebar@latest
```

The binary goes to `$(go env GOPATH)/bin`, so make sure that directory is on your `PATH`.

## Set up a read-only API token

Run these on the Proxmox VE host. The `PVEAuditor` role can read everything and change nothing.

```bash
pveum user add pvebar@pve --comment "pvebar (read-only)"
pveum acl modify / --users pvebar@pve --roles PVEAuditor
pveum user token add pvebar@pve bar --privsep 0
```

Copy the `value` from the output. It is shown only once.
With `--privsep 0` the token gets exactly the user's permissions, which here is `PVEAuditor` only.

## Configure

Create `~/.config/pvebar/config.json`:

```json
{
  "endpoint": "https://192.168.1.10:8006",
  "token": "pvebar@pve!bar=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "insecure": true
}
```

```bash
chmod 600 ~/.config/pvebar/config.json
```

`insecure: true` skips TLS certificate verification. Proxmox VE uses a self-signed certificate by default;
set it to `false` if your host has a trusted certificate.

Check it works:

```bash
pvebar
pvebar --format waybar
```

## tmux

```tmux
set -g status-interval 30
set -g status-right '#(pvebar) | %H:%M'
```

## Waybar

```jsonc
"custom/pvebar": {
  "exec": "pvebar --format waybar",
  "return-type": "json",
  "interval": 30,
  "format": "{}"
}
```

Hover the module to see every VM and container with its status.
The CSS class is `ok`, `warning`, `critical` or `error`:

```css
#custom-pvebar.warning  { color: #e5c07b; }
#custom-pvebar.critical { color: #e06c75; }
#custom-pvebar.error    { color: #e06c75; }
```

## Options

| Flag | Default | Description |
| --- | --- | --- |
| `--format` | `plain` | `plain` or `waybar` |
| `--config` | `~/.config/pvebar/config.json` | Path to the config file |
| `--timeout` | `5s` | Timeout for the API request |
| `--warn` | `80` | Percent of CPU or RAM that sets the `warning` class |
| `--crit` | `90` | Percent of CPU or RAM that sets the `critical` class |
| `--version` | | Print the version |

CPU is averaged across nodes, weighted by core count. Templates are not counted as guests.

On an error pvebar prints `PVE !` and the details to stderr.
In `waybar` mode it prints an `error` JSON object and exits 0, so the bar keeps showing the error state.

## Development

```bash
go test ./...
go build .
```

## License

MIT
