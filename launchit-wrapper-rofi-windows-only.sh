#!/usr/bin/env bash
"$(dirname "$0")/launchit" write --source windows --columns=name,type --widths=69,11 | rofi -dmenu -display-columns 1 | "$(dirname "$0")/launchit" read
