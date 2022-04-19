package Database

import (
	//"reflect"
)

type Instance func() *DB

func (f Instance) Add(data interface{}) int {
	//if reflect.ValueOf(data).Type().Kind() != reflect.Ptr { panic("params must be pointer") }
	result := f().Create(data)
	if result.Error != nil { panic(result.Error) }
	return int(result.RowsAffected)
}

func (f Instance) Save(data interface{}) int64 {
	//if reflect.ValueOf(data).Type().Kind() != reflect.Ptr { panic("params must be pointer") }
	result := f().Save(data);_ = result
	if result.Error != nil { panic(result.Error) }
	return int64(result.RowsAffected)
}
