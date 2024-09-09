package main

import (
	"app/route"
	"app/model"
)

func main() {
	// 引用数据库
	model.InitDb()
	// 引入路由组件
	route.InitRouter()

}