#!/bin/bash
cd "$(dirname "$0")"
./launchit | rofi -dmenu -normal-window -display-columns 1 | ./launchit --read
