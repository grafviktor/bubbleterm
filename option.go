package termview

type Option func(*Model)

func WithCommand(cmd string, args ...string) Option {
	return func(tw *Model) {
		tw.command = cmd
		tw.commandArgs = append([]string(nil), args...)
	}
}

func WithInitialWidth(width int) Option {
	return func(tw *Model) {
		tw.width = width
	}
}

func WithInitialHeight(height int) Option {
	return func(tw *Model) {
		tw.height = height
	}
}

func WithClosedMessage(message string) Option {
	return func(tw *Model) {
		tw.closedMessage = message
	}
}
