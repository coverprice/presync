# Presync

Improves rsync clone performance by renaming files in the target that have only
been renamed/moved in the source.

## Usage

    presync [-dry-run] [-debug] -source=SOURCE_DIR -target=TARGET_DIR

## Purpose

This utility solves a problem when using rsync to clone the contents of
a source dir to a target: If a file is renamed or moved in the source, rsync
is too dumb to recognize that and rename it, and instead it will delete the
file on the target and re-copy the source file. For large files, this is very
inefficient; simply renaming a file in the source will result in having to
re-transfer over the entire source file again.

This utility aims to fix that. It should be run just before the rsync
clone command. It will scan the src/target directories for files that have
only been renamed/moved, and move them on the target to match the source,
thus avoiding the need to delete/re-copy the source file.

## How it works

It reads both the source and target directory trees, and searches for any
target files that aren't present in the source. This must mean that the
file was either deleted in the source, or has been moved/renamed. It
uses file sizes + checksums to verify if it was a move/rename, and renames
the target file to match the new location.

## Building

    go get github.com/coverprice/presync
    cd github.com/coverprice/presync/cmd/presync
    go install
