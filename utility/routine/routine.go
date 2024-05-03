package routine

import (
	"fmt"
)

func NoWait(executeHandle func(...any), panicHandle func(error, ...any), args ...any) {
	go func(){
		defer func() {
			if r := recover(); r != nil {
				if panicHandle != nil { panicHandle(fmt.Errorf("panic occurred: %v", r), args...) }
			}
		}()
		executeHandle(args...)
	}()
}
