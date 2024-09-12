package v1

import (
	"app/model"
	"fmt"
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
)

// func ShowVccBalance(c *gin.Context) {

// 	var req model.Transaction
// 	// 使用BindJSON方法解析请求体到req变量中
// 	if err := c.BindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"code": 500,
// 			"data": "",
// 			"msg":  err.Error(),
// 		})
// 		return
// 	}
// 	Balance, err := model.CalVccBalance(req.CardNumber)
// 	formattedBalance := fmt.Sprintf("%.3f", Balance)
// 	c.JSON(
// 		http.StatusOK, gin.H{
// 			"code": 200,
// 			"data": formattedBalance,
// 			"msg":  err,
// 		},
// 	)
// }

// func ShowVccDeplete(c *gin.Context) {

// 	var req model.Transaction
// 	// 使用BindJSON方法解析请求体到req变量中
// 	if err := c.BindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"code": 500,
// 			"data": "",
// 			"msg":  err.Error(),
// 		})
// 		return
// 	}
// 	Deplete, err := model.CalVccTotalDeplete(req.CardNumber)

// 	c.JSON(
// 		http.StatusOK, gin.H{
// 			"code": 200,
// 			"data": Deplete,
// 			"msg":  err,
// 		},
// 	)
// }

type PaginationResult struct {  
    CurrentPage map[string]struct{ Balance, Deplete float64 }  
}  

func ShowVccBalanceAndDeplete(c *gin.Context) {
	pageSize, _ := strconv.Atoi(c.Query("pagesize"))
	pageNum, _ := strconv.Atoi(c.Query("pagenum"))
	startTime, _ := strconv.Atoi(c.Query("start_time"))
	endTime, _ := strconv.Atoi(c.Query("end_time"))
	fb_id := c.Query("account")
	id := c.Query("id")
	var IDs []string
	if fb_id != "" && id == ""{
		_, IDs, _ = model.ShowFBID(fb_id)
		
	} else if fb_id != "" && id != ""{
		// IDs, _ = model.ShowVccID()
		IDs = append([]string{id})
	}
	
	paginationResult, err, total := model.ShowVccBalanceAndDepletes(fb_id, IDs, pageSize, pageNum, startTime, endTime)
	
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"data": paginationResult,
			"msg":  err,
			"total": total,
		})
	
}

func ShowVccID(c *gin.Context) {
	IDs, err := model.ShowVccID()
	if err != nil {
		c.JSON(
			http.StatusBadRequest, gin.H{
				"code": 500,
				"data": "",
				"msg":  err,
			},
		)
	} else {
		c.JSON(
			http.StatusOK, gin.H{
				"code": 200,
				"data": IDs,
				"msg":  err,
			},
		)
	}

}

func ShowFBID(c *gin.Context) {
	Account := c.Query("account")
	FBIDs, VCCIDs ,err := model.ShowFBID(Account)
	if err != nil {
		c.JSON(
			http.StatusBadRequest, gin.H{
				"code": 500,
				"data": "",
				"msg":  err,
			},
		)
	} else {
		c.JSON(
			http.StatusOK, gin.H{
				"code": 200,
				"data1": FBIDs,
				"data2": VCCIDs,
				"msg":  err,
			},
		)
	}

}

type VccDepleteRequest struct {
	Year       int    `json:"year"`
	Month      int    `json:"month"`
	CardNumber string `json:"card_number"` // 注意这里使用了card_number而不是vccid
}

func ShowVccDepleteByDate(c *gin.Context) {
	var req VccDepleteRequest
	// 使用BindJSON方法解析请求体到req变量中
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 500,
			"data": "",
			"msg":  err.Error(),
		})
		return
	}

	Deplete, err := model.CalVccDepleteByDate(req.Year, req.Month, req.CardNumber)
	formattedDeplete := fmt.Sprintf("%.3f", Deplete)
	c.JSON(
		http.StatusOK, gin.H{
			"code": 200,
			"data": formattedDeplete,
			"msg":  err,
		},
	)
}
