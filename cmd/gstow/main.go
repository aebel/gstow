package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aebel/gstow/internal/config"
	"github.com/aebel/gstow/internal/stow"
)

var (
	flagDir      = flag.String("d", "", "stow directory (default: current directory)")
	flagTarget   = flag.String("t", "", "target directory")
	flagDelete   = flag.Bool("D", false, "delete (unstow) the packages")
	flagRestow   = flag.Bool("R", false, "restow (unstow then stow again)")
	flagSimulate = flag.Bool("n", false, "simulate; don't make any changes")
	flagVerbose  = flag.Bool("v", false, "verbose output")
	flagHelp     = flag.Bool("h", false, "show help")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: gstow [options] <package>...\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  gstow nvim              # stow nvim package\n")
	fmt.Fprintf(os.Stderr, "  gstow -D nvim           # unstow nvim package\n")
	fmt.Fprintf(os.Stderr, "  gstow -R nvim           # restow nvim package\n")
	fmt.Fprintf(os.Stderr, "  gstow -t ~/.config nvim # override target directory\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *flagHelp {
		usage()
		os.Exit(0)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Error: no packages specified\n\n")
		usage()
		os.Exit(1)
	}

	stowDir := *flagDir
	if stowDir == "" {
		var err error
		stowDir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to get current directory: %v\n", err)
			os.Exit(1)
		}
	}
	stowDir, err := filepath.Abs(stowDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to resolve stow directory: %v\n", err)
		os.Exit(1)
	}

	cfg := config.New()
	if *flagTarget != "" {
		cfg.Target = *flagTarget
	}

	s := stow.New(stowDir, cfg)
	s.SetSimulate(*flagSimulate)
	s.SetVerbose(*flagVerbose)

	if *flagSimulate && *flagVerbose {
		fmt.Fprintf(os.Stderr, "=== SIMULATION MODE ===\n")
	}

	var actionErr error
	switch {
	case *flagDelete:
		actionErr = s.Unstow(args...)
	case *flagRestow:
		actionErr = s.Restow(args...)
	default:
		actionErr = s.Stow(args...)
	}

	if actionErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", actionErr)
		os.Exit(1)
	}
}
