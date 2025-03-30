package utils

import (
	"fmt"
)

func TypewriterEffect(text string, done bool) {
	fmt.Print(text)
	if done {
		fmt.Println()
	}
}

var DEBUG = false

func DebugPrintln(e ...any) {
	if DEBUG {
		fmt.Println(e...)
	}
}

func DebugPrintf(f string, e ...any) {
	if DEBUG {
		fmt.Printf(f, e...)
	}
}
