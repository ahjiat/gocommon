package Database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"fmt"
	//"time"
)

type DB = gorm.DB

var dbContainer = map[string]*DB{}

func AddDBConnection(sessionName string, dsn string) {
	if _, found := dbContainer[sessionName]; found {
		panic(fmt.Sprintf("AddDBConnection sessionName [%v] exists", sessionName))
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger: logger.Default.LogMode(logger.Silent),
	}); if err != nil { panic(err) }
	//sqlDB, err := db.DB()
	//sqlDB.SetMaxIdleConns(20)
	//sqlDB.SetMaxOpenConns(100)
	//sqlDB.SetConnMaxLifetime(time.Hour)
	dbContainer[sessionName] = db
}
func GetSession(sessionName string) *DB {
	return dbContainer[sessionName]
}
