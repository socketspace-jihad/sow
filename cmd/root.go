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
	flagFixScale  bool // set by PersistentPreRun
	flagPattern   string
	flagSkip      int
	flagRowStream  int
	flagRowMode    string
	flagAddHeader  []string
	flagSkipColumn []int
)

var rootCmd = &cobra.Command{
	Use:   "sow [chart-type] [file]",
	Short: "Interactive terminal charts",
	Long:  "sow renders interactive, colorful charts in your terminal.\nSupports CSV, TSV, XLSX files and piped stdin (including live streaming).",
	Args:  cobra.MaximumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		flagFixScale = cmd.Flags().Changed("ymin") || cmd.Flags().Changed("ymax")
	},
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
	rootCmd.PersistentFlags().StringVarP(&flagPattern, "pattern", "p", "", "Pattern for parsing log lines: {label}={value}")
	rootCmd.PersistentFlags().IntVar(&flagSkip, "skip", 0, "Skip first N lines of input before parsing")
	rootCmd.PersistentFlags().IntVar(&flagRowStream, "row-stream", 0, "Group every N data lines as separate overlay series")
	rootCmd.PersistentFlags().StringVar(&flagRowMode, "row-mode", "overlay", "Row-stream display: overlay or stack")
	rootCmd.PersistentFlags().StringSliceVar(&flagAddHeader, "add-header", nil, "Define column names manually (repeatable, no header line expected)")
	rootCmd.PersistentFlags().IntSliceVar(&flagSkipColumn, "skip-column", nil, "Skip column by position (1-based, repeatable)")
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
		YMin:       flagYMin,
		YMax:       flagYMax,
		PieStyle:   flagPieStyle,
		RowStream:  flagRowStream,
		RowMode:    flagRowMode,
		HiddenRows: make(map[string]bool),
	}

	cfg.FixScale = flagFixScale

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

	m := chart.NewModel(cfg, ds, nil, nil)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func runLive(cfg *chart.Config) error {
	ch := make(chan data.DataPoint, 100)
	infoCh := make(chan data.StreamInfo, 1)
	colsCh := make(chan []string, 1)

	// Compile pattern parser if --pattern is set
	var patternParser *data.PatternParser
	if flagPattern != "" {
		var err error
		patternParser, err = data.NewPatternParser(flagPattern)
		if err != nil {
			return fmt.Errorf("invalid pattern %q: %w", flagPattern, err)
		}
	}

	lp := data.NewLiveParser(cfg.YCol, patternParser, flagSkip, flagRowStream, flagHeader, flagAddHeader, flagSkipColumn)

	go func() {
		defer close(ch)
		sentInfo := false
		sentCols := false
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			points := lp.Feed(line)

			// Send stream info once detected
			if !sentInfo {
				if info := lp.GetStreamInfo(); info != nil {
					infoCh <- *info
					sentInfo = true
				}
			}

			// Send detected columns once
			if !sentCols {
				if cols := lp.Columns(); len(cols) > 0 {
					colsCh <- cols
					sentCols = true
				}
			}

			for i := range points {
				ch <- points[i]
			}
		}
	}()

	// Wait briefly for stream info (ping header, etc.) but not for columns
	// Columns arrive dynamically via colsCh → model
	var streamInfo *data.StreamInfo
	select {
	case info := <-infoCh:
		streamInfo = &info
	case <-time.After(500 * time.Millisecond):
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

	if cfg.XLabel == "" {
		cfg.XLabel = "time"
	}
	if cfg.YLabel == "" {
		cfg.YLabel = "value"
	}

	m := chart.NewModel(cfg, nil, ch, colsCh)
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

	return data.ParseText(string(raw), flagDelimiter, headerMode(), flagSkip)
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
