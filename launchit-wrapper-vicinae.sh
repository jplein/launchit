#!/usr/bin/env bash
"$(dirname "$0")/launchit" write --columns=name,type --widths=65,11 --icons=false | vicinae dmenu | "$(dirname "$0")/launchit" read
