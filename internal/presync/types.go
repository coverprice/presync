package presync

// Properties which uniquely identify a file's contents. When comparing two files,
// the Size is used as a quick short-circuit check to avoid having to compute
// the checksum of both files.
type FileHash struct {
	Size         int64
	FastChecksum string
}

// Since the FastChecksum uses sampling, this function only returns if 2 files *may*
// be the same.
func (this *FileHash) MayBeSame(cmp *FileHash) bool {
	return this.Size == cmp.Size && this.FastChecksum == cmp.FastChecksum
}

// Maps a file's path to its size/hash.
type TFileHash map[string]*FileHash

func (this TFileHash) doesPathExist(path string) bool {
	var _, exists = this[path]
	return exists
}

func (this TFileHash) findByFileHash(needle *FileHash) (path string, found bool) {
	var hash *FileHash
	for path, hash = range this {
		if hash.MayBeSame(needle) {
			found = true
			return
		}
	}
	return
}
