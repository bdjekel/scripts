#!/bin/bash

# Required parameters:
# @raycast.schemaVersion 1
# @raycast.title Open Scratch
# @raycast.mode silent

# Optional parameters:
# @raycast.icon 🤖
# @raycast.packageName Dev Utils

# Documentation:
# @raycast.description Opens scratch.md file in Kiro IDE for brain dump and quick thoughts
# @raycast.author Brian Jekel
# @raycast.authorURL https://github.com/bdjekel

# Path to your markdown file
NOTE_FILE="/Users/brianjekel/Library/CloudStorage/OneDrive-FirstOrion/Brian Jekel - FO/_SCRATCH.md"

# Open in Kiro
kiro -g "$NOTE_FILE:999"
