package filehash

// See README.md for description and usage

import (
	"crypto/sha256"
	"fmt"
	"github.com/kalafut/imohash"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

// A fast checksum of a file. The algorithm uses sampling so it does not look at
// all bytes. Thus, if 2 files don't have the same hash then they are *definitely* different,
// and you can short-circuit your comparison. Conversely, if the hashes are identical,
// then you must do a comprehensive hash on both files to determine if they're actually identical.
func FastChecksumFile(filepath string) string {
	hash := imohash.New()
	checksum, err := hash.SumFile(filepath)
	if err != nil {
		log.Fatalf("Error fast-checksumming file: %v\n%v", filepath, err)
	}
	return fmt.Sprintf("%x", checksum)
}

// A comprehensive but slow checksum of a file.
func DeepChecksumFile(filepath string) string {
	f, err := os.Open(filepath)
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
