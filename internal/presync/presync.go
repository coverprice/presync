package presync

import (
	"github.com/coverprice/presync/internal/file"
	"github.com/coverprice/presync/internal/filehash"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var (
	srcDir      string
	targetDir   string
	isDryRun    bool
	srcFiles    TFileHash
	targetFiles TFileHash
)

func CompareAndRenameTargetFiles(source_dir, target_dir string, is_dry_run bool) {
	srcDir = source_dir
	targetDir = target_dir
	isDryRun = is_dry_run

	// Gather srcFiles
	srcFiles = make(TFileHash)
	file.WalkDir(srcDir, func(path string, info os.FileInfo) error {
		if info.Mode().IsRegular() {
			log.Debugf("Source: %v", path)
			srcFiles[path] = &FileHash{
				Size:         info.Size(),
				FastChecksum: filehash.FastChecksumFile(path),
			}
		}
		return nil
	})

	// Gather targetFiles
	targetFiles = make(TFileHash)
	file.WalkDir(targetDir, func(path string, info os.FileInfo) error {
		if info.Mode().IsRegular() {
			log.Debugf("Target: %v", path)
			targetFiles[path] = &FileHash{
				Size: info.Size(),
			}
		}
		return nil
	})

	// Note: we walk the entire target dir before we start comparing & renaming, to avoid
	// complications with walking the target whilst possibly making changes to the
	// underlying directories.
	log.Debugf("Comparing / renaming files in the target %v", targetDir)
	for target_path, target_filehash := range targetFiles {
		compareAndRenameTargetFile(target_path, target_filehash)
	}
}

// Given a target file, attempts to see if a corresponding file exists in src. It first looks at
// the relative path, and bails Attempts to find the given target file in src by only looking at its
// contents (i.e. FileHash), and if
// if found it will

// If same-filepath-exists-in-src -> do nothing
// else if different-filepath-but-same-contents-exists -> attempt to rename
func compareAndRenameTargetFile(target_path string, target_filehash *FileHash) {
	if srcFiles.doesPathExist(target_path) {
		// log.Debugf("Skipping: target exists in src: %v", target_path)
		return
	}

	// The target file does not exist in src, so either:
	// Case 1) it's been deleted in src
	// Case 2) it was moved/renamed in src.
	// To verify if Case 2, search for it by content checksum.
	src_path, found := findSrcFileWithSameContents(target_path, target_filehash)
	if !found {
		// There was no src file that matched the target file's contents. This means
		// the target file should be deleted... but we'll let rsync handle that.
		log.Debugf("Skipping: no src file has the same size/checksum as this target: %v", target_path)
		return
	}

	// A file exists in src with exactly the same contents as target_path, so we can just
	// rename target_path to the src_path, so rsync will consider this
	// a no-op instead of expensively removing the target file and re-copying the
	// source file.
	if err := file.RenameTargetFile(targetDir, target_path, src_path, isDryRun); err != nil {
		log.Fatal(err)
	}
}

// Attempt to find a file in Src with the same contents as the target_file
func findSrcFileWithSameContents(target_path string, target_filehash *FileHash) (src_path string, found bool) {
	target_filehash.FastChecksum = filehash.FastChecksumFile(target_path)
	src_path, found = srcFiles.findByFileHash(target_filehash)
	if !found {
		log.Debugf("No src file has the same size/checksum as this target: %v", target_path)
		return
	}

	if src_path == target_path {
		// This should never happen because we already looked up the target_path
		// in srcFiles (above) and didn't find it.
		log.Fatalf("Internal error: src_path==target_path : %v", src_path)
	}

	// According to the fast-checksum, the found src file *might* have the same contents
	// as the target file. To confirm if it's *definitely* the same
	// contents, we must do an (expensive) deep checksum on both.
	srcChecksum := filehash.DeepChecksumFile(filepath.Join(srcDir, src_path))
	targetChecksum := filehash.DeepChecksumFile(filepath.Join(targetDir, target_path))
	found = (srcChecksum == targetChecksum)
	return
}
