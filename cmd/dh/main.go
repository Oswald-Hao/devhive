package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/Oswald-Hao/devhive/internal/tui"
	"github.com/mattn/go-isatty"
)

const version = "0.2.0"

const helpText = `DevHive — multi-agent software development assistant.

USAGE:
  dh [flags]

FLAGS:
  -h, --help        Show this help
  -v, --version     Show version
  -q, --quiet       Suppress non-essential output (banners, spinners)
  --json            Output in JSON format (for scripting)
  --no-tui          Disable interactive TUI, read a single prompt from stdin
  --resume          Resume the last session from ~/.devhive/sessions/
  --model <name>    Override AI model (default: deepseek-v4-pro)

EXAMPLES:
  dh                              Start interactive chat
  dh --help                       Show this help
  dh --version                    Print version and exit
  dh --resume                     Resume previous session
  dh --model claude-sonnet-4-6    Use a different model
  echo "explain Go interfaces" | dh --no-tui
  dh --no-tui --json <<< "what is DevHive?"`

func main() {
	help := flag.Bool("help", false, "")
	helpShort := flag.Bool("h", false, "")
	showVersion := flag.Bool("version", false, "")
	showVersionShort := flag.Bool("v", false, "")
	quiet := flag.Bool("quiet", false, "")
	quietShort := flag.Bool("q", false, "")
	jsonOut := flag.Bool("json", false, "")
	noTUI := flag.Bool("no-tui", false, "")
	resume := flag.Bool("resume", false, "")
	model := flag.String("model", "", "")

	flag.Usage = func() {
		fmt.Fprint(os.Stdout, helpText+"\n")
	}

	// Custom flag parsing to handle unknown flags gracefully
	args := os.Args[1:]
	flag.CommandLine.Parse(args)

	if *help || *helpShort {
		fmt.Fprint(os.Stdout, helpText+"\n")
		return
	}

	if *showVersion || *showVersionShort {
		fmt.Fprintf(os.Stdout, "DevHive v%s\n", version)
		return
	}

	// Check for unknown flags
	for _, a := range args {
		if strings.HasPrefix(a, "-") && !isKnownFlag(a) {
			suggestion := findClosestFlag(a)
			msg := tui.ErrorPrefix.Render() + " Unknown flag: " + a
			if suggestion != "" {
				msg += "\n   Did you mean " + suggestion + "?"
			}
			msg += "\n   Run 'dh --help' for usage."
			fmt.Fprintln(os.Stderr, msg)
			os.Exit(1)
		}
	}

	// Detect TTY
	isTerminal := isatty.IsTerminal(os.Stdout.Fd())

	// --no-tui mode: single prompt from stdin
	if *noTUI {
		runNoTUI(*model, *jsonOut, *quiet || *quietShort, isTerminal)
		return
	}

	// Default: interactive TUI
	if !isTerminal && !*quiet && !*quietShort {
		fmt.Fprintln(os.Stderr, tui.WarningPrefix.Render()+" stdout is not a terminal; TUI may not work correctly. Use --no-tui for scripted input.")
	}

	runChat(*model, *resume)
}

func isKnownFlag(f string) bool {
	known := map[string]bool{
		"--help": true, "-h": true,
		"--version": true, "-v": true,
		"--quiet": true, "-q": true,
		"--json": true,
		"--no-tui": true,
		"--resume": true,
		"--model": true,
	}
	// --model=value or --model value are both fine
	if strings.HasPrefix(f, "--model=") {
		return true
	}
	return known[f]
}

func findClosestFlag(input string) string {
	known := []string{"--help", "--version", "--quiet", "--json", "--no-tui", "--resume", "--model"}
	best := ""
	bestDist := 3 // threshold
	input = strings.TrimLeft(input, "-")
	for _, k := range known {
		kTrim := strings.TrimLeft(k, "-")
		d := levenshtein(input, kTrim)
		if d < bestDist {
			bestDist = d
			best = k
		}
	}
	return best
}

func levenshtein(a, b string) int {
	al, bl := len(a), len(b)
	if al == 0 {
		return bl
	}
	if bl == 0 {
		return al
	}
	dp := make([][]int, al+1)
	for i := range dp {
		dp[i] = make([]int, bl+1)
		dp[i][0] = i
	}
	for j := 0; j <= bl; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= al; i++ {
		for j := 1; j <= bl; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			dp[i][j] = min(dp[i-1][j]+1, min(dp[i][j-1]+1, dp[i-1][j-1]+cost))
		}
	}
	return dp[al][bl]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
