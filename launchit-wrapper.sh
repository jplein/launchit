#!/bin/bash
cd "$(dirname "$0")"
./launchit | fuzzel --dmenu --with-nth=1 | ./launchit --read
