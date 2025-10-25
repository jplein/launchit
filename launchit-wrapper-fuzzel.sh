#!/bin/bash
"$(dirname "$0")/launchit" | fuzzel --dmenu --with-nth=1 | "$(dirname "$0")/launchit" read
