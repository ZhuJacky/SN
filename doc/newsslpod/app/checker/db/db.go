// Package db provides ...
package db

import (
	"time"

	"mysslee_qcloud/app/checker/prom"
	"mysslee_qcloud/brand"
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

	// init brand database
	brand.Init(gormDB.DB())

	//go timer()
}

func timer() {
	t := time.NewTicker(time.Minute)
	for range t.C {
		err = gormDB.DB().Ping()
		if err != nil {
			prom.PromDBError.Inc()
			logrus.Error("Ping: ", err)
			continue
		}
	}
}
