package Database

import (
	"errors"
	"github.com/go-sql-driver/mysql"
)

const (
	ErrDuplicated = 1062
)

func IsError(f func()) (int, error) {
	var rtnVal int
	rtnErr := new(error)
	errorHandle(f, &rtnVal, rtnErr)
	return rtnVal, *rtnErr
}

func errorHandle(f func(), rtnVal *int, rtnErr *error) bool {
	defer func() {
		if errmsg := recover(); errmsg != nil {
			*rtnErr = errmsg.(error)
			var mysqlErr *mysql.MySQLError
			if errors.As(*rtnErr, &mysqlErr) {
				*rtnVal = int(mysqlErr.Number)
			}
		}
	}()
	f()
	return true
}
