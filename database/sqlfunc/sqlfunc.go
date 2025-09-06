package sqlfunc

import (
	"fmt"
	"reflect"
)

type Extend[T any] struct {
	Records *[]T
}

func (b Extend[T]) IsEmpty() bool { return b.Records == nil }

// ForEach calls fn(element, args...) for each element of b.Records.
// fn must be a function whose first parameter is compatible with T (or *T),
// followed by parameters matching the types of args...
func (b Extend[T]) ForEach(fn any, args ...any) {
	if b.Records == nil || len(*b.Records) == 0 {
		return
	}

	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic("ForEach expects a function")
	}
	ft := v.Type()

	// fn must accept T (+ len(args) extras)
	if ft.NumIn() != 1+len(args) {
		panic(fmt.Sprintf("function must take %d params (T + %d extra)", 1+len(args), len(args)))
	}

	// Pre-convert fixed args to the function's parameter types.
	fixedArgs := make([]reflect.Value, len(args))
	for i, a := range args {
		want := ft.In(i + 1) // skip first param (T)
		var av reflect.Value
		if a == nil {
			av = reflect.Zero(want)
		} else {
			av = reflect.ValueOf(a)
			if !av.Type().AssignableTo(want) {
				if av.Type().ConvertibleTo(want) {
					av = av.Convert(want)
				} else {
					panic(fmt.Sprintf("arg %d: %v not assignable/convertible to %v", i, av.Type(), want))
				}
			}
		}
		fixedArgs[i] = av
	}

	// Iterate and call.
	for i := range *b.Records {
		elem := reflect.ValueOf((*b.Records)[i])
		firstParam := ft.In(0)

		var firstArg reflect.Value
		switch {
		case elem.Type().AssignableTo(firstParam):
			firstArg = elem
		case elem.CanAddr() && elem.Addr().Type().AssignableTo(firstParam):
			firstArg = elem.Addr() // supports callbacks wanting *T
		case elem.Type().ConvertibleTo(firstParam):
			firstArg = elem.Convert(firstParam)
		default:
			panic(fmt.Sprintf("element %v not assignable to first param %v", elem.Type(), firstParam))
		}

		call := make([]reflect.Value, 0, 1+len(fixedArgs))
		call = append(call, firstArg)
		call = append(call, fixedArgs...)
		v.Call(call)
	}
}
