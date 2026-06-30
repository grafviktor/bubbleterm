package terminal

type Option func(*TermWindow)

var OptionWithCommand = func(cmd string) Option {
	return func(tw *TermWindow) {
		tw.command = cmd
	}
}

var OptionWithInitialSize = func(width, height int) Option {
	return func(tw *TermWindow) {
		tw.width = width
		tw.height = height
	}
}
