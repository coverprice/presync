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

type FileHash struct {
	Size     int64
	Checksum string
}

type TFileHash map[string]*FileHash

func (this *FileHash) IsSame(cmp *FileHash) bool {
	return this.Size == cmp.Size && this.Checksum == cmp.Checksum
}

var (
	srcDir       string
	targetDir    string
	logLevelFlag string
	isDryRun     bool
	isDebug      bool
	srcFiles     TFileHash
	targetFiles  TFileHash
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

	srcFiles = make(TFileHash)
	targetFiles = make(TFileHash)
	log.Debugf("Source dir: %q", srcDir)
	log.Debugf("Target dir: %q", targetDir)
	return
}

func walkFiles(hashes TFileHash, rootDir string, doChecksum bool) {
	err := os.Chdir(rootDir)
	if err != nil {
		log.Fatal("Could not change to directory %q: %v", rootDir, err)
	}
	log.Debugf("Walking all files in %q", rootDir)
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Could not access the path %q: %v", path, err)
		}
		if info.Mode().IsRegular() {
			log.Debugf("Found file %q", path)
			var checksum string
			if doChecksum {
				checksum = checksumFile(path)
			}
			filehash := FileHash{
				Size:     info.Size(),
				Checksum: checksum,
			}
			hashes[path] = &filehash
		}
		return nil
	})
	if err != nil {
		log.Fatal("Error walking the tree %q: %v\n", rootDir, err)
	}
}

func compareAndRenameTargetFiles() {
	err := os.Chdir(targetDir)
	if err != nil {
		log.Fatal("Could not change to target directory %q: %v", targetDir, err)
	}
	log.Debugf("Comparing files in the target %q", targetDir)
	for target_path, filehash := range targetFiles {
		var _, ok = srcFiles[target_path]
		if !ok {
			handleTargetFile(target_path, filehash)
		} else {
			log.Debugf("Skipping: target exists in src: %q", target_path)
		}
	}
}

func findSrcFileByFileHash(filehash *FileHash) (src_path string, found bool) {
	var src_filehash *FileHash
	for src_path, src_filehash = range srcFiles {
		if src_filehash.IsSame(filehash) {
			found = true
			return
		}
	}
	return
}

func handleTargetFile(target_path string, target_filehash *FileHash) {
	// The target does not exist in src, so it's either been deleted in src,
	// or it was moved/renamed in src. To see if it was the latter, checksum the
	// target search for it in the srcFiles.
	target_filehash.Checksum = checksumFile(target_path)
	var new_target_path, found = findSrcFileByFileHash(target_filehash)
	if !found {
		// There was no src file that matched the target file's size & checksum. This means
		// the target should be deleted... but we'll let rsync handle that.
		log.Debugf("Skipping: no src file has the same size/checksum as this target: %q", target_path)
		return
	}

	_, err := os.Stat(new_target_path)
	if !os.IsNotExist(err) {
		log.Infof("Skipping rename of %q to %q because target already exists.", target_path, new_target_path)
		return
	}

	log.Infof("Renaming %q to %q", target_path, new_target_path)
	if isDryRun {
		log.Debugf("Skipping: dry-run mode")
		return
	}

	new_target_dir := filepath.Dir(new_target_path)
	log.Debugf("Creating directory %q", new_target_dir)
	if err = os.MkdirAll(new_target_dir, 0755); err != nil {
		log.Fatal("Error creating new target directory %v", new_target_dir)
	}
	if err = os.Rename(target_path, new_target_path); err != nil {
		log.Fatal("Error renaming %v to %v", target_path, new_target_path)
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
	walkFiles(srcFiles, srcDir, true)
	walkFiles(targetFiles, targetDir, false)
	compareAndRenameTargetFiles()
	os.Exit(0)
}
