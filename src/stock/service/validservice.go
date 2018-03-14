package main

import (
	"fmt"
	"stock/data"
	"stock/pojo"
	"stock/util"
	"strconv"
)

var stocks []pojo.Stock

func validupavg() {
	stocks = data.GetAllDayStockInfoFromQQ(30)
	if len(stocks) < 100 {
		return
	}
	num := 0
	ok := 0
	for _, stock := range stocks {
		le := len(stock.Days)
		if le < 21 || stock.Pe > 70 {
			continue
		}
		des := ""
		num1 := 0
		ok1 := 0

		for i, day := range stock.Days {

			avg := (day.Avg5 - day.Avg20) / day.Avg20
			vavg := (day.Vavg5 - day.Vavg20) / day.Vavg5
			status := "faield"
			tmp := stock.K[i] - stock.D[i]

			if avg > -0.05 && avg < 0.2 && vavg > 0.5 && tmp > 0 && stock.Pe < 70 {
				if i+3 > le-1 {
					continue
				}
				num++
				num1++
				day1 := stock.Days[i+1]
				day2 := stock.Days[i+2]
				day3 := stock.Days[i+3]

				if day1.Close > day.Close || day2.Close > day.Close || day3.Close > day.Close {

					if day2.Close > day.Close || day3.Close > day.Close || day1.Close > day.Close {
						ok++
						ok1++
						status = "success"

					}
				}
				des = des + ";" + day2.Time + "--" + status + ";"
			}

		}
		if num1 > 0 {
			fmt.Println(stock.Name+"--"+stock.Code+"--"+des+",一共"+strconv.Itoa(num1), "成功"+strconv.Itoa(ok1))
		}
	}
	if num > 0 {
		tmp2 := 100 * ok / num
		fmt.Println("一共" + strconv.Itoa(num) + ",成功" + strconv.Itoa(ok) + ",成功率" + strconv.Itoa(tmp2) + "%")
	}

}

func validlow2() {
	stocks = data.GetAllDayStockInfoFromQQ(30)
	if len(stocks) < 100 {
		return
	}
	num := 0
	ok := 0
	for _, stock := range stocks {
		le := len(stock.Days)
		if le < 21 || stock.Pe > 70 {
			continue
		}
		des := ""
		num1 := 0
		ok1 := 0
		for i, day := range stock.Days {

			if i+4 > le-1 {
				continue
			}
			if i-15 < 0 {
				continue
			}

			day1 := stock.Days[i+1]
			day2 := stock.Days[i+2]
			day3 := stock.Days[i+3]
			day4 := stock.Days[i+3]

			daypre5 := stock.Days[i-1].Avg5
			daypre10 := stock.Days[i-5].Avg5
			daypre15 := stock.Days[i-10].Avg5

			if daypre5 > daypre10 && daypre10 > daypre15 && day1.Close < day.Close && day2.Close < day1.Close {
				num++
				num1++
				if day3.Close > day2.Close || day4.Close > day2.Close {
					ok++
					ok1++
					des = des + "--" + day2.Time
				}
			}
		}
		fmt.Println(stock.Name+"--"+stock.Code+"--"+des+",一共"+strconv.Itoa(num1), "成功"+strconv.Itoa(ok1))
	}
	tmp2 := 100 * ok / num
	fmt.Println("一共" + strconv.Itoa(num) + ",成功" + strconv.Itoa(ok) + ",成功率" + strconv.Itoa(tmp2) + "%")
}

func validkdj60() {
	stocks = data.GetAllDayStockInfoFromQQ(30)
	if len(stocks) < 100 {
		return
	}
	num := 0
	ok := 0
	for _, stock := range stocks {
		le := len(stock.Days)
		if le < 21 {
			continue
		}
		des := ""
		num1 := 0
		ok1 := 0

		m := make(map[string]pojo.Stock)
		m3 := make(map[string]pojo.Stock)

		for numt, day := range stock.Days {
			day.Num = numt
			m[day.Time] = day
		}
		for numt1, min1 := range stock.M30 {
			min1.Num = numt1
			m3[min1.Time] = min1
		}

		for i, min := range stock.M60 {

			if i < 1 {
				continue
			}
			nowHour := util.Substr(min.Time, 8, len(min.Time))

			if util.ParseInt(nowHour) < 1300 {
				continue
			}
			kd := stock.K60[i] - stock.D60[i]
			t1 := (stock.K60[i] - stock.K60[i-1]) / stock.K60[i-1]

			if kd > -5 && t1 > 0 {

				nowDay := util.Substr(min.Time, 0, 8)

				//				min30, o1 := m3[min.Time]
				//				if !o1 || min30.Num < 1 || stock.K30[min30.Num] < stock.D30[min30.Num] || stock.K30[min30.Num] < stock.K30[min30.Num-1] {
				//					continue
				//				}

				day, o := m[nowDay]

				if o {

					status := "faield"
					numt := day.Num

					if numt < 2 {
						continue
					}
					//&& day.Avg5 > day.Avg10 && day.Avg10 > day.Avg20 && day.Avg20 > day.Avg60

					if numt+1 < le {

						num++
						num1++

						rate := (stock.Days[numt+1].Close - min.Close) / min.Close

						if rate > 0.01 {
							ok++
							ok1++
							status = "success"
						}
						des = des + "--" + min.Time + "--" + status

					}

				}
			}
		}
		if num1 > 0 {
			tmp3 := 100 * ok1 / num1
			if tmp3 > 80 {
				fmt.Println(stock.Name+"--"+stock.Code+"--"+des+",一共"+strconv.Itoa(num1), "成功"+strconv.Itoa(ok1)+"\n")

			}

		}
	}
	if num > 0 {
		tmp2 := 100 * ok / num
		fmt.Println("一共" + strconv.Itoa(num) + ",成功" + strconv.Itoa(ok) + ",成功率" + strconv.Itoa(tmp2) + "%")
	}

}
func validlow() {
	stocks = data.GetAllDayStockInfoFromQQ(30)
	if len(stocks) < 100 {
		return
	}
	num := 0
	ok := 0
	for _, stock := range stocks {
		le := len(stock.Days)
		if le < 21 {
			continue
		}
		des := ""
		num1 := 0
		ok1 := 0

		for numt, day := range stock.Days {
			if numt > 1 && day.Avg10 > day.Avg20 && day.Avg20 > day.Avg60 {

				if day.Low < day.Min20 {
					num1++
					num++

					if numt+2 < le && stock.Days[numt+1].Close > day.Close {
						ok++
						ok1++
					}
				}
			}
		}
		if num1 > 0 {
			fmt.Println(stock.Name+"--"+stock.Code+des+"--一共"+strconv.Itoa(num1), "成功"+strconv.Itoa(ok1))
		}
	}
	if num > 0 {
		tmp2 := 100 * ok / num
		fmt.Println("一共" + strconv.Itoa(num) + ",成功" + strconv.Itoa(ok) + ",成功率" + strconv.Itoa(tmp2) + "%")
	}
}
func upAfternoon() {
	stocks = data.GetAllDayStockInfoFromQQ(30)
	if len(stocks) < 100 {
		return
	}
	num := 0
	ok := 0
	for _, stock := range stocks {
		le := len(stock.Days)
		if le < 21 {
			continue
		}
		des := ""
		num1 := 0
		ok1 := 0
		m30 := make(map[string]pojo.Stock)

		for numt1, min1 := range stock.M30 {
			min1.Num = numt1
			m30[min1.Time] = min1
		}

		for i, day := range stock.Days {
			rate := (day.Close - day.Open) / day.Open
			time := day.Time + "1330"

			so, o := m30[time]

			if o {
				rate2 := (so.High - day.Open) / day.Open

				if rate2 < 0.01 && rate > 0.03 {
					num++
					num1++
					if i+1 < le {
						if stock.Days[i+1].High > day.Close {
							ok1++
							ok++
							des = des + "-" + stock.Time
						}
					}
				}
			}

		}
		if num1 > 0 {
			fmt.Println(stock.Name+"--"+stock.Code+"--"+des+",一共"+strconv.Itoa(num1), "成功"+strconv.Itoa(ok1))

		}
	}
	if num > 0 {
		tmp2 := 100 * ok / num
		fmt.Println("一共" + strconv.Itoa(num) + ",成功" + strconv.Itoa(ok) + ",成功率" + strconv.Itoa(tmp2) + "%")
	}
}

func validUpKdj15() {
	stocks = data.GetAllDayStockInfoFromQQ(30)
	if len(stocks) < 100 {
		return
	}
	num := 0
	ok := 0
	for _, stock := range stocks {
		le := len(stock.Days)
		if le < 21 {
			continue
		}
		des := ""
		num1 := 0
		ok1 := 0
		m := make(map[string]pojo.Stock)
		for numt1, d1 := range stock.Days {
			d1.Num = numt1
			m[d1.Time] = d1
		}
		for i, min := range stock.M30 {

			if i < 1 {
				continue
			}
			kd := stock.K30[i] - stock.D30[i]
			kk := stock.K30[i] - stock.K30[i-1]
			status := "failed"

			if kd > -10 && kd < 10 && kk > 5 {
				nowDay := util.Substr(min.Time, 0, 8)
				so, o := m[nowDay]
				if o && so.Num > 2 {
					soPre2 := stock.Days[so.Num-2]
					if soPre2.Avg5 < so.Avg5 && so.Avg5 > so.Avg10 {
						num1++
						num++
						fmt.Printf(soPre2.Time + "," + so.Time)
						if min.Close < so.Close {
							ok1++
							ok++
							status = "suc"
						}
						des = des + ";" + min.Time + ":" + status

					}
				}
			}

		}
		if num1 > 0 {
			fmt.Println(stock.Name+"--"+stock.Code+"--"+des+",一共"+strconv.Itoa(num1), "成功"+strconv.Itoa(ok1))
		}

	}
	if num > 0 {
		tmp2 := 100 * ok / num
		fmt.Println("一共" + strconv.Itoa(num) + ",成功" + strconv.Itoa(ok) + ",成功率" + strconv.Itoa(tmp2) + "%")
	}
}

func main() {
	validUpKdj15()
}
