package sqlfunc

import (
	"fmt"
	"reflect"
	"common/utility/pool"
)

type Extend[T any] struct {
	Records      *T
	conncurrency int
}

func (b Extend[T]) IsEmpty() bool {
	if b.Records == nil {
		return true
	}
	rv := reflect.ValueOf(*b.Records)
	if rv.Kind() == reflect.Slice {
		return rv.Len() == 0
	}
	// Single value present.
	return false
}

func (b Extend[T]) Parallel(conncurrency int) *Extend[T] {
	b.conncurrency = conncurrency
	return &b
}

// ForEach calls fn(element, args...) over each element.
// Works for both T (single) and []T (slice).
func (b Extend[T]) ForEach(fn any, args ...any) {
	if b.Records == nil {
		return
	}

	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic("ForEach expects a function")
	}
	ft := v.Type()

	if ft.NumIn() != 1+len(args) {
		panic(fmt.Sprintf("function must take %d params (T + %d extra)", 1+len(args), len(args)))
	}

	// Pre-convert fixed args to the function's parameter types.
	fixedArgs := make([]reflect.Value, len(args))
	for i, a := range args {
		want := ft.In(i + 1)
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

	val := reflect.ValueOf(*b.Records)
	firstParam := ft.In(0)

	callOne := func(elem reflect.Value) {
		var firstArg reflect.Value
		switch {
		case elem.Type().AssignableTo(firstParam):
			firstArg = elem
		case elem.CanAddr() && elem.Addr().Type().AssignableTo(firstParam):
			firstArg = elem.Addr()
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

	if val.Kind() == reflect.Slice {
		n := val.Len()
		if n == 0 {
			return
		}
		if b.conncurrency > 1 {
			p := pool.New(b.conncurrency)
			for i := 0; i < n; i++ {
				p.Go(func(i int) {
					callOne(val.Index(i))
				}, i)
			}
			p.Wait()
		} else {
			for i := 0; i < n; i++ {
				callOne(val.Index(i))
			}
		}
		return
	}

	// Single value
	callOne(val)
}
