# Termview - TUI terminal component for Go #

Termview is a terminal component which is designed to work with [Bubble Tea](https://github.com/charmbracelet/bubbletea) v2. It embeds a real shell session in your app using the same model/update/view pattern as other [Charm Bracelet](https://github.com/charmbracelet) components.

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/grafviktor/termview/develop/LICENSE)

## 1. Functional demo ##

This demo represents routing PTY I/O into x/vt emulators, rendering with [bubbletea](https://github.com/charmbracelet/bubbletea) library.

![Two-pane terminal multiplexer demo with focus switching and horizontal/vertical layout toggle](examples/terminal-multiplexer/terminal-multiplexer.gif)

## 2. Installation and usage ##

```bash
go get github.com/grafviktor/termview@v0.1.0
```

```go
import "github.com/grafviktor/termview"

...
term, err := termview.New(
    termview.WithCommand("/bin/bash"),
    termview.WithInitialWidth(80),
    termview.WithInitialHeight(24),
)
```

Also see [terminal-simple](examples/terminal-simple) and [terminal-multiplexer](examples/terminal-multiplexer) for the examples.

## 3. License ##

MIT - see [LICENSE](LICENSE).
