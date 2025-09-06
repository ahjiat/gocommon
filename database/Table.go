package Database

import (
	"gorm.io/gorm"
	"reflect"
	"fmt"
	"common/database/sqlfunc"
	"common/utility/pool"
)

type tblInterface interface {
	TableName() string
}

func Table[T tblInterface](db *gorm.DB) *tbl[T] {
	return &tbl[T]{DB:db}
}

type tbl[T tblInterface] struct {
	DB *gorm.DB
	chainDB *gorm.DB
	conncurrency int
}

func (self tbl[T]) Select(query interface{}, args ...interface{}) *tbl[T] {
	if self.chainDB == nil { self.chainDB = self.DB.Model(new(T)) }
	self.chainDB = self.chainDB.Select(query, args...)
	return &self
}

func (self tbl[T]) Where(query interface{}, args ...interface{}) *tbl[T] {
	if self.chainDB == nil { self.chainDB = self.DB.Model(new(T)) }
	self.chainDB = self.chainDB.Where(query, args...)
	return &self
}

func (self tbl[T]) Limit(value int) *tbl[T] {
	if self.chainDB == nil { self.chainDB = self.DB.Model(new(T)) }
	self.chainDB = self.chainDB.Limit(value)
	return &self
}

func (self tbl[T]) OrderBy(values ...string) *tbl[T] {
	if self.chainDB == nil { self.chainDB = self.DB.Model(new(T)) }
	for _, value := range values {
		self.chainDB = self.chainDB.Order(value)
	}
	return &self
}

func (self tbl[T]) Parallel(conncurrency int) *tbl[T] {
	if self.chainDB == nil { self.chainDB = self.DB.Model(new(T)) }
	self.conncurrency = conncurrency
	return &self
}

func (self tbl[T]) ToSql() string {
	if self.chainDB == nil { self.chainDB = self.DB.Model(new(T)) }
	sql := self.chainDB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(new(T))
	})
	return sql
}

func (self tbl[T]) Delete(emptyOrdata ...*T) int64 {
	var result *gorm.DB
	if self.chainDB != nil {
		// for chain
		result = self.chainDB
	} else {
		// for set default where condition using primary key "id"
		result = self.DB
	}
	if len(emptyOrdata) > 0 {
		// for set default where condition using primary key "id"
		result = result.Delete(emptyOrdata[0])
	} else {
		// for chain
		result = result.Delete(nil)
	}
	if result.Error != nil { panic(result.Error) }
	return result.RowsAffected
}

func (self tbl[T]) Update(data *T, col string, cols ...string) int64 {
	var result *gorm.DB
	if self.chainDB != nil {
		// for chain
		result = self.chainDB
	} else {
		// for set default where condition using primary key "id"
		result = self.DB
	}
	if len(cols) > 0 {
		result = result.Select(col, cols)
	} else {
		result = result.Select(col)
	}
	result.Updates(data)
	if result.Error != nil { panic(result.Error) }
	return result.RowsAffected
}

func (self tbl[T]) Row() *T {
	data := new(T)
	if self.chainDB == nil { self.chainDB = self.DB.Model(data) }
	result := self.chainDB.Take(data);
	if result.Error != nil { panic(result.Error) }
	if result.RowsAffected == 0 { return nil }
	return data
}

func (self tbl[T]) Rows() *[]T {
	data := new([]T)
	if self.chainDB == nil { self.chainDB = self.DB.Model(data) }
	result := self.chainDB.Find(data);
	if result.Error != nil { panic(result.Error) }
	if result.RowsAffected == 0 { return nil }
	return data
}

func (self tbl[T]) Execute() sqlfunc.Extend[T] {
	return sqlfunc.Extend[T]{Records: self.Rows()}
}

func (self tbl[T]) Value() T {
	data := new(T)
	if self.chainDB == nil { self.chainDB = self.DB.Model(data) }
	result := self.chainDB.Take(data);
	if result.Error != nil { panic(result.Error) }
	return *data
}

func (self tbl[T]) ValueOrDefault(def T) T {
	data := new(T)
	if self.chainDB == nil { self.chainDB = self.DB.Model(data) }
	result := self.chainDB.Take(data);
	if result.Error != nil { panic(result.Error) }
	if result.RowsAffected == 0 { *data = def }
	return *data
}

func (self tbl[T]) Save(data *T) *T {
	result := self.DB.Model(data).Save(data);
	if result.Error != nil { panic(result.Error) }
	return data
}

func (self tbl[T]) Add(data *T) *T {
	result := self.DB.Model(data).Create(data);
	if result.Error != nil { panic(result.Error) }
	return data
}

// ForEach iterates rows in batches and invokes a user callback via reflection.
// The callback must be a function whose first parameter is either T or *T.
// Any remaining parameters must be compatible with the supplied args.
func (self tbl[T]) ForEach(cb any, args ...any) {
	if self.chainDB == nil { self.chainDB = self.DB.Model(new(T)) }

	fnVal := reflect.ValueOf(cb)
	if fnVal.Kind() != reflect.Func {
		panic("ForEach: cb must be a function")
	}
	fnType := fnVal.Type()
	if fnType.NumIn() < 1 {
		panic("ForEach: cb must accept at least one parameter (T or *T)")
	}

	// Determine T and *T types
	var zeroPtr *T
	tPtr := reflect.TypeOf(zeroPtr)     // *T
	tElem := tPtr.Elem()                // T

	firstParam := fnType.In(0)
	wantPtr := false
	switch {
	case firstParam == tPtr:
		wantPtr = true
	case firstParam == tElem:
		wantPtr = false
	default:
		panic(fmt.Sprintf("ForEach: cb first param must be %v or %v, got %v", tElem, tPtr, firstParam))
	}

	// Verify extra param arity/types
	if fnType.NumIn() != 1+len(args) {
		panic(fmt.Sprintf("ForEach: cb wants %d args after the row param, but %d provided",
			fnType.NumIn()-1, len(args)))
	}
	coercedArgs := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		got := reflect.ValueOf(args[i])
		want := fnType.In(i + 1)

		// Handle nils for pointer/interface/slice/map/func types
		if !got.IsValid() {
			if want.Kind() == reflect.Interface ||
				want.Kind() == reflect.Pointer ||
				want.Kind() == reflect.Slice ||
				want.Kind() == reflect.Map ||
				want.Kind() == reflect.Func {
				coercedArgs[i] = reflect.Zero(want)
				continue
			}
			panic(fmt.Sprintf("ForEach: arg %d is nil but target type is %v", i, want))
		}

		if got.Type().AssignableTo(want) {
			coercedArgs[i] = got
		} else if got.Type().ConvertibleTo(want) {
			coercedArgs[i] = got.Convert(want)
		} else {
			panic(fmt.Sprintf("ForEach: arg %d type %v not assignable/convertible to %v", i, got.Type(), want))
		}
	}

	const batchSize = 1000
	var (
		batch []T
	)

	res := self.chainDB.FindInBatches(&batch, batchSize, func(tx *gorm.DB, _ int) error {
		if self.conncurrency > 1 {
			p := pool.New(self.conncurrency)
			for i := range batch {
				p.Go(func(i int){
					var first reflect.Value
					if wantPtr {
						first = reflect.ValueOf(&batch[i])
					} else {
						first = reflect.ValueOf(batch[i])
					}

					callArgs := make([]reflect.Value, 1+len(coercedArgs))
					callArgs[0] = first
					copy(callArgs[1:], coercedArgs)

					fnVal.Call(callArgs)
				}, i)
			}
			p.Wait()
		} else {
			for i := range batch {
				var first reflect.Value
				if wantPtr {
					first = reflect.ValueOf(&batch[i])
				} else {
					first = reflect.ValueOf(batch[i])
				}

				callArgs := make([]reflect.Value, 1+len(coercedArgs))
				callArgs[0] = first
				copy(callArgs[1:], coercedArgs)

				fnVal.Call(callArgs)
			}
		}
		return nil
	})
	if res.Error != nil { panic(res.Error) }
}


