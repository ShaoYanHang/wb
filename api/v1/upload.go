package v1

import (
	"app/model"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"os"
)

func Upload1(c *gin.Context) {
	file, err := c.FormFile("f1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"data": "",
			"msg":  err.Error()})
	}
	if file.Filename[len(file.Filename)-4:] != "xlsx" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"data": "",
			"msg":  "非xlsx格式"})
	}
	dst := fmt.Sprintf("files1/%s", file.Filename)
	// 上传文件到指定的目录
	err = c.SaveUploadedFile(file, dst)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"data": "",
			"msg":  err.Error()})
	} else {
		err1 := model.ImportTransactionsFromXLSX(dst) // 假设 db 是你的数据库连接
		if err1 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"data": "",
				"msg":  err})
		}else {
			c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": "",
			"msg":  "成功导入虚拟卡文件"})
		}
		_ = os.Remove(dst)  
	}
}

func Upload2(c *gin.Context) {
	file, err := c.FormFile("f1")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"data": "",
			"msg":  err.Error()})
	}
	if file.Filename[len(file.Filename)-3:] != "csv" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"data": "",
			"msg":  "非csv格式"})
	}
	dst := fmt.Sprintf("files2/%s", file.Filename)
	// 上传文件到指定的目录
	err = c.SaveUploadedFile(file, dst)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"data": "",
			"msg":  err.Error()})
	} else {
		err1 := model.ImportTransactionRecordFromCSV(dst) // 假设 db 是你的数据库连接
		if err1 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"data": "",
				"msg":  err1.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": "",
			"msg":  "成功导入FB文件"})
		}
		_ = os.Remove(dst)
	}
}
