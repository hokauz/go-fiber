package logx

// Fatal logs the message with the error and exits with code 1.
func Fatal(msg string, err error) {
	Default().Fatal(msg, err)
}
