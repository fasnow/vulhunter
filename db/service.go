package db

import (
	"github.com/yitter/idgenerator-go/idgen"
	"gorm.io/gorm"
	"sync"
)

type Service struct {
	db *gorm.DB
}

var (
	serviceInstance *Service
	serviceOnce     sync.Once
)

func GetServiceSingleton() *Service {
	// 通过 sync.Once 确保仅执行一次实例化操作
	serviceOnce.Do(func() {
		serviceInstance = &Service{db: GetSingleton()}
	})
	return serviceInstance
}

func (r *Service) Insert(cve GithubCVE) error {
	cve.SID = idgen.NextId()
	if err := r.db.Create(&cve).Error; err != nil {
		return err
	}
	return nil
}

func (r *Service) FindByCVEName(name string) ([]GithubCVE, error) {
	var items []GithubCVE
	if err := r.db.Where("name = ?", name).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Service) FindAllByAuthorAndCVE(author, cve string) ([]GithubCVE, error) {
	var items []GithubCVE
	if err := r.db.Where("author = ? AND name = ?", author, cve).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Service) FindAuthorsByCVE(cve string) []string {
	var authors []string
	r.db.Model(&GithubCVE{}).Select("DISTINCT author").Where("name = ?", cve).Find(&authors)
	return authors
}
