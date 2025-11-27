#!/usr/bin/env bash
"$(dirname "$0")/launchit" write --source commands --columns=name --widths=80 --icons=false | vicinae dmenu | "$(dirname "$0")/launchit" read
