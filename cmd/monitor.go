package cmd

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muhammadjihad/sow/internal/chart"
	"github.com/muhammadjihad/sow/internal/data"
	"github.com/muhammadjihad/sow/internal/theme"
	"github.com/spf13/cobra"
)

var (
	flagMetric    string
	flagChartType string
)

func init() {
	// Register a subcommand for each known tool
	for name, profile := range data.Profiles {
		p := profile // capture loop variable
		cmd := &cobra.Command{
			Use:   name + " [args...]",
			Short: fmt.Sprintf("Monitor %s in real-time", name),
			Long:  buildMonitorHelp(p),
			Args:  cobra.ArbitraryArgs,
			RunE: func(cmd *cobra.Command, args []string) error {
				return runMonitor(p, args)
			},
		}
		rootCmd.AddCommand(cmd)
	}

	rootCmd.PersistentFlags().StringVarP(&flagMetric, "metric", "m", "", "Metric/column to graph (tool-specific)")
	rootCmd.PersistentFlags().StringVar(&flagChartType, "chart", "streamline", "Chart type for monitor commands")
}

func buildMonitorHelp(p *data.ToolProfile) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Spawn %s and graph its output in real-time.\n\n", p.Name)
	fmt.Fprintf(&b, "Available metrics (use -m / --metric):\n")
	for name, desc := range p.Metrics {
		marker := "  "
		if name == p.DefaultMetric {
			marker = "* "
		}
		fmt.Fprintf(&b, "  %s%-10s %s\n", marker, name, desc)
	}
	fmt.Fprintf(&b, "\n  * = default\n")
	fmt.Fprintf(&b, "\nExamples:\n")
	switch p.Name {
	case "ping":
		fmt.Fprintf(&b, "  sow ping google.com\n")
		fmt.Fprintf(&b, "  sow ping -t \"My Ping\" -c ocean google.com\n")
	case "vmstat":
		fmt.Fprintf(&b, "  sow vmstat\n")
		fmt.Fprintf(&b, "  sow vmstat -m id              # graph idle CPU\n")
		fmt.Fprintf(&b, "  sow vmstat -m us -c sunset     # user CPU, sunset theme\n")
	case "iostat":
		fmt.Fprintf(&b, "  sow iostat\n")
		fmt.Fprintf(&b, "  sow iostat -m %%user\n")
	case "sar":
		fmt.Fprintf(&b, "  sow sar\n")
		fmt.Fprintf(&b, "  sow sar -m %%idle\n")
	case "pidstat":
		fmt.Fprintf(&b, "  sow pidstat\n")
		fmt.Fprintf(&b, "  sow pidstat -m %%CPU\n")
	}
	return b.String()
}

func runMonitor(profile *data.ToolProfile, toolArgs []string) error {
	metric := flagMetric
	if metric == "" {
		metric = profile.DefaultMetric
	}

	// Validate metric
	if _, ok := profile.Metrics[metric]; !ok {
		fmt.Fprintf(rootCmd.ErrOrStderr(), "Unknown metric %q for %s. Available:\n", metric, profile.Name)
		for name, desc := range profile.Metrics {
			marker := "  "
			if name == profile.DefaultMetric {
				marker = "* "
			}
			fmt.Fprintf(rootCmd.ErrOrStderr(), "  %s%-10s %s\n", marker, name, desc)
		}
		return fmt.Errorf("invalid metric: %s", metric)
	}

	// Build command arguments
	var args []string
	if profile.Name == "ping" {
		if len(toolArgs) == 0 {
			return fmt.Errorf("ping requires a host argument: sow ping google.com")
		}
		args = toolArgs
	} else if len(toolArgs) > 0 {
		args = toolArgs
	} else {
		args = profile.DefaultArgs
	}

	// Spawn the tool
	cmd := exec.Command(profile.Command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating pipe for %s: %w", profile.Name, err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("creating stderr pipe for %s: %w", profile.Name, err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting %s: %w", profile.Name, err)
	}

	// Build config
	th := theme.Get(flagColor)
	title := flagTitle
	if title == "" {
		title = profile.Title
	}

	unit := profile.Unit
	yLabel := metric
	if desc, ok := profile.Metrics[metric]; ok {
		yLabel = desc
	}

	if profile.Name == "ping" && len(toolArgs) > 0 {
		host := toolArgs[len(toolArgs)-1]
		yLabel = host + " latency"
		unit = "ms"
	}

	cfg := &chart.Config{
		ChartType: flagChartType,
		Title:     title,
		Live:      true,
		Theme:     th,
		XLabel:    "time",
		YLabel:    yLabel,
		Unit:      unit,
	}

	ch := make(chan data.DataPoint, 100)

	go func() {
		defer close(ch)
		defer cmd.Wait()

		scanner := bufio.NewScanner(stdout)
		var columns []string
		var headerLines []string
		headerDone := false

		for scanner.Scan() {
			line := scanner.Text()

			if !headerDone {
				if profile.HeaderDetect(line) {
					headerLines = append(headerLines, line)
					continue
				}
				if len(headerLines) > 0 {
					columns = profile.ColumnParse(headerLines)
					headerDone = true
				} else if profile.Name == "ping" {
					headerDone = true
				}
			}

			values := profile.DataParse(line, columns)
			if values == nil {
				continue
			}

			if v, ok := values[metric]; ok {
				ch <- data.DataPoint{
					Label: metric,
					Value: v,
					Time:  time.Now(),
				}
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
		}
	}()

	m := chart.NewModel(cfg, nil, ch)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()

	cmd.Process.Kill()

	return err
}
