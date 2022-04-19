package Database

import (
	"gorm.io/gorm"
)

func Query[T any](db *gorm.DB, sql string, values ...interface{}) *T {
	data := new(T)
	result := db.Raw(sql, values...).Scan(data);
	if result.Error != nil { panic(result.Error) }
	if result.RowsAffected == 0 { return nil }
	return data
}
