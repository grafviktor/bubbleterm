## KeyPressMsg flow ##

### 1. Write to the shell ###

* User types a character and Bubble Tea generates a corresponding tea.KeyPressMsg
* Terminal forwards this message to github.com/charmbracelet/x/vt/key.go#SendKey (`tw.emu.SendKey(vt.KeyPressEvent(msg))`)
* x/vt converts the message into bytes and write them into emulator's internal output buffer (`io.WriteString(e.pw, seq)`)
* It unblocks the writeShell() goroutine which is waiting on `tw.emu.Read(buf)` and picks up those bytes
* writeShell writes them to the PTY: `tw.pty.Write(buf[:n])`
* The shell (zsh) running on the other end of the PTY receives those bytes as input.

Now we must read the output from the shell and send it back to Bubble Tea.

### 2. Read from the shell ###

That is the continuation of the process described in point 1.

* The shell processes the input and produces output (e.g., it echoes the character back, or executes a command)
* That output is written back to the PTY by the shell
* Now readShell reads from the PTY and gets the shell's response
* A TermOutputMsg is created and sent back to Update
* The output is written to the emulator for rendering on screen