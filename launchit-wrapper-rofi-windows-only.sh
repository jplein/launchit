#!/usr/bin/env bash
launchit write --sort-by-most-recent=false --source windows --columns=name,type --widths=69,11 | rofi -dmenu -display-columns 1 | launchit read
