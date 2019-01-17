package main

// See README.md for description and usage

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

var (
	srcDir       string
	targetDir    string
	logLevelFlag string
	isDryRun     bool
	isDebug      bool
	srcFiles     map[string]string
)

func init() {
	flag.StringVar(&srcDir, "source", "", "Path to the source directory.")
	flag.StringVar(&targetDir, "target", "", "Path to the target directory.")
	flag.BoolVar(&isDryRun, "dry-run", false, "True to operate without doing anything")
	flag.BoolVar(&isDebug, "debug", false, "Print additional information")
}

func initialize() {
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
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Could not get current directory. %v", err)
	}
	if !path.IsAbs(srcDir) {
		srcDir = filepath.Join(cwd, srcDir)
	}
	if !path.IsAbs(targetDir) {
		targetDir = filepath.Join(cwd, targetDir)
	}
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		log.Fatal("The source directory does not exist. %v", srcDir)
	}
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		log.Fatal("The target directory does not exist. %v", targetDir)
	}
	// TODO: check that src/target are not inside each other.

	srcFiles = make(map[string]string)
	log.Debugf("Source dir: %q", srcDir)
	log.Debugf("Target dir: %q", targetDir)
	return
}

func getAllSourceFiles() {
	err := os.Chdir(srcDir)
	if err != nil {
		log.Fatal("Could not change to source directory %q: %v", srcDir, err)
	}
	log.Debugf("Checksumming all files in %q", srcDir)
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Could not access the source path %q: %v", path, err)
		}
		if info.Mode().IsRegular() {
			log.Debugf("Found source file %q", path)
			srcFiles[path] = checksumFile(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal("Error walking the source tree %q: %v\n", srcDir, err)
	}
}

func compareAndRenameTargetFiles() {
	err := os.Chdir(targetDir)
	if err != nil {
		log.Fatal("Could not change to target directory %q: %v", targetDir, err)
	}
	log.Debugf("Comparing files in the target %q", targetDir)
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Could not access the target path %q: %v", path, err)
		}
		if info.Mode().IsRegular() {
			log.Debugf("Found target file %q", path)
			handleTargetFile(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal("Error walking the target tree %q: %v\n", srcDir, err)
	}
}

func handleTargetFile(target_filepath string) {
	var _, ok = srcFiles[target_filepath]
	if ok {
		log.Debugf("Skipping: target exists in src: %q", target_filepath)
		// target exists in src, so do nothing.
		return
	}

	// The target does not exist in src, so it *may* have been a src file that was
	// renamed. Checksum the target and see if it matches the sum of any other file.
	var target_checksum = checksumFile(target_filepath)
	var found = false
	var src_filepath string
	var checksum string
	for src_filepath, checksum = range srcFiles {
		if checksum == target_checksum {
			found = true
			break
		}
	}
	if !found {
		// There was no src file that matched the target file's checksum. This means
		// the target should be deleted... but we'll let rsync handle that.
		log.Debugf("Skipping: no src file has the same checksum as this target: %q", target_filepath)
		return
	}

	var new_target_filepath = src_filepath
	_, err := os.Stat(new_target_filepath)
	if !os.IsNotExist(err) {
		log.Infof("Skipping rename of %q to %q because target already exists.", target_filepath, new_target_filepath)
		return
	}

	log.Infof("Renaming %q to %q", target_filepath, new_target_filepath)
	if isDryRun {
		log.Debugf("Skipping: dry-run mode")
		return
	}

	new_target_dir := filepath.Join(targetDir, filepath.Dir(src_filepath))
	log.Debugf("Creating directory %q", new_target_dir)
	if err = os.MkdirAll(new_target_dir, 0755); err != nil {
		log.Fatal("Error creating new target directory %v", new_target_dir)
	}
	if err = os.Rename(target_filepath, new_target_filepath); err != nil {
		log.Fatal("Error renaming %v to %v", target_filepath, new_target_filepath)
	}
}

func checksumFile(file string) string {
	log.Debugf("Checksumming %q", file)
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func main() {
	initialize()
	getAllSourceFiles()
	compareAndRenameTargetFiles()
	os.Exit(0)
}
