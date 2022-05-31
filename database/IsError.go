package Database

import (
	"errors"
	"github.com/go-sql-driver/mysql"
)

const (
	ErrDuplicated = 1062
)

func IsError(f func(), errNum int) bool {
	var rtnVal int
	errorHandle(f, &rtnVal)
	if errNum != rtnVal { return false }
	return true
}

func errorHandle(f func(), rtnVal *int) bool {
	defer func() {
		if errmsg := recover(); errmsg != nil {
			err := errmsg.(error)
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) {
				*rtnVal = int(mysqlErr.Number)
			}
		}
	}()
	f()
	return true
}
