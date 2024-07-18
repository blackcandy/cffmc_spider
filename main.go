package main

import (
	"fmt"
)

func main() {
	// 输入日期
	fmt.Println("请输入日期（YYYY-MM-DD）:")
	var date string
	fmt.Scanln(&date)
	// 下载结算单
	res, err := DownloadSettlementDocument(account, date, byType, path)
	if err != nil {
		fmt.Println("下载结算单失败:", err)
	} else if res {
		fmt.Println("下载结算单成功")
	} else {
		fmt.Println("下载结算单失败")
	}
	select {}
}
