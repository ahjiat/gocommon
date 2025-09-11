package Database

import (
	"gorm.io/gorm"
	"common/database/sqlfunc"
)

func Query[T any](db *gorm.DB, sql string, values ...interface{}) *T {
	data := new(T)
	result := db.Raw(sql, values...).Scan(data);
	if result.Error != nil { panic(result.Error) }
	if result.RowsAffected == 0 { return nil }
	return data
}

func QueryToSql[T any](db *gorm.DB, sql string, values ...interface{}) string {
	result := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		data := new(T)
		return tx.Raw(sql, values...).Scan(data);
	})
	return result
}

func Execute[T any](db *gorm.DB, sql string, values ...interface{}) *sqlfunc.Extend[T] {
	data := new([]T) // *[]T
	result := db.Raw(sql, values...).Scan(data)
	if result.Error != nil { panic(result.Error) }
	return sqlfunc.New[T](*data)
}
