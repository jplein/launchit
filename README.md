# Launchit

Data source and reader for use with dmenu-style launchers. To use, run it like this:

```
launchit write | rofi -dmenu -display-columns 1 | launchit read
```

`launchit write` writes out a series of lines with tab-delimited fields. The first field is the description, and the launcher should be configured to show only this field. The second field is a unique ID for the entry. The third field is a hidden field that can be used to search, e.g., the app ID for a window.

The launcher itself will select a line, and then launchit will read the ID from the launcher's output, and use it to launch an application, run a command, switch to a window, etc.

Launchit is in a very early pre-alpha state:

- Documentation is incomplete
- Niri is the only supported window manager for getting a list of windows. There is no support for choosing a workspace.
- Custom commands can be read from YAML, but the YAML file is compiled into the application, not read from a file
