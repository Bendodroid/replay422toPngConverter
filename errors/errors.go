package errors

import (
	"log"
)

// Check panics with an error message if err != nil
func Check(err error, logMsg string) {
	if err != nil {
		log.Panicf("[ERROR] %s\n%s", logMsg, err)
	}
}
