package sqlfunc

import (
	"fmt"
	"reflect"
	"common/utility/pool"
)

type Extend[T any] struct {
	records      []T
	conncurrency int
}

func New[T any](records []T) *Extend[T] {
	return &Extend[T]{records: records}
}

func (b Extend[T]) IsEmpty() bool {
	return len(b.records) == 0
}

func (b Extend[T]) Parallel(conncurrency int) *Extend[T] {
	b.conncurrency = conncurrency
	return &b
}

func (b Extend[T]) First() T {
	var zero T
	if len(b.records) == 0 {
		return zero
	}
	return b.records[0]
}

func (b Extend[T]) All() []T {
	return b.records
}

func (b Extend[T]) Count() int {
	return len(b.records)
}

func (b Extend[T]) ForEach(fn any, args ...any) {
	if len(b.records) == 0 {
		return
	}

	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic("ForEach expects a function")
	}
	ft := v.Type()

	// fn must accept T or *T (+ len(args) extras)
	if ft.NumIn() != 1+len(args) {
		panic(fmt.Sprintf("function must take %d params (T or *T + %d extra)", 1+len(args), len(args)))
	}

	// Pre-convert fixed args to the function's parameter types.
	fixedArgs := make([]reflect.Value, len(args))
	for i, a := range args {
		want := ft.In(i + 1) // skip first param (T or *T)
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

	// Take an addressable view of the slice so elements are addressable (for *T).
	// Using &b.records here is fine even though b is a value receiver: the slice header
	// is copied, but it still points to the same backing array, so element updates
	// (when using *T) will modify the underlying data as expected.
	sliceV := reflect.ValueOf(&b.records).Elem() // []T (addressable)
	if sliceV.Kind() != reflect.Slice {
		panic("records must be a slice")
	}

	callOne := func(elem reflect.Value) {
		firstParam := ft.In(0)

		var firstArg reflect.Value
		switch {
		case elem.Type().AssignableTo(firstParam): // T -> T
			firstArg = elem
		case elem.CanAddr() && elem.Addr().Type().AssignableTo(firstParam): // T -> *T
			firstArg = elem.Addr()
		case elem.Kind() == reflect.Ptr && elem.Elem().Type().AssignableTo(firstParam): // *T -> T
			firstArg = elem.Elem()
		case elem.Kind() == reflect.Ptr && elem.Elem().Type().ConvertibleTo(firstParam): // *T -> T'
			firstArg = elem.Elem().Convert(firstParam)
		case elem.Type().ConvertibleTo(firstParam): // T -> T'
			firstArg = elem.Convert(firstParam)
		default:
			panic(fmt.Sprintf("element %v not assignable to first param %v", elem.Type(), firstParam))
		}

		call := make([]reflect.Value, 0, 1+len(fixedArgs))
		call = append(call, firstArg)
		call = append(call, fixedArgs...)
		v.Call(call)
	}

	// Iterate and call (optionally in parallel)
	if b.conncurrency > 1 {
		p := pool.New(b.conncurrency)
		for i := 0; i < sliceV.Len(); i++ {
			p.Go(func(i int) {
				elem := sliceV.Index(i)
				callOne(elem)
			}, i)
		}
		p.Wait()
	} else {
		for i := 0; i < sliceV.Len(); i++ {
			elem := sliceV.Index(i)
			callOne(elem)
		}
	}
}
