package v1

import (
	"app/model"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

func ShowFile1(c *gin.Context) {
	// 假设db.DB是你在db包中初始化的*gorm.DB实例
	var result []model.Transaction
	pageSize, _ := strconv.Atoi(c.Query("pagesize"))
	pageNum, _ := strconv.Atoi(c.Query("pagenum"))
	cardNumber  := c.Query("card_number")
	transactionType  := c.Query("transaction_type")
	startTime = strconv.Atoi(c.Query("start_time"))
	endTime = strconv.Atoi(c.Query("end_time"))
	switch {
	case pageSize >= 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}

	if pageNum == 0 {
		pageNum = 1
	}

	result, err, total := model.GetTransactions(pageSize, pageNum, cardNumber, transactionType, startTime ,endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"data": "",
			"msg":  "Failed to retrieve transactions",
			"total": 0})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": result,
		"msg":  "",
		"total": total})
}


func ShowFile2(c *gin.Context) {
	// 假设db.DB是你在db包中初始化的*gorm.DB实例
	var result []model.TransactionRecord
	pageSize, _ := strconv.Atoi(c.Query("pagesize"))
	pageNum, _ := strconv.Atoi(c.Query("pagenum"))
	Account := c.Query("account")
	PaymentMethod := c.Query("payment_method")
	switch {
	case pageSize >= 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}

	if pageNum == 0 {
		pageNum = 1
	}

	result, err, total := model.GetTransactionRecords(pageSize, pageNum, Account, PaymentMethod)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code": "500",
				"data": "",
				"msg": "Failed to retrieve transactions",
				"total": 0})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{
			"code": 200,
			"msg": err,
			"data": result,
			"tatal": total})
}

func ShowVirtualCardDataByaccount(c *gin.Context) {
	var req model.TransactionRecord
	// 使用BindJSON方法解析请求体到req变量中
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 500,
			"data": "Invalid JSON body",
			"msg":  err.Error(),
		})
		return
	}
	result, err := model.CalFBbyaccount(req.Account, req.PaymentMethod)

	c.JSON(
		http.StatusOK, gin.H{
			"status": 200,
			"data":   result,
			"msg":    err,
		},
	)
}

func ShowFBDataByaccountList(c *gin.Context) {
	var req model.TransactionRecord
	// 使用BindJSON方法解析请求体到req变量中
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 500,
			"data": "Invalid JSON body",
			"msg":  err.Error(),
		})
		return
	}
	result, err := model.CalFBbyaccountList(req.Account, req.PaymentMethod)

	c.JSON(
		http.StatusOK, gin.H{
			"code": 200,
			"data": result,
			"msg":  err,
		},
	)
}

func Showfb_vccdata(c *gin.Context) {
	var req model.TransactionRecord
	// 使用BindJSON方法解析请求体到req变量中
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 500,
			"data": "Invalid JSON body",
			"msg":  err.Error(),
		})
		return
	}
	result, err := model.Showfb_vccdata(req.Account)

	c.JSON(
		http.StatusOK, gin.H{
			"code": 200,
			"data": result,
			"msg":  err,
		},
	)
}

func UpdateTransactionRecord(c *gin.Context) {
	var req model.TransactionRecord
	// 使用BindJSON方法解析请求体到req变量中
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 500,
			"data": "Invalid JSON body",
			"dsg":  err.Error(),
		})
		return
	}
	err := model.UpdateTransactionRecord(req.TransactionID, req.IsTicked, req.Note)

	c.JSON(
		http.StatusOK, gin.H{
			"code": 200,
			"data": "",
			"dsg":  err,
		},
	)
}
