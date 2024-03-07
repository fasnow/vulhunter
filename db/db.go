package db

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"sync"
)

var (
	instance *gorm.DB
	once     sync.Once
)

func GetSingleton() *gorm.DB {
	// 通过 sync.Once 确保仅执行一次实例化操作
	once.Do(func() {
		db, err := gorm.Open(sqlite.Open("db.db"), &gorm.Config{})
		if err != nil {
			panic(fmt.Errorf("failed to connect database:%s", err.Error()))
		}
		instance = db
		//if err = instance.AutoMigrate(&BaseModel{}); err != nil {
		//	panic(err.Error())
		//}
		if err = instance.AutoMigrate(&GithubCVE{}, &AVD{}); err != nil {
			panic(err.Error())
		}
	})
	return instance
}
