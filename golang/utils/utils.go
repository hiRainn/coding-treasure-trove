package utils

import "runtime"

func recoverFunc() {
	if err := recover(); err != nil {
		stack := make([]byte, 1024*8)
		stack = stack[:runtime.Stack(stack, false)]
	}
}
