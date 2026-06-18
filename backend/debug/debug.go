package debug

import (
	"log"
	"os"
)

var debugEnabled bool

func init() {
	debugEnabled = os.Getenv("SSC_DEBUG_ENABLED") == "true"
}

func Log(format string, v ...interface{}) {
	if debugEnabled {
		log.Printf("[DEBUG] "+format, v...)
	}
}

func Println(v ...interface{}) {
	if debugEnabled {
		args := append([]interface{}{"[DEBUG]"}, v...)
		log.Println(args...)
	}
}

func IsEnabled() bool {
	return debugEnabled
}
