# Presync

Improves rsync clone performance by rename files in the target that have only
been renamed/moved in the source.

## Usage

    presync [-dry-run] [-debug] -source=SOURCE_DIR -target=TARGET_DIR

## Purpose

This utility solves a problem when using rsync to clone the contents of
a src dir to a target. If a file is renamed or moved in the source, rsync
does not just rename the target file, but instead dumbly considers it to be
2 separate operations: deletion of the old filename on the target, and copying
the "new" file to the target. For large files, this is very inefficient; simply
renaming a file in the source will result in having to transfer over the entire
file.

This utility aims to fix that. It is intended to be run just before the rsync
clone command. It will scan the src/target directories for files that have
only been renamed/moved, and move them on the target so that they match the
source. Then, when rsync clone is run, it will skip recopying the renamed source
file since its target file now matches the src location.

## Building

    go get github.com/coverprice/presync
    cd github.com/coverprice/presync
    go install
