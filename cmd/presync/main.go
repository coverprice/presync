package main

// See README.md for description and usage

import (
	"flag"
	"github.com/coverprice/presync/internal/presync"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	srcDir    string
	targetDir string
	isDryRun  bool
	isDebug   bool
)

func init() {
	flag.StringVar(&srcDir, "source", "", "Path to the source directory.")
	flag.StringVar(&targetDir, "target", "", "Path to the target directory.")
	flag.BoolVar(&isDryRun, "dry-run", false, "Print what would be done, but don't rename any files")
	flag.BoolVar(&isDebug, "debug", false, "Print additional information")
}

func parseArgs() {
	flag.Parse()

	if runtime.GOOS == "windows" {
		// This fixes up the terminal colors in Windows
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(colorable.NewColorableStdout())
	}
	if isDebug {
		log.SetLevel(log.DebugLevel)
	}

	// Parse and validate params
	if srcDir == "" || targetDir == "" {
		log.Fatal("You must specify a source and target directory")
	}
	log.Debugf("Raw source dir: %v", srcDir)
	log.Debugf("Raw target dir: %v", targetDir)
	var err error
	if srcDir, err = filepath.Abs(srcDir); err != nil {
		log.Fatal("Error converting source directory to absolute path.")
	}
	if targetDir, err = filepath.Abs(targetDir); err != nil {
		log.Fatal("Error converting target directory to absolute path.")
	}
	log.Debugf("Absolute source dir: %v", srcDir)
	log.Debugf("Absolute target dir: %v", targetDir)
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		log.Fatalf("The source directory does not exist. %v", srcDir)
	}
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		log.Fatalf("The target directory does not exist. %v", targetDir)
	}
	if strings.HasPrefix(srcDir, targetDir) || strings.HasPrefix(targetDir, srcDir) {
		log.Fatal("One directory cannot be within the other")
	}
}

func main() {
	parseArgs()
	presync.CompareAndRenameTargetFiles(srcDir, targetDir, isDryRun)
	os.Exit(0)
}
