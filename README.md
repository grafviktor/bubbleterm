# Terminal Multiplexer - Proof of concept #

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/grafviktor/terminal-multiplexer/develop/LICENSE)

## 1. Description ##

Minimal proof-of-concept terminal split viewer that embeds two shell sessions in a single TUI window. Demonstrates routing PTY I/O into x/vt emulators, rendering with [bubbletea](https://github.com/charmbracelet/bubbletea) library, and layout switching.

![Two-pane terminal multiplexer demo with focus switching and horizontal/vertical layout toggle](demo/demo.gif)

## 2. Shortcuts ##

* `ctrl+w` - move focus to next pane
* `ctrl+n` - switch panes layout
* `ctrl+q` - close the app

## 3. Run ##

```bash
go run main.go
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

* Does not support Windows (at least yet).
* In horizontal mode you cannot select text from a single pane only.
* If close one of the terminals, the application will exit.
* Lots of other issues of course!

## 5. License ##

MIT — see [LICENSE](LICENSE).
