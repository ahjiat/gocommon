package method

import (
	"runtime"
	"fmt"
	"strconv"
)

func DisplayInformation() {
	pcs := make([]uintptr, 20)
	n := runtime.Callers(0, pcs)
	pcs = pcs[:n]
	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		if !more { break }
		fmt.Printf("Function -> %v, Ptr -> %v\n", frame.Function, frame.PC)
	}
}

func Name(skip ...int) string {
	var skipIdx int
	if len(skip) == 0 { skipIdx = 1 } else { skipIdx = skip[0] }
	pc, _, _, ok := runtime.Caller(skipIdx)
	if ! ok { panic("runtime.Caller fail to call") }
	details := runtime.FuncForPC(pc)
	if details == nil { panic("runtime.FuncForPC return nil") }
	return details.Name()
}

func AppendName(v interface{}, skip ...int) string {
	var name string
	switch v.(type) {
		case int:
			name = strconv.Itoa(v.(int))
		case string:
			name = v.(string)
		case nil:
			name = ""
		default:
			panic(fmt.Sprintf("%v access param int/string", Name()))
	}
	if len(skip) == 0 {
		return Name(2) + "-" + name
	}
	return Name(skip...) + "-" + name
}

func Address(skip ...int) uintptr {
	var skipIdx int
	if len(skip) == 0 { skipIdx = 1 } else { skipIdx = skip[0] }
	pc, _, _, ok := runtime.Caller(skipIdx)
	if ! ok { panic("runtime.Caller fail to call") }
	details := runtime.FuncForPC(pc)
	if details == nil { panic("runtime.FuncForPC return nil") }
	return details.Entry()
}
