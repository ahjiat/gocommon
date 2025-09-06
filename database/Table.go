package Database

import (
	"gorm.io/gorm"
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

func (self tbl[T]) ToSql() string {
	if self.chainDB == nil { self.chainDB = self.DB.Model(new(T)) }
	sql := self.chainDB.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Scan(new(T))
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
	result := self.chainDB.Scan(data);
	if result.Error != nil { panic(result.Error) }
	if result.RowsAffected == 0 { return nil }
	return data
}

func (self tbl[T]) Rows() *[]T {
	data := new([]T)
	if self.chainDB == nil { self.chainDB = self.DB.Model(data) }
	result := self.chainDB.Scan(data);
	if result.Error != nil { panic(result.Error) }
	if result.RowsAffected == 0 { return nil }
	return data
}

func (self tbl[T]) Value() T {
	data := new(T)
	if self.chainDB == nil { self.chainDB = self.DB.Model(data) }
	result := self.chainDB.Scan(data);
	if result.Error != nil { panic(result.Error) }
	return *data
}

func (self tbl[T]) ValueOrDefault(def T) T {
	data := new(T)
	if self.chainDB == nil { self.chainDB = self.DB.Model(data) }
	result := self.chainDB.Scan(data);
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
