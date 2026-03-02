package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aebel/gstow/internal/config"
	"github.com/aebel/gstow/internal/stow"
)

var version = "dev"

var (
	flagDir      = flag.String("d", "", "stow directory (default: current directory)")
	flagTarget   = flag.String("t", "", "target directory")
	flagDelete   = flag.Bool("D", false, "delete (unstow) the packages")
	flagRestow   = flag.Bool("R", false, "restow (unstow then stow again)")
	flagSimulate = flag.Bool("n", false, "simulate; don't make any changes")
	flagVerbose  = flag.Bool("v", false, "verbose output")
	flagAll      = flag.Bool("a", false, "stow/unstow all packages")
	flagVersion  = flag.Bool("V", false, "show version")
	flagHelp     = flag.Bool("h", false, "show help")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: gstow [options] <package>...\n")
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  gstow nvim              # stow nvim package\n")
	fmt.Fprintf(os.Stderr, "  gstow -a                # stow all packages\n")
	fmt.Fprintf(os.Stderr, "  gstow -D nvim           # unstow nvim package\n")
	fmt.Fprintf(os.Stderr, "  gstow -R nvim           # restow nvim package\n")
	fmt.Fprintf(os.Stderr, "  gstow -t ~/.config nvim # override target directory\n")
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *flagVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if *flagHelp {
		usage()
		os.Exit(0)
	}

	args := flag.Args()

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

	if *flagAll || (len(args) == 1 && args[0] == ".") {
		packages, err := s.ListPackages()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to list packages: %v\n", err)
			os.Exit(1)
		}
		args = packages
	}

	if len(args) == 0 {
		packages, _ := s.ListPackages()
		if len(packages) > 0 {
			fmt.Fprintf(os.Stderr, "Error: no packages specified\n\n")
			fmt.Fprintf(os.Stderr, "Available packages:\n")
			for _, pkg := range packages {
				fmt.Fprintf(os.Stderr, "  %s\n", pkg)
			}
			fmt.Fprintf(os.Stderr, "\nUse -a to stow all packages.\n")
		} else {
			fmt.Fprintf(os.Stderr, "Error: no packages found in %s\n", stowDir)
		}
		os.Exit(1)
	}

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
