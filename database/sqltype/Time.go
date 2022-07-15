package sqltype

import (
	"fmt"
	"database/sql/driver"
	"time"
)

type Time struct {
	time.Time
}

func (self *Time) Scan(value interface{}) error {
	switch value.(type) {
		case time.Time:
			v, ok := value.(time.Time)
			if ! ok { panic("Gorm Custom Time conversion time.Time error") }
			*self = Time{v}
		// for create only, cos Value() return string -> "0001-01-01" due to ZeroTime
		case string:
			v, ok := value.(string); if ! ok {
				panic("Gorm Custom Time conversion string error")
			}
			d, err := time.Parse("2006-01-02 15:04:05", v); if err != nil {
				panic(fmt.Sprintf("Gorm Custom Time time.Parse[%v], value -> %s", err, value))
			}
			*self = Time{d}
		default:
			panic(fmt.Sprintf("Gorm Custom Time unknow type %T", value))
	}
	return nil
}

func (t Time) Value() (driver.Value, error) {
	// for create/update
	if t.Time.IsZero() { return "0001-01-01 00:00:00", nil }
	return t.Time, nil
}
