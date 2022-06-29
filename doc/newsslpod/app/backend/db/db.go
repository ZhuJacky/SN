// Package db provides ...
package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"mysslee_qcloud/app/backend/prom"
	"mysslee_qcloud/config"
	"mysslee_qcloud/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

const DEFAULT_PLAN_RESOURCE = "sslpod-00000000001"

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

	// Migrate the schema
	gormDB.AutoMigrate(model.Account{})
	gormDB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").AutoMigrate(model.DomainResult{})
	gormDB.AutoMigrate(model.AccountDomain{})
	gormDB.AutoMigrate(model.NoticeInfo{})
	gormDB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").AutoMigrate(model.NoticeMsg{})
	gormDB.AutoMigrate(model.DomainClaim{})
	gormDB.AutoMigrate(model.DashboardResult{})
	gormDB.AutoMigrate(model.DomainCert{})
	gormDB.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").AutoMigrate(model.CertInfo{})
	gormDB.AutoMigrate(model.TagDomain{})
	gormDB.AutoMigrate(model.LimitConsume{})
	gormDB.AutoMigrate(model.DomainRegionalResult{})
	gormDB.AutoMigrate(model.DomainIps{})
	// qcloud
	gormDB.AutoMigrate(model.Resource{})
	gormDB.AutoMigrate(model.Product{})
	gormDB.AutoMigrate(model.Order{})

	initData()

	// go timer()
}

func initData() {
	// init product
	data, err := ioutil.ReadFile(config.ProductPath)
	if err != nil {
		panic(err)
	}
	var products []*model.Product
	err = json.Unmarshal(data, &products)
	if err != nil {
		panic(err)
	}
	for _, v := range products {
		v.Content = v.ContentRawMessage
		if IsExistProduct(v.Id) {
			continue
		}
		// 添加产品
		err = AddProduct(v)
		if err != nil {
			panic(err)
		}
	}

	// init default plan resource
	if !IsExistResource(DEFAULT_PLAN_RESOURCE) {
		detail := []byte(fmt.Sprintf(`{"pid":%d}`, config.Conf.Backend.BasicPlan))
		resource := &model.Resource{
			ResourceId:        DEFAULT_PLAN_RESOURCE,
			Uin:               "ALL",
			AppId:             0,
			ProjectId:         0,
			RenewFlag:         0,
			Region:            0,
			ZoneId:            0,
			Status:            1,
			PayMode:           1,
			IsolatedTimestamp: model.TimeZeroAt,
			CreateTime:        time.Now(),
			ExpireTime:        model.TimeZeroAt,
			GoodsDetail:       detail,
		}
		err := AddResourceWithoutFlowId(resource)
		if err != nil {
			panic(err)
		}
	}
}

func timer() {
	t := time.NewTicker(time.Minute)
	for range t.C {

		// 账户数
		total, err := GetAccountCount()
		if err != nil {
			logrus.Error("GetAccountCount ", err)
			continue
		}
		prom.PromAccountCount.Set(float64(total))
		// 站点数
		total, err = GetMonitorSiteCount()
		if err != nil {
			logrus.Error("GetAccountCount ", err)
			continue
		}
		prom.PromMonitorSiteCount.Set(float64(total))
		// 购买套餐数
		plans, err := GetResourcePlanCount()
		if err != nil {
			logrus.Error("GetResourcePlanCount ", err)
			continue
		}
		for k, v := range plans {
			prom.PromBoughtPlan.WithLabelValues(k).Set(float64(v))
		}

		// db ping
		err = gormDB.DB().Ping()
		if err != nil {
			prom.PromDBError.Inc()
			logrus.Error("Ping: ", err)
			continue
		}
	}
}
