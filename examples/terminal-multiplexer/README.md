# Terminal Multiplexer #

## 1. Description ##

Minimal proof-of-concept terminal split viewer that embeds two shell sessions in a single TUI window. Demonstrates routing PTY I/O into x/vt emulators, rendering with [bubbletea](https://github.com/charmbracelet/bubbletea) library, and layout switching.

![Two-pane terminal multiplexer demo with focus switching and horizontal/vertical layout toggle](terminal-multiplexer.gif)

## 2. Shortcuts ##

* `ctrl+w` - move focus to next pane
* `ctrl+n` - switch panes layout
* `ctrl+q` - close the app

## 3. Run ##

```bash
go run .
```

If running from VS Code, make sure to use integrated terminal. VS Code config example:

```json
{
  "configurations": [
    {
      ...
      "console": "integratedTerminal"
    }
  ]
}
```

## 4. Limitations and "features" ##

* In horizontal mode you cannot select text from a single pane only.
* If close one of the terminals, the application will exit.
