package model

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type TransactionRecord struct {
	Account                string  `gorm:"type:varchar(20)" json:"account"`
	Date                   string  `gorm:"type:varchar(20)" json:"date"`
	TransactionID          string  `gorm:"type:varchar(100);primaryKey" json:"transaction_id"` // 假设Transaction ID是唯一的，可以用作主键
	PaymentMethod          string  `gorm:"type:varchar(100)" json:"payment_method"`
	Amount                 float64 `gorm:"type:decimal(10,2)" json:"amount"` // 假设金额是浮点数，可以调整精度
	Currency               string  `gorm:"type:varchar(10)" json:"currency"`
	IsTicked               bool    `gorm:"type:boolean" json:"is_ticked"`                // 假设这是一个布尔字段，表示交易是否授权
	IsTradingAuthorization bool    `gorm:"type:boolean" json:"is_trading_authorization"` // 假设这是一个布尔字段，表示交易是否授权
	Note                   string  `gorm:"type:varchar(500)" json:"note"`
}

type Transaction struct {
	TransactionID       string  `gorm:"type:varchar(50);primaryKey" json:"transaction_id"`
	TransactionTime     string  `gorm:"type:varchar(29)" json:"transaction_time"`
	CardNumber          string  `gorm:"type:varchar(200);sensitive" json:"card_number"` // 假设卡号需要特殊处理
	Nickname            string  `gorm:"type:varchar(100)" json:"nickname"`
	BillName            string  `gorm:"type:varchar(255)" json:"bill_name"`
	TransactionType     string  `gorm:"type:varchar(100)" json:"transaction_type"`
	OrderAmount         float64 `gorm:"type:decimal(10,2)" json:"order_amount"` // 假设金额为小数
	OrderCurrency       string  `gorm:"type:varchar(10)" json:"order_currency"`
	TransactionAmount   float64 `gorm:"type:decimal(10,2)" json:"transaction_amount"`
	TransactionFee      float64 `gorm:"type:decimal(10,2)" json:"transaction_fee"` // 假设费用也为小数
	TransactionCurrency string  `gorm:"type:varchar(10)" json:"transaction_currency"`
	TransactionStatus   string  `gorm:"type:varchar(100)" json:"transaction_status"`
	AuthorizationCode   string  `gorm:"type:varchar(100)" json:"authorization_code,omitempty"` // 如果可能为空，使用omitempty
	ResultCode          string  `gorm:"type:varchar(100)" json:"result_code"`
	ResultDescription   string  `gorm:"type:varchar(255)" json:"result_description"`
	SettlementStatus    string  `gorm:"type:varchar(100)" json:"settlement_status"`
	IsJudge             bool    `gorm:"type:boolean" json:"is_judge"`
}

type ByTransactionTime []Transaction

func (a ByTransactionTime) Len() int      { return len(a) }
func (a ByTransactionTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTransactionTime) Less(i, j int) bool {
	t1, _ := time.Parse("2024-08-22 19:36:34", a[i].TransactionTime)
	t2, _ := time.Parse("2024-08-22 19:36:34", a[j].TransactionTime)
	return t1.Before(t2)
}

func ImportTransactionsFromXLSX(filePath string) error {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return err
	}

	// 假设数据在第一张表的第二行开始（第一行为标题行）
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return err
	}

	if len(rows) <= 1 { // 至少应该有一行数据
		return errors.New("no data found in the Excel file")
	}

	// 跳过标题行
	transactions := make([]Transaction, 0, len(rows)-1)

	for _, row := range rows[1:] {
		if len(row) < 14 { // 确保有足够的数据列
			continue
		}

		// 解析行数据到 Transaction 结构
		var trans Transaction
		trans.TransactionID = row[0]
		trans.TransactionTime = row[1]
		trans.CardNumber = row[2][len(row[2])-4:]
		trans.Nickname = row[3]
		trans.BillName = row[4]
		trans.TransactionType = row[5]
		// 注意：这里可能需要处理浮点数和字符串的转换
		trans.OrderAmount, _ = strconv.ParseFloat(row[6], 64)
		trans.OrderCurrency = row[7]
		trans.TransactionAmount, _ = strconv.ParseFloat(row[8], 64)
		trans.TransactionFee, _ = strconv.ParseFloat(row[9], 64)
		trans.TransactionCurrency = row[10]
		trans.TransactionStatus = row[11]
		trans.AuthorizationCode = row[12]
		trans.ResultCode = row[13]
		trans.ResultDescription = row[14]
		trans.SettlementStatus = row[15]
		// 如果还有更多字段，继续解析

		var existing Transaction
		db.Table("transaction").Where("transaction_id = ?", trans.TransactionID).First(&existing)
		if existing.TransactionID != "" {
			// 如果已存在，则跳过此条记录
			continue
		}

		transactions = append(transactions, trans)
	}

	sort.Sort(ByTransactionTime(transactions))

	// 保存到数据库
	for _, trans := range transactions {
		result := db.Create(&trans)
		if result.Error != nil {
			log.Printf("Failed to save transaction: %v\n", result.Error)
			// 可以选择继续或中断处理，这里选择继续
			// 如果需要中断，可以使用 return err
		}
	}
	return nil
}

func parseDate(dateStr string) (string, error) {
	const layout = "01/02/2006" // Go 的日期格式是固定的，这里是月/日/年
	parsed, err := time.Parse(layout, dateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse date: %w", err)
	}
	return parsed.Format("2006-01-02"), nil
}

func ImportTransactionRecordFromCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // 允许字段数量不一致

	accountNumber := ""
	// 跳过不必要的行
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// 读取第5行第一列以获取accountNumber
		line := strings.TrimSpace(record[0])
		if strings.HasPrefix(line, "Account: ") {
			accountNumber = extractAccountNumber(line)
		}

		if len(record) > 5 {
			return fmt.Errorf("只能识别Date、Transaction ID、Payment Method、Amount、Currency 五列 %s", record)
		}

		if len(record) == 5 && record[0] == "Date" && record[1] == "Transaction ID" {
			break
		}
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 检查是否为总计行
		if record[2] == "Total Amount Billed" || record[2] == "Total Amount billed" {
			break
		}

		if len(record) < 5 {
			continue // 跳过字段不足的记录
		}

		// 解析行数据到 Transaction 结构
		var trans TransactionRecord
		trans.Date, err = parseDate(record[0]) // 假设parseDate能够正确处理日期格式
		if err != nil {
			// err := fmt.Errorf("%s %s %s", record[0], record[1], record[2])
			return err
		}
		x, err := strconv.ParseFloat(record[3], 64)
		trans.Account = accountNumber
		trans.TransactionID = record[1]
		trans.PaymentMethod = record[2][len(record[2])-4:]
		trans.Amount = x
		trans.Currency = record[4]
		if err != nil {
			log.Printf("Failed to parse amount: %v\n", err)
			return err
		}

		var target Transaction
		db.Table("transaction").Where("transaction_type = ?", "交易清算").
			Where("nickname = ?", accountNumber).
			Where("card_number = ?", record[2][len(record[2])-4:]).
			Where("is_judge = ?", "false").
			Where("order_amount = ?", -x).
			First(&target)

		if target.TransactionID != "" {
			trans.IsTradingAuthorization = true
			db.Model(&target).Update("is_judge", true)
		}

		var existing TransactionRecord
		db.Where("transaction_id = ? AND account = ?", trans.TransactionID, trans.Account).First(&existing)
		if existing.TransactionID != "" {
			// 如果存在，则更新记录
			db.Model(&existing).Updates(map[string]interface{}{
				"is_trading_authorization": trans.IsTradingAuthorization,
				// 根据需要更新其他字段
			})
		} else {
			result := db.Create(&trans)
			if result.Error != nil {
				log.Printf("Failed to save transaction: %v\n", result.Error)
				// 可以选择继续或中断处理
				return result.Error
			}
		}
		// 保存到数据库

	}

	return nil
}

// extractAccountNumber 从给定的字符串中提取账户数字部分
// 假设格式为 "Account: 123456789"，这里仅作为示例
func extractAccountNumber(s string) string {
	parts := strings.Split(s, " ")
	if len(parts) > 1 && strings.HasPrefix(parts[0], "Account:") {
		return parts[1]
	}
	return "" // 如果格式不正确，返回空字符串
}

func GetTransactionRecords(pageSize int, pageNum int, Account string, PaymentMethod string, startTime int, endTime int, set int) ([]TransactionRecord, error, int64) {
	

	var transactionRecords []TransactionRecord  
    query := db.Model(&TransactionRecord{})  
    countQuery := db.Model(&TransactionRecord{}) // 初始化countQuery
	
    // 构造查询条件  
    // query := db.Select("*").  
    //     Limit(pageSize).  
    //     Offset((pageNum - 1) * pageSize)
	if set == 0 {  
		query = query.Order("date ASC")  
		countQuery = countQuery.Order("date ASC")  
	} else if set == 1 {  
		query = query.Order("date DESC")  
		countQuery = countQuery.Order("date DESC")  
	}  
    // 如果 Account 不是空字符串，则添加 Account 过滤条件  
    if Account != "" {  
        query = query.Where("account = ?", Account)  
        countQuery = countQuery.Where("account = ?", Account)  
    }  
  
    // 如果 PaymentMethod 不是空字符串，则添加 PaymentMethod 过滤条件  
    if PaymentMethod != "" {  
        query = query.Where("payment_method = ?", PaymentMethod)  
        countQuery = countQuery.Where("payment_method = ?", PaymentMethod)  
    }  

	if startTime != 0 && endTime != 0 { 
		startTimeT := time.Unix(int64(startTime), 0).UTC()  
        endTimeT := time.Unix(int64(endTime), 0).UTC()
		startDate := startTimeT.Format("2006-01-02")  
		endDate := endTimeT.Format("2006-01-02")  
		query = query.Where("Date BETWEEN ? AND ?", startDate, endDate)  
		countQuery = countQuery.Where("Date BETWEEN ? AND ?", startDate, endDate)  
	}  
  
    // 执行查询并获取交易记录  
    // result := query.Find(&transactionRecords).Count(&total)   
    result := query.  
        Select("*").  
        Limit(pageSize).  
        Offset((pageNum - 1) * pageSize).  
        Order("Date ASC").  
        Find(&transactionRecords)  
  
    var total int64  
    countQuery.Count(&total) 
    // 单独执行计数查询以获取总记录数  
    // db.Model(&TransactionRecord{}).Count(&total)  
  
    // 检查查询过程中是否发生错误  
    if result.Error != nil {  
        // 如果查询过程中发生错误，返回错误  
        return nil, errors.New("failed to retrieve transactions: " + result.Error.Error()), 0  
    }  
  
    // 如果没有错误，返回查询到的交易记录列表、nil（表示没有错误）和总记录数  
    return transactionRecords, nil, total  
}

func GetTransactions(pageSize int, pageNum int ,cardNumber string, transactionType string, startTime int, endTime int, is_judge int) ([]Transaction, error, int) {
	 
	var transactions []Transaction  
  
    // 构建查询条件  
    query := db.Model(&Transaction{})  
    countQuery := db.Model(&Transaction{}) // 初始化countQuery  
  
    if cardNumber != "" {  
        query = query.Where("card_number LIKE ?", "%"+cardNumber+"%")  
        countQuery = countQuery.Where("card_number LIKE ?", "%"+cardNumber+"%")  
    }  
  
    if transactionType != "" {  
        query = query.Where("transaction_type = ?", transactionType)  
        countQuery = countQuery.Where("transaction_type = ?", transactionType)  
    }  
  
    if startTime != 0 && endTime != 0 {  
        startTimeT := time.Unix(int64(startTime), 0).UTC()  
        endTimeT := time.Unix(int64(endTime), 0).UTC()  
        query = query.Where("transaction_time BETWEEN ? AND ?", startTimeT, endTimeT)  
        countQuery = countQuery.Where("transaction_time BETWEEN ? AND ?", startTimeT, endTimeT)  
    }  
	if is_judge != -1 {
		if is_judge == 0 {  
			query = query.Where("is_judge = ? AND transaction_type = ?", 0, "交易清算")  
			countQuery = countQuery.Where("is_judge = ? AND transaction_type = ?", 0, "交易清算")  
		} else if is_judge == 1 {  
			query = query.Where("is_judge = ? AND transaction_type = ?", 1, "交易清算")  
			countQuery = countQuery.Where("is_judge = ? AND transaction_type = ?", 1, "交易清算")  
		}
	}
      
    
    // 应用分页和排序  
    result := query.  
        Select("*").  
        Limit(pageSize).  
        Offset((pageNum - 1) * pageSize).  
        Order("transaction_time ASC").  
        Find(&transactions)  
  
    var total int64  
    countQuery.Count(&total)  
  
    if result.Error != nil {  
        return nil, errors.New("failed to retrieve transactions: " + result.Error.Error()), 0  
    }  
  
    return transactions, nil, int(total) 
}

func CalVccBalance(cardnumber string, startTime int, endTime int) (float64, error) {
	
    // 初始化变量  
    var initialAmount, increaseAmount, decreaseAmount float64  
  
    // 查找与特定卡号相关的开卡交易以获取初始金额  
    var initTrans Transaction  
    if err := db.Where("card_number = ? AND transaction_type = ?", cardnumber, "开卡").First(&initTrans).Error; err != nil {  
        if errors.Is(err, gorm.ErrRecordNotFound) {  
            return 0, fmt.Errorf("没有找到与卡号 %s 相关的开卡交易", cardnumber)  
        }  
        return 0, err  
    }  
    initialAmount = initTrans.OrderAmount  
  
    // 准备时间条件  
    var timeCondition string  
    var timeValues []interface{}  
    if startTime != 0 && endTime != 0 {  
        // 如果startTime和endTime都不为0，则添加时间范围条件  
        timeCondition = "AND transaction_time BETWEEN ? AND ?"  
        timeValues = append(timeValues, time.Unix(int64(startTime), 0).UTC(), time.Unix(int64(endTime), 0).UTC())  
    }
  
    // 计算增加余额的交易总和  
    var sumIncrease float64  
    if err := db.Table("transaction").  
        Select("SUM(order_amount) as total").  
        Where("card_number = ? AND transaction_type IN ? "+timeCondition, append([]interface{}{cardnumber, []string{"卡充值", "交易退款"}}, timeValues...)...).  
        Scan(&sumIncrease).Error; err != nil {  
        return 0, err  
    }  
    increaseAmount = sumIncrease  
  
    // 计算减少余额的交易总和  
    var sumDecrease float64  
    if err := db.Table("transaction").  
        Select("SUM(order_amount) as total").  
        Where("card_number = ? AND transaction_type IN ? "+timeCondition, append([]interface{}{cardnumber, []string{"交易授权", "卡充退"}}, timeValues...)...).  
        Scan(&sumDecrease).Error; err != nil {  
        return 0, err  
    }  
    decreaseAmount = sumDecrease  
  
    // 计算最终余额  
    balance := initialAmount + increaseAmount + decreaseAmount  // 注意这里应该是增加-减少  
  
    return balance, nil  

}

func CalVccTotalDeplete(cardnumber string, startTime int, endTime int) (float64, error) {

	// var sumDecrease float64
	// if err := db.Table("transaction").
	// 	Select("SUM(order_amount) as total").
	// 	Where("card_number = ? AND transaction_type IN ?", cardnumber, []string{"交易授权"}).Scan(&sumDecrease).Error; err != nil {
	// 	return 0, err // 如果查询或扫描失败，返回错误
	// }
	// decreaseAmount := sumDecrease

	// // 计算最终余额
	// balance := decreaseAmount

	// return balance, nil
	var sumDecrease float64  
    // 构建查询条件  
    var conditions []interface{}  
    query := db.Table("transaction").  
        Select("SUM(order_amount) as total")  
  
    // 添加卡号条件  
    conditions = append(conditions, cardnumber)  
    query = query.Where("card_number = ?", cardnumber)  
  
    // 如果 startTime 和 endTime 都非零，则添加时间范围条件  
    if startTime != 0 && endTime != 0 {  
        // 假设 transaction_time 是 UNIX 时间戳（int 类型）  
		startTime1 := time.Unix(int64(startTime), 0).UTC()
		endTime1 := time.Unix(int64(endTime), 0).UTC()
        conditions = append(conditions, startTime1, endTime1)  
        query = query.Where("transaction_time BETWEEN ? AND ?", startTime1, endTime1)  
    }  
  
    // 添加交易类型条件  
    conditions = append(conditions, []string{"交易授权"}) // 注意使用 ... 来展开切片  
    query = query.Where("transaction_type IN ?", conditions[len(conditions)-1:]) // 注意这里的切片操作  
  
    // 执行查询并扫描结果  
    if err := query.Scan(&sumDecrease).Error; err != nil {  
        return 0, err // 如果查询或扫描失败，返回错误  
    }  
  
    // 计算最终余额（这里假设 decreaseAmount 就是你要的余额减少量）  
    // 注意：通常我们需要一个初始的余额值来减去 decreaseAmount，但这里没有提供，所以只返回减少量  
    balance := sumDecrease // 如果 balance 是减少量，这里直接赋值  
  
    return balance, nil
}

type PaginationResult struct {
	CurrentPage map[string]struct {
		Balance string
		Deplete string
	} // 当前页的结果
}

// 实现分页的 ShowVccBalanceAndDeplete 函数
func ShowVccBalanceAndDepletes(IDs []string, pageSize int, pageNum int, startTime int, endTime int) (*PaginationResult, error, int) {

	if pageSize <= 0 || pageNum <= 0 {
		return nil, errors.New("pageSize and pageNum must be positive integers"), 0
	}

	// 计算总项数
	total := len(IDs)
	fmt.Print(total)
	// 计算当前页应该包含的ID索引范围
	startIndex := (pageNum - 1) * pageSize
	if startIndex >= total {
		// 如果没有足够的ID来填充当前页，则返回一个空的当前页结果
		return &PaginationResult{CurrentPage: map[string]struct {
			Balance string
			Deplete string
		}{}}, nil, 0
	}

	endIndex := startIndex + pageSize
	if endIndex > total {
		endIndex = total
	}

	// 只处理当前页范围内的ID
	currentPageIDs := IDs[startIndex:endIndex]

	result := make(map[string]struct {
		Balance string
		Deplete string
	})
	for _, id := range currentPageIDs {
		balance, _ := CalVccBalance(id, startTime, endTime)
		deplete, _ := CalVccTotalDeplete(id ,startTime, endTime)
		result[id] = struct {
			Balance string
			Deplete string
		}{
		  Balance: fmt.Sprintf("%.2f", balance), 
		  Deplete: fmt.Sprintf("%.2f", deplete)}
	}

	return &PaginationResult{CurrentPage: result}, nil, total
}

func ShowVccID() ([]string, error) {
	var cardNumbers []string
	// 使用GORM的Group和Having子句（如果GORM版本支持）或者子查询来找到不同的CardNumber
	// 但由于GORM直接支持可能有限，这里使用原生SQL查询作为示例
	err := db.Raw("SELECT DISTINCT card_number FROM transaction").Scan(&cardNumbers).Error
	if err != nil {
		return nil, errors.New("failed to query unique card numbers: " + err.Error())
	}
	return cardNumbers, nil
}

func ShowFBID(Account string) ([]string, []string, error) {  
    var accounts []string  
    var paymentMethods []string  
  
    // 如果 Account 为空，则仅查询所有不同的 account  
    if Account == "" {  
        err := db.Raw("SELECT DISTINCT account FROM transaction_record").Scan(&accounts).Error  
        if err != nil {  
            return nil, nil, errors.New("failed to query unique accounts: " + err.Error())  
        }  
        return accounts, []string{}, nil // 第二个数组为空，因为没有针对 Account 的 PaymentMethod  
    }  
  
    // 如果 Account 非空，则查询该 Account 下的所有不同 PaymentMethod  
    // 注意：这里我们实际上并不查询 account，但为了保持函数签名，我们仍然声明 accounts 变量  
    // 但我们不会填充它，因为我们只关心 PaymentMethod  
    err := db.Table("transaction_record").  
        Select("DISTINCT payment_method").  
        Where("account = ?", Account).  
        Scan(&paymentMethods).Error  
  
    if err != nil {  
        return nil, nil, errors.New("failed to query unique payment methods for account: " + err.Error())  
    }  
  
    // 由于我们实际上没有查询 account，这里可以返回一个空数组或特定值  
    // 但为了保持函数签名的一致性，我们仍然返回 accounts（空）  
    return []string{}, paymentMethods, nil  
}

func CalVccDepleteByDate(year, month int, cardNumber string) (float64, error) {
	// 将cardNumber模糊处理，仅保留前几位和后几位，以保护隐私
	maskedCardNumber := cardNumber

	// 计算开始和结束时间
	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1).Add(24 * time.Hour) // 加上24小时以包括当天的23:59:59
	fmt.Print(start.Format("2006-01-02 00:00:00"), end.Format("2006-01-02 00:00:00"))
	// 使用GORM查询
	var totalAmount float64
	err := db.Table("transaction").
		Where("transaction_type = ? AND card_number LIKE ? AND DATE(transaction_time) BETWEEN ? AND ?", "交易授权", maskedCardNumber, start.Format("2006-01-02 00:00:00"), end.Format("2006-01-02 00:00:00")).
		Select("SUM(transaction_amount) as total_amount").
		Scan(&totalAmount).
		Error

	if err != nil {
		return 0, err
	}

	return totalAmount, nil
}

func CalFBbyaccount(account string, card_id string ,startTime int,endTime int) (float64, error) {
	var totalAmount float64
	if startTime != 0 && endTime != 0 { 
		startTimeT := time.Unix(int64(startTime), 0).UTC()  
        endTimeT := time.Unix(int64(endTime), 0).UTC()
		startDate := startTimeT.Format("2006-01-02")  
		endDate := endTimeT.Format("2006-01-02")  
		err := db.Table("transaction_record").
		Where("account = ? AND payment_method LIKE ?", account, card_id).
		Where("Date BETWEEN ? AND ?", startDate, endDate).
		Select("SUM(amount) as total_amount").
		Scan(&totalAmount).
		Error 
		if err != nil {
			return 0, err
		}
	
		return totalAmount, nil
	} else {
		err := db.Table("transaction_record").
		Where("account = ? AND payment_method LIKE ?", account, card_id).
		Select("SUM(amount) as total_amount").
		Scan(&totalAmount).
		Error
		if err != nil {
			return 0, err
		}
	
		return totalAmount, nil
	}
	
}

func CalFBbyaccountList(account string, card_id string) ([]TransactionRecord, error) {
	var results []TransactionRecord
	err := db.Table("transaction_record").
		Where("account = ? AND payment_method LIKE ?", account, "%"+card_id+"%").
		Order("date ASC").
		Find(&results).
		Error

	if err != nil {
		return nil, err
	}

	return results, err
}

type TransactionSummary map[string][]struct {
	Date     string
	Amount   float64
	IsTicked bool
	Note     string
}

func Showfb_vccdata(account string) (TransactionSummary, error) {
	var records []TransactionRecord
	err := db.Table("transaction_record").
		Where("account = ?", account).
		Order("payment_method ASC, date ASC").
		Find(&records).
		Error

	if err != nil {
		return nil, err
	}

	summary := make(TransactionSummary)
	for _, record := range records {
		if _, exists := summary[record.PaymentMethod]; !exists {
			summary[record.PaymentMethod] = []struct {
				Date     string
				Amount   float64
				IsTicked bool
				Note     string
			}{}
		}
		summary[record.PaymentMethod] = append(summary[record.PaymentMethod], struct {
			Date     string
			Amount   float64
			IsTicked bool
			Note     string
		}{
			Date:     record.Date,
			Amount:   record.Amount,
			IsTicked: record.IsTicked,
			Note:     record.Note,
		})
	}

	return summary, nil
}

func UpdateTransactionRecord(transaction_id string, isTicked bool, note string) error {
	err := db.Table("transaction_record").
		Where("transaction_id = ?", transaction_id).
		Order("payment_method ASC, date ASC").
		Updates(map[string]interface{}{
			"is_ticked": isTicked,
			"note":      note,
		}).
		Error

	if err != nil {
		return err
	}
	return nil
}
