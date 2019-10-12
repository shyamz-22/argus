package exceptions

import "log"

// To use in defer blocks where, errors does not affect Business as usual
// Ex: APM errors
func Log(log *log.Logger, f func() error) {
	if err := f(); err != nil {
		log.Printf("Error: %v", err)
	}
	return
}

// To use in defer blocks where, errors affect Business as usual
// Ex: Reading config files
func LogFatalError(log *log.Logger, message string, f func() error) {
	if err := f(); err != nil {
		log.Fatalf("Error: %s: %v", message, err)
	}
}
