// Package db provides ...
package db

import (
	"mysslee_qcloud/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

var (
	gormDB *gorm.DB
	err    error
)

func Init() {
	// postgres
	gormDB, err = gorm.Open(config.Conf.Database.Driver, config.Conf.Database.Source)
	if err != nil {
		panic(err)
	}
	if config.Conf.OrmDebug {
		gormDB = gormDB.Debug()
	}
	gormDB.SetLogger(logrus.StandardLogger())
}
