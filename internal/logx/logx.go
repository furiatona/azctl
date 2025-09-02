package logx

import (
	"fmt"
	"log"
	"os"
)

var verbose bool

func Init(v bool) {
	verbose = v
	log.SetFlags(0)
}

func Infof(format string, args ...any) {
	if verbose {
		log.Printf("[INFO] "+format, args...)
	}
}

// Printf always prints the message regardless of verbose flag
func Printf(format string, args ...any) {
	log.Printf("[INFO] "+format, args...)
}

func Warnf(format string, args ...any) {
	log.Printf("[WARN] "+format, args...)
}

func Errorf(format string, args ...any) error {
	msg := fmt.Sprintf("[ERROR] "+format, args...)
	fmt.Fprintln(os.Stderr, msg)
	return fmt.Errorf("[ERROR] "+format, args...)
}
