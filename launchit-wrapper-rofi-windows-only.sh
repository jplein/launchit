#!/bin/bash
"$(dirname "$0")/launchit" write --source windows | rofi -dmenu -normal-window -display-columns 1 | "$(dirname "$0")/launchit" read
