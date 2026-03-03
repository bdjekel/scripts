#!/bin/bash

source .env

export AUTO_ARCHIVE_SOURCE_DIR
SOURCE_DIR=$AUTO_ARCHIVE_SOURCE_DIR

# Name of the subfolder for old files
ARCHIVE_FOLDER="Archive (30d+)"

# Find and move files older than 30 days (not directories, not hidden files)
find "$SOURCE_DIR" -maxdepth 1 -type f -mtime +30 -not -name ".*" -exec mv {} "$SOURCE_DIR/$ARCHIVE_FOLDER/" \;