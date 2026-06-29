# Terminal Multiplexer - Proof of concept #

## Description ##

Minimal proof-of-concept terminal split viewer that embeds two shell sessions in a single TUI window. Demonstrates routing PTY I/O into x/vt emulators, rendering with [bubbletea](https://github.com/charmbracelet/bubbletea) library, and layout switching.

## Shortcuts ##

* `ctrl+w` - move focus to next pane
* `ctrl+n` - switch panes layout
* `ctrl+q` - close the app

## Run ##

```go
go run main.go
```

## Limitations and "features" ##

* Only 2-pane layout is supported.
* In horizontal mode you cannot select text from a single pane only.
* If close one of the terminals, the application will exit.
* Possible other issues...