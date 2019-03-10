package file

// See README.md for description and usage

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type DirEntryHandler func(path string, info os.FileInfo) error

func WalkDir(rootDir string, handler DirEntryHandler) {
	err := os.Chdir(rootDir)
	if err != nil {
		log.Fatalf("Could not change to directory %v: %v", rootDir, err)
	}
	log.Debugf("Walking all files in %q", rootDir)
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Could not access the path %v: %v", path, err)
		}
		return handler(path, info)
	})
	if err != nil {
		log.Fatalf("Error walking the tree %v: %v\n", rootDir, err)
	}
}

func RenameTargetFile(baseDir, original_path, new_path string, isDryRun bool) error {
	abs_original_path := filepath.Join(baseDir, original_path)
	abs_new_path := filepath.Join(baseDir, new_path)
	_, err := os.Stat(abs_new_path)
	if !os.IsNotExist(err) {
		log.Infof("Skipping rename of %v to %v because target already exists.", original_path, new_path)
		return nil
	}

	var logprefix = ""
	if isDryRun {
		logprefix = "[Dry-run]: "
	}
	log.Infof("%vRenaming %v to %v", logprefix, original_path, new_path)
	if isDryRun {
		return nil
	}

	new_path_dir := filepath.Dir(abs_new_path)
	log.Debugf("Creating directory: %v", new_path_dir)
	if err = os.MkdirAll(new_path_dir, 0755); err != nil {
		return fmt.Errorf("Error creating directory: %v", new_path_dir)
	}
	if err = os.Rename(abs_original_path, abs_new_path); err != nil {
		return fmt.Errorf("Error renaming %v to %v", abs_original_path, abs_new_path)
	}
	return nil
}
