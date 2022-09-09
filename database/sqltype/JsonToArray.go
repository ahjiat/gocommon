package sqltype

import (
	"fmt"
	"database/sql/driver"
	"encoding/json"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type JsonToArray[T any] []T

func (self *JsonToArray[T]) Scan(value interface{}) error {
	switch value.(type) {
		case []uint8:
			b, ok := value.([]byte); if ! ok {
				panic("2 Gorm Custom Json[T] conversion byte error")
			}
			dat := new([]T)
			if err := json.Unmarshal(b, dat); err != nil {
				panic(err)
			}
			*self = *dat
		default:
			panic(fmt.Sprintf("Gorm Custom Json[T] unknow type %T", value))
	}
	return nil
}

func (self JsonToArray[T]) Value() (driver.Value, error) {
	// for create/update
	if len(self) == 0 { return []byte("[]"), nil }
	b, err := json.Marshal(self); if err != nil {
		panic(err)
	}
	return b, nil
}

// must tell gorm its type is text or unsupport data type will raise
func (JsonToArray[T]) GormDataType() string {
  return "text"
}
func (JsonToArray[T]) GormDBDataType(db *gorm.DB, field *schema.Field) string {
  return "text"
}
