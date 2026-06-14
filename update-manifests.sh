#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status
set -e

# Find all .ebuild files and update their manifests
find . -name "*.ebuild" -print0 | while IFS= read -r -d '' ebuild_path; do
    # Extract the directory and filename
    ebuild_dir=$(dirname "$ebuild_path")
    ebuild_file=$(basename "$ebuild_path")
    
    echo "Processing: $ebuild_file in $ebuild_dir"
    
    # Run the ebuild command inside the target directory
    (cd "$ebuild_dir" && ebuild "$ebuild_file" manifest)
done