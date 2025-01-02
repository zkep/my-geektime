package service

import _ "embed"

//go:embed mkdocs.tpl
var MkdocsYML string

type Mkdocs struct {
	SiteName string
	Navs     []Nav
}

type Nav struct {
	Name  string
	Items []string
}
