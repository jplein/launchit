#!/usr/bin/env bash
launchit write --source windows --columns=name,type --widths=69,11 | rofi -dmenu -display-columns 1 | launchit read
