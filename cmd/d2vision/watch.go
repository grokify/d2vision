package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	watchOutput    string
	watchLint      bool
	watchDebounce  int
	watchD2Args    []string
	watchOnSuccess string
)

var watchCmd = &cobra.Command{
	Use:   "watch <file.d2> [output.svg]",
	Short: "Watch D2 files and re-render on changes",
	Long: `Watch D2 files for changes and automatically re-render to SVG.

This enables a fast development workflow:
  1. Edit your D2 file in your editor
  2. Save the file
  3. d2vision automatically renders the updated SVG
  4. View the SVG in your browser (with live reload if supported)

Features:
  - Automatic re-rendering on file save
  - Optional linting before render (--lint)
  - Debouncing to prevent multiple renders on rapid saves
  - Custom d2 arguments (--d2-args)
  - Post-render hooks (--on-success)

Examples:
  # Basic watch
  d2vision watch diagram.d2

  # Watch with explicit output
  d2vision watch diagram.d2 output.svg

  # Watch with linting
  d2vision watch diagram.d2 --lint

  # Watch with custom d2 arguments
  d2vision watch diagram.d2 --d2-args="--theme=200"

  # Watch with post-render command (e.g., refresh browser)
  d2vision watch diagram.d2 --on-success="open output.svg"

Press Ctrl+C to stop watching.
`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runWatch,
}

func init() {
	watchCmd.Flags().StringVarP(&watchOutput, "output", "o", "", "Output SVG file (default: input with .svg extension)")
	watchCmd.Flags().BoolVar(&watchLint, "lint", false, "Run d2vision lint before rendering")
	watchCmd.Flags().IntVar(&watchDebounce, "debounce", 100, "Debounce delay in milliseconds")
	watchCmd.Flags().StringSliceVar(&watchD2Args, "d2-args", nil, "Additional arguments to pass to d2")
	watchCmd.Flags().StringVar(&watchOnSuccess, "on-success", "", "Command to run after successful render")
}

func runWatch(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Determine output file
	outputFile := watchOutput
	if len(args) > 1 {
		outputFile = args[1]
	}
	if outputFile == "" {
		ext := filepath.Ext(inputFile)
		outputFile = strings.TrimSuffix(inputFile, ext) + ".svg"
	}

	// Verify input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", inputFile)
	}

	// Verify d2 is installed
	if _, err := exec.LookPath("d2"); err != nil {
		return fmt.Errorf("d2 not found in PATH. Install from https://d2lang.com")
	}

	// Create watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	// Watch the input file's directory (to catch renames/recreates)
	dir := filepath.Dir(inputFile)
	if dir == "" {
		dir = "."
	}
	if err := watcher.Add(dir); err != nil {
		return fmt.Errorf("watching directory: %w", err)
	}

	// Also watch the file directly
	if err := watcher.Add(inputFile); err != nil {
		// Not fatal - directory watch will catch it
		fmt.Printf("Note: Could not watch file directly: %v\n", err)
	}

	// Get absolute path for comparison
	absInput, err := filepath.Abs(inputFile)
	if err != nil {
		absInput = inputFile
	}

	fmt.Printf("Watching %s → %s\n", inputFile, outputFile)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Initial render
	renderD2(inputFile, outputFile)

	// Debounce timer
	var debounceTimer *time.Timer
	debounceDuration := time.Duration(watchDebounce) * time.Millisecond

	// Watch loop
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Check if this event is for our file
			absEvent, _ := filepath.Abs(event.Name)
			if absEvent != absInput && filepath.Base(event.Name) != filepath.Base(inputFile) {
				continue
			}

			// Only react to writes and creates
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}

			// Debounce: reset timer on each event
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceDuration, func() {
				renderD2(inputFile, outputFile)
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("Watch error: %v\n", err)
		}
	}
}

func renderD2(input, output string) {
	timestamp := time.Now().Format("15:04:05")

	// Optional: lint first
	if watchLint {
		fmt.Printf("[%s] Linting %s...\n", timestamp, input)
		lintResult := lintD2(input, readFileContent(input))
		if len(lintResult.Issues) > 0 {
			fmt.Printf("[%s] ⚠ Found %d issue(s):\n", timestamp, len(lintResult.Issues))
			for _, issue := range lintResult.Issues {
				fmt.Printf("  Line %d: %s\n", issue.Line, issue.Message)
			}
		}
	}

	// Build d2 command
	args := []string{input, output}
	args = append(args, watchD2Args...)

	fmt.Printf("[%s] Rendering %s...\n", timestamp, input)
	cmd := exec.Command("d2", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("[%s] ✘ Render failed (%v)\n", timestamp, duration.Round(time.Millisecond))
		return
	}

	fmt.Printf("[%s] ✓ Rendered %s (%v)\n", timestamp, output, duration.Round(time.Millisecond))

	// Run post-render command if specified
	if watchOnSuccess != "" {
		runPostCommand(watchOnSuccess, output)
	}
}

func readFileContent(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func runPostCommand(command, output string) {
	// Replace %s with output file path
	command = strings.ReplaceAll(command, "%s", output)

	// Use shell to run the command
	var cmd *exec.Cmd
	if isWindows() {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Post-render command failed: %v\n", err)
	}
}

func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}
