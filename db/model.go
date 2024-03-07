package db

import (
	"cveHunter/entry"
	"gorm.io/gorm"
)

type BaseModel struct {
	gorm.Model
	SID int64
}

type GithubCVE struct {
	BaseModel
	Name        string
	Author      string
	HtmlUrl     string
	Description string
}

type AVD struct {
	BaseModel
	entry.AVD
}
