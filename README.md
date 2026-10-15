# sow

Interactive, colorful terminal charts for Linux and macOS.

Pipe data from any command, open CSV/TSV/XLSX files, or monitor system tools like `ping`, `vmstat`, `iostat` directly — all rendered as live, switchable graphs in your terminal.

Built with [Charm](https://charm.sh) libraries (Bubble Tea, Lip Gloss, Bubbles) and [ntcharts](https://github.com/NimbleMarkets/ntcharts).

## Install

### From releases

```bash
# Linux (amd64)
curl -sL https://github.com/socketspace-jihad/sow/releases/latest/download/sow_*_linux_amd64.tar.gz | tar xz
sudo mv sow /usr/local/bin/

# macOS (Apple Silicon)
curl -sL https://github.com/socketspace-jihad/sow/releases/latest/download/sow_*_darwin_arm64.tar.gz | tar xz
sudo mv sow /usr/local/bin/
```

### Dev build (latest main)

```bash
curl -sL https://github.com/socketspace-jihad/sow/releases/download/dev/sow_dev_linux_amd64.tar.gz | tar xz
chmod +x sow_dev_linux_amd64
```

### From source

```bash
go install github.com/socketspace-jihad/sow@latest
```

## Quick Start

```bash
# Bar chart from CSV
echo -e "Name,Value\nAlpha,42\nBeta,87\nGamma,23" | sow bar

# Line chart from file
sow line examples/sales.csv --title "Monthly Revenue"

# Live ping monitoring
sow ping google.com

# Live CPU monitoring
sow vmstat -m us -c ocean
```

## Chart Types

| Command | Description |
|---------|-------------|
| `sow bar` | Vertical bar chart **(default)** |
| `sow hbar` | Horizontal bar chart |
| `sow line` | Line chart |
| `sow scatter` | Scatter / dot plot |
| `sow dot` | Alias for scatter |
| `sow sparkline` | Compact sparkline |
| `sow streamline` | Scrolling line **(default for --live)** |
| `sow heatmap` | 2D color-mapped grid |
| `sow pie` | Proportional colored bar with legend |
| `sow table` | Interactive data table |

Press **Tab** to cycle between chart types live.

## Data Sources

### Files (CSV, TSV, XLSX)

```bash
sow bar data.csv
sow line report.tsv
sow pie spreadsheet.xlsx --sheet Sheet2
sow table data.csv
```

### Piped input

```bash
echo -e "A,10\nB,20\nC,30" | sow bar
cat data.csv | sow line
awk '{print $1, $2}' data.txt | sow scatter
```

### Piped input (live streaming)

```bash
ping google.com | sow line --live
vmstat 1 | awk 'NR>2 {print $13}' | sow sparkline --live
while true; do echo $((RANDOM % 100)); sleep 0.5; done | sow streamline --live
```

### Awk formatting

sow auto-detects delimiters (comma, tab, whitespace) and headers. Common awk patterns work out of the box:

```bash
# Single column — numbers only
vmstat 1 | awk 'NR>2 {print $13}' | sow sparkline --live

# Label + value
awk '{print $1, $2}' data.txt | sow bar

# With header row
awk 'BEGIN{print "host","latency"} {print $1,$2}' log.txt | sow line

# CSV output
awk -F: 'BEGIN{OFS=","} {print $1,$3}' /etc/passwd | sow bar

# Tab-separated
awk 'BEGIN{OFS="\t"; print "name","score"} {print $1,$2}' results.txt | sow table
```

## Built-in System Monitors

sow can spawn and graph system tools directly — no piping needed:

```bash
# Ping latency
sow ping google.com
sow ping -c ocean -t "CDN Latency" 8.8.8.8

# CPU / memory / I/O (vmstat)
sow vmstat                      # user CPU % (default)
sow vmstat -m id                # idle CPU %
sow vmstat -m cs -c sunset      # context switches

# Disk I/O (iostat)
sow iostat
sow iostat -m %iowait

# System activity (sar)
sow sar
sow sar -m %user

# Per-process (pidstat)
sow pidstat
sow pidstat -m %CPU
```

Use `sow <tool> --help` to see all available metrics for each tool.

## Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--title` | `-t` | | Chart title |
| `--x` | | `0` | X-axis column (index or name) |
| `--y` | | `1` | Y-axis column (index or name) |
| `--color` | `-c` | `default` | Color theme |
| `--width` | `-w` | auto | Chart width |
| `--height` | `-H` | auto | Chart height |
| `--live` | `-l` | `false` | Live streaming mode |
| `--metric` | `-m` | | Metric for monitor commands |
| `--chart` | | `streamline` | Chart type for monitor commands |
| `--delimiter` | `-d` | auto | Field delimiter |
| `--header` | | auto | First row is header |
| `--no-header` | | | First row is NOT header |
| `--sheet` | `-s` | | Sheet name for XLSX |

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `q` / `Ctrl+C` | Quit |
| `Tab` | Next chart type |
| `Shift+Tab` | Previous chart type |
| `t` | Cycle color theme |
| `?` | Toggle help |
| `r` | Redraw |

## Color Themes

5 built-in themes, cycle with `t` or set with `-c`:

- **default** — Dracula-inspired (cyan, green, orange, pink, purple)
- **ocean** — Blues and teals
- **forest** — Greens and golds
- **sunset** — Warm reds, oranges, pinks
- **mono** — Grayscale

## Auto-Detection

sow detects context automatically:

| Input | X label | Y label | How |
|-------|---------|---------|-----|
| CSV with headers | column name | column name | header row (all non-numeric) |
| `sow ping host` | time | `host latency (ms)` | tool profile |
| `sow vmstat -m us` | time | `user CPU % (%)` | metric description |
| `ping \| sow --live` | time | `host latency (ms)` | `PING <host>` line regex |
| Single-column pipe | index | col 0 | auto-adjust Y to col 0 |
| Generic numbers | time (live) / col N | value (live) / col N | fallback |

## Live Stats

In live/streaming mode, a stats bar shows:

```
cur: 28.3 ms  min: 22.1 ms  max: 45.7 ms  avg: 30.2 ms  n: 42
```

## Project Structure

```
sow/
├── main.go                    # Entry point + version info
├── cmd/
│   ├── root.go                # CLI framework, global flags
│   ├── bar.go                 # bar / hbar subcommands
│   ├── line.go                # line subcommand
│   ├── scatter.go             # scatter / dot subcommands
│   ├── sparkline.go           # sparkline subcommand
│   ├── streamline.go          # streamline subcommand
│   ├── heatmap.go             # heatmap subcommand
│   ├── pie.go                 # pie subcommand
│   ├── table.go               # table subcommand
│   └── monitor.go             # ping/vmstat/iostat/sar/pidstat
├── internal/
│   ├── data/
│   │   ├── types.go           # DataSet, DataPoint, StreamInfo
│   │   ├── parser.go          # Smart text/line parser
│   │   ├── csv.go             # CSV/TSV reader
│   │   ├── xlsx.go            # XLSX reader (excelize)
│   │   ├── stdin.go           # Stdin detection
│   │   └── profiles.go        # Tool profiles (ping, vmstat, etc.)
│   ├── chart/
│   │   ├── config.go          # Chart config + type cycling
│   │   ├── model.go           # Bubble Tea model + legend + stats
│   │   ├── renderer.go        # Renderer interface + factory
│   │   ├── bar.go             # Bar chart (ntcharts barchart)
│   │   ├── line.go            # Line chart (ntcharts wavelinechart)
│   │   ├── scatter.go         # Scatter plot
│   │   ├── sparkline.go       # Sparkline (ntcharts sparkline)
│   │   ├── streamline.go      # Streamline (ntcharts streamlinechart)
│   │   ├── heatmap.go         # Heatmap (ntcharts heatmap)
│   │   ├── pie.go             # Pie chart (custom, Lip Gloss)
│   │   └── table.go           # Table view (Bubbles table)
│   └── theme/
│       └── theme.go           # 5 color themes
└── examples/
    └── sales.csv              # Sample data
```

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUI components
- [ntcharts](https://github.com/NimbleMarkets/ntcharts) — Terminal charts
- [Excelize](https://github.com/xuri/excelize) — XLSX reader
- [Cobra](https://github.com/spf13/cobra) — CLI framework

## License

MIT
