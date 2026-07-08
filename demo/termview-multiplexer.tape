Output ../examples/terminal-multiplexer/terminal-multiplexer.gif
Require echo

Set Shell "bash"
Set FontSize 13
Set Width 800
Set Height 600
Set Padding 20
Set LoopOffset 65%

Hide
Type "cd ../examples/terminal-multiplexer"
Enter
Show

Type "go run ."
Enter
Sleep 1500ms

Ctrl+w
Sleep 1000ms

Ctrl+w
Sleep 1000ms

Type "top"
Sleep 1500ms
Enter

Ctrl+w
Sleep 500ms

Type "iostat 1"
Enter
Sleep 2000ms

Ctrl+n
Sleep 3000ms

Ctrl+n
Sleep 2000ms

Ctrl+w
Sleep 1000ms

Ctrl+q
Sleep 500ms
