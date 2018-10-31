package utils

import "log"

// ErrHandlePrintln log error
func ErrHandlePrintln(err error, msg string) {
	if err != nil {
		log.Println(msg, err)
	}
}

// ErrHandleFatalln log error and quit
func ErrHandleFatalln(err error, msg string) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
