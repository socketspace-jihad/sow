package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/socketspace-jihad/sow/internal/chart"
	"github.com/socketspace-jihad/sow/internal/data"
	"github.com/socketspace-jihad/sow/internal/theme"
	"github.com/spf13/cobra"
)

var (
	flagTitle     string
	flagXCol      string
	flagYCol      string
	flagColor     string
	flagWidth     int
	flagHeight    int
	flagHeader    bool
	flagNoHeader  bool
	flagDelimiter string
	flagSheet     string
	flagLive      bool
	flagYMin      float64
	flagYMax      float64
	flagPieStyle  string
	flagOverlay   []string
)

var rootCmd = &cobra.Command{
	Use:   "sow [chart-type] [file]",
	Short: "Interactive terminal charts",
	Long:  "sow renders interactive, colorful charts in your terminal.\nSupports CSV, TSV, XLSX files and piped stdin (including live streaming).",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		defaultType := "bar"
		if flagLive {
			defaultType = "streamline"
		}
		return runChart(defaultType, args)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagTitle, "title", "t", "", "Chart title")
	rootCmd.PersistentFlags().StringVar(&flagXCol, "x", "0", "X-axis column (index or name)")
	rootCmd.PersistentFlags().StringVar(&flagYCol, "y", "1", "Y-axis column (index or name)")
	rootCmd.PersistentFlags().StringVarP(&flagColor, "color", "c", "default", "Color theme (default, ocean, forest, sunset, mono)")
	rootCmd.PersistentFlags().IntVarP(&flagWidth, "width", "w", 0, "Chart width (0 = auto)")
	rootCmd.PersistentFlags().IntVarP(&flagHeight, "height", "H", 0, "Chart height (0 = auto)")
	rootCmd.PersistentFlags().BoolVar(&flagHeader, "header", false, "First row is header")
	rootCmd.PersistentFlags().BoolVar(&flagNoHeader, "no-header", false, "First row is NOT header")
	rootCmd.PersistentFlags().StringVarP(&flagDelimiter, "delimiter", "d", "", "Field delimiter (auto-detect if empty)")
	rootCmd.PersistentFlags().StringVarP(&flagSheet, "sheet", "s", "", "Sheet name for XLSX files")
	rootCmd.PersistentFlags().BoolVarP(&flagLive, "live", "l", false, "Enable live/streaming mode")
	rootCmd.PersistentFlags().Float64Var(&flagYMin, "ymin", 0, "Fixed Y-axis minimum")
	rootCmd.PersistentFlags().Float64Var(&flagYMax, "ymax", 0, "Fixed Y-axis maximum")
	rootCmd.PersistentFlags().StringVar(&flagPieStyle, "pie-style", "bar", "Pie chart style: bar or circle")
	rootCmd.PersistentFlags().StringSliceVar(&flagOverlay, "overlay", nil, "Overlay columns (name or index, repeatable)")
}

// SetVersion sets version info from ldflags.
func SetVersion(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// runChart is the shared factory used by all subcommands.
func runChart(chartType string, args []string) error {
	th := theme.Get(flagColor)

	cfg := &chart.Config{
		ChartType: chartType,
		Title:     flagTitle,
		XCol:      flagXCol,
		YCol:      flagYCol,
		Width:     flagWidth,
		Height:    flagHeight,
		Theme:     th,
		Live:      flagLive,
		YMin:      flagYMin,
		YMax:      flagYMax,
		FixScale:  flagYMin != 0 || flagYMax != 0,
		PieStyle:  flagPieStyle,
	}

	if flagLive {
		return runLive(cfg)
	}
	return runStatic(cfg, args)
}

func runStatic(cfg *chart.Config, args []string) error {
	var ds *data.DataSet
	var err error

	if len(args) > 0 {
		ds, err = readFile(args[0])
	} else {
		ds, err = readStdin()
	}
	if err != nil {
		return err
	}
	if ds == nil || len(ds.Rows) == 0 {
		return fmt.Errorf("no data to display")
	}

	cfg.XColIdx = ds.ResolveColumn(cfg.XCol)
	cfg.YColIdx = ds.ResolveColumn(cfg.YCol)

	// Auto-adjust for single-column data:
	// If only 1 column, use it as Y (the value) and generate row index as X
	numCols := ds.NumColumns()
	if numCols == 1 {
		cfg.XColIdx = -1 // signal: use row index as X
		cfg.YColIdx = 0
	} else if cfg.YColIdx >= numCols {
		cfg.YColIdx = numCols - 1
	}

	// Auto-detect legend from CSV/TSV headers
	if len(ds.Headers) > 0 {
		if cfg.XColIdx >= 0 && cfg.XColIdx < len(ds.Headers) {
			cfg.XLabel = ds.Headers[cfg.XColIdx]
		}
		if cfg.YColIdx >= 0 && cfg.YColIdx < len(ds.Headers) {
			cfg.YLabel = ds.Headers[cfg.YColIdx]
		}
	}
	if cfg.XLabel == "" {
		if cfg.XColIdx < 0 {
			cfg.XLabel = "index"
		} else {
			cfg.XLabel = fmt.Sprintf("col %d", cfg.XColIdx)
		}
	}
	if cfg.YLabel == "" {
		cfg.YLabel = fmt.Sprintf("col %d", cfg.YColIdx)
	}

	// Pre-populate overlays from --overlay flag
	for i, ov := range flagOverlay {
		colIdx := ds.ResolveColumn(ov)
		name := ov
		if colIdx < len(ds.Headers) {
			name = ds.Headers[colIdx]
		}
		cfg.Overlays = append(cfg.Overlays, chart.OverlayConfig{
			Name:   name,
			ColIdx: colIdx,
			Color:  i + 1,
		})
	}

	m := chart.NewModel(cfg, ds, nil)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func runLive(cfg *chart.Config) error {
	ch := make(chan data.DataPoint, 100)
	infoCh := make(chan data.StreamInfo, 1)

	go func() {
		defer close(ch)
		detected := false
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()

			// Try to detect stream source from header lines (first few lines)
			if !detected {
				if info := data.DetectStream(line); info != nil {
					infoCh <- *info
					detected = true
					continue // header line, not data
				}
			}

			dp := data.ParseLine(line)
			if dp != nil {
				dp.Time = time.Now()
				// If we detected the source, tag the label
				ch <- *dp
			}
		}
	}()

	// Wait briefly for stream info detection before starting TUI
	var streamInfo *data.StreamInfo
	select {
	case info := <-infoCh:
		streamInfo = &info
	case <-time.After(500 * time.Millisecond):
		// No header detected, proceed with defaults
	}

	if streamInfo != nil {
		if cfg.XLabel == "" {
			cfg.XLabel = streamInfo.XLabel
		}
		if cfg.YLabel == "" {
			cfg.YLabel = streamInfo.YLabel
		}
		if cfg.Unit == "" {
			cfg.Unit = streamInfo.Unit
		}
	}

	// Defaults for live mode
	if cfg.XLabel == "" {
		cfg.XLabel = "time"
	}
	if cfg.YLabel == "" {
		cfg.YLabel = "value"
	}

	m := chart.NewModel(cfg, nil, ch)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func readFile(path string) (*data.DataSet, error) {
	ext := strings.ToLower(path)
	switch {
	case strings.HasSuffix(ext, ".xlsx"), strings.HasSuffix(ext, ".xls"):
		return data.ReadXLSX(path, flagSheet)
	case strings.HasSuffix(ext, ".tsv"):
		return data.ReadCSV(path, '\t', headerMode())
	case strings.HasSuffix(ext, ".csv"):
		return data.ReadCSV(path, ',', headerMode())
	default:
		// Try CSV with auto-detect delimiter
		return data.ReadCSVAuto(path, headerMode())
	}
}

func readStdin() (*data.DataSet, error) {
	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice != 0 {
		return nil, fmt.Errorf("no input: provide a file argument or pipe data via stdin")
	}

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}

	return data.ParseText(string(raw), flagDelimiter, headerMode())
}

// headerMode returns: 1 = has header, -1 = no header, 0 = auto-detect
func headerMode() int {
	if flagHeader {
		return 1
	}
	if flagNoHeader {
		return -1
	}
	return 0
}
