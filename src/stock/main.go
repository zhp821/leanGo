package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"runtime/debug"
	"stock/data"
	"stock/pojo"
	"stock/util"
	"strconv"
	"time"
)

var stocks []pojo.Stock
var running bool
var threadNum int
var wtime int

func vup(w http.ResponseWriter, req *http.Request) {

	var views []pojo.View

	if len(stocks) < 100 {
		return
	}

	for _, stock := range stocks {
		le := len(stock.Days)
		if le < 20 {
			continue
		}
		v1 := (float32(stock.Days[le-1].Volume) - stock.Vavg20) / stock.Vavg20
		v2 := (float32(stock.Days[le-2].Volume) - stock.Vavg20) / stock.Vavg20
		v3 := (float32(stock.Days[le-3].Volume) - stock.Vavg20) / stock.Vavg20
		avg := (stock.Price - stock.Avg5) / stock.Price

		if avg > -0.01 && v1 > 0.2 && v2 > 0.2 && v3 < 0.1 {
			m := make(map[string]pojo.Stock)

			for numt, day := range stock.Days {
				day.Num = numt
				m[day.Time] = day
			}
			view := pojo.View{stock.Code, stock.Name, stock.Time, stock.Price, stock.Open, stock.Avg5, stock.Avg20, "", ""}
			view.Url = "http://finance.sina.com.cn/realstock/company/" + stock.Code + "/nc.shtml"
			views = append(views, view)
		}

	}
	t, err := template.New("webpage").Parse(pojo.Tmpl)

	check(err)
	//	data, _ := json.Marshal(views)
	data := make(map[string]interface{})
	data["stocks"] = views
	data["total"] = len(views)
	data["intro"] = "连续2天成交量放大20%"
	t.Execute(w, data)
}

func kdj(w http.ResponseWriter, req *http.Request) {
	var views []pojo.View
	if len(stocks) < 100 {
		return
	}
	for _, stock := range stocks {
		m := make(map[string]pojo.Stock)
		m3 := make(map[string]pojo.Stock)
		num1 := 0
		ok1 := 0
		le := len(stock.Days)
		if le < 21 || stock.Price == 0 || stock.Volume < 10 {
			continue
		}

		if len(stock.K30) < 21 || len(stock.K60) < 21 {
			continue
		}
		kv60 := stock.K60[len(stock.K60)-1]
		kk60 := kv60 - stock.K60[len(stock.K60)-2]
		kd60 := kv60 - stock.D60[len(stock.D60)-1]

		kv30 := stock.K30[len(stock.K30)-1]
		kk30 := kv30 - stock.K30[len(stock.K30)-2]
		kd30 := kv60 - stock.D30[len(stock.D30)-1]

		if kk60 > 0 && kd60 < 10 && kd60 > -5 && kk30 > 0 && kd30 > 0 {
			view := pojo.View{stock.Code, stock.Name, stock.Time, stock.Price, stock.Open, stock.Avg5, stock.Avg20, "", ""}
			view.Url = "http://finance.sina.com.cn/realstock/company/" + stock.Code + "/nc.shtml"

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

				kd := stock.K60[i] - stock.D60[i]
				t1 := (stock.K60[i] - stock.K60[i-1]) / stock.K60[i-1]
				if kd > -5 && t1 > 0 && kd < 15 {
					nowDay := util.Substr(min.Time, 0, 8)

					min30, o1 := m3[min.Time]
					if !o1 || min30.Num < 1 || stock.K30[min30.Num] < stock.D30[min30.Num] || stock.K30[min30.Num] < stock.K30[min30.Num-1] {
						continue
					}

					day, o := m[nowDay]
					status := "faield"
					numt := day.Num

					if o {
						if numt < 2 {
							continue
						}

						if numt+1 < le {

							num1++

							rate := (stock.Days[numt+1].High - min.Close) / min.Close

							if rate > 0.01 {

								ok1++
								status = "success"

							}
							view.Des = view.Des + "--" + min.Time + "--" + status + ";"
						}
					}
				}
			}
			view.Des = strconv.Itoa(ok1) + "/" + strconv.Itoa(num1) + ";" + view.Des

			if num1 > 0 && (ok1*100/num1 > 74) {
				views = append(views, view)
			}
		}

	}
	t, err := template.New("webpage").Parse(pojo.Tmpl)

	check(err)
	//	data, _ := json.Marshal(views)
	data := make(map[string]interface{})
	data["stocks"] = views
	data["total"] = len(views)
	data["intro"] = "kdj60转向上升并即将或已金叉,kdj30已金叉"
	t.Execute(w, data)
}

func avg(w http.ResponseWriter, req *http.Request) {

	//w.Header().Set("Content-Type", "application/json")

	var views []pojo.View
	if len(stocks) < 100 {
		return
	}
	for _, stock := range stocks {

		avg := (stock.Avg5 - stock.Avg20) / stock.Price
		vavg := (stock.Vavg5 - stock.Vavg20) / stock.Vavg5
		//vavg := (float32(stock.Volume) - stock.Vavg20) / stock.Vavg20
		le := len(stock.Days)

		if le > 21 && avg > -0.1 && vavg > 0.2 {

			kv60 := stock.K60[len(stock.K60)-1]
			kk60 := kv60 - stock.K60[len(stock.K60)-2]

			tmp := (stock.Price - stock.Days[le-3].Close) / stock.Days[le-3].Close

			if kk60 > 1 && tmp > 0.01 {

				m := make(map[string]pojo.Stock)

				for numt, day := range stock.Days {
					day.Num = numt
					m[day.Time] = day
				}
				view := pojo.View{stock.Code, stock.Name, stock.Time, stock.Price, stock.Open, stock.Avg5, stock.Avg20, "", ""}
				view.Url = "http://finance.sina.com.cn/realstock/company/" + stock.Code + "/nc.shtml"
				views = append(views, view)
			}
		}
	}

	t, err := template.New("webpage").Parse(pojo.Tmpl)

	check(err)
	//	data, _ := json.Marshal(views)
	data := make(map[string]interface{})
	data["stocks"] = views
	data["total"] = len(views)
	t.Execute(w, data)

}
func low(w http.ResponseWriter, req *http.Request) {

	//w.Header().Set("Content-Type", "application/html")
	var views []pojo.View
	if len(stocks) < 100 {
		return
	}
	for _, stock := range stocks {

		le := len(stock.Days)

		if le < 22 {
			continue
		}

		//change := (stock.Price - stock.Open) * 100 / stock.Open
		changeLow := (stock.Days[le-2].Close - stock.Days[le-3].Close) * 100 / stock.Days[le-3].Close

		if stock.Pe < 50 && stock.Open > 0 && changeLow < -1.5 {

			c5 := (stock.Days[le-5].Close - stock.Days[le-5].Open) * 100 / stock.Days[le-5].Open
			c4 := (stock.Days[le-4].Close - stock.Days[le-4].Open) * 100 / stock.Days[le-4].Open
			c3 := (stock.Days[le-3].Close - stock.Days[le-3].Open) * 100 / stock.Days[le-3].Open

			if c5 < 4 && c4 < 4 && c3 < 4 {
				continue
			}
			kv60 := stock.K60[len(stock.K60)-1]
			kk60 := kv60 - stock.K60[len(stock.K60)-2]
			kd60 := kv60 - stock.D60[len(stock.D60)-1]

			if kk60 > 1 && kd60 > 0 {
				view := pojo.View{stock.Code, stock.Name, stock.Time, stock.Price, stock.Open, stock.Avg5, stock.Avg20, "", ""}
				view.Url = "http://finance.sina.com.cn/realstock/company/" + stock.Code + "/nc.shtml"
				views = append(views, view)
			}
		}
	}
	t, err := template.New("webpage").Parse(pojo.Tmpl)

	check(err)
	//	data, _ := json.Marshal(views)
	data := make(map[string]interface{})
	data["stocks"] = views
	data["total"] = len(views)
	t.Execute(w, data)
}

func up(w http.ResponseWriter, req *http.Request) {

	//w.Header().Set("Content-Type", "application/json")
	var views []pojo.View
	if len(stocks) < 100 {
		return
	}

	for _, stock := range stocks {

		le := len(stock.Days)
		change := (stock.Price - stock.Open) * 100 / stock.Open

		if le > 21 && stock.Open > 0 && change > 3.5 {

			kv60 := stock.K60[len(stock.K60)-1]
			kk60 := kv60 - stock.K60[len(stock.K60)-2]

			kd60 := kv60 - stock.D60[len(stock.D60)-1]

			if kk60 > 1 && kd60 > -1 {
				view := pojo.View{stock.Code, stock.Name, stock.Time, stock.Price, stock.Open, stock.Avg5, stock.Avg20, "", ""}
				view.Url = "http://finance.sina.com.cn/realstock/company/" + stock.Code + "/nc.shtml"
				views = append(views, view)
			}
		}
	}

	t, err := template.New("webpage").Parse(pojo.Tmpl)

	check(err)
	//	data, _ := json.Marshal(views)
	data := make(map[string]interface{})
	data["stocks"] = views
	data["total"] = len(views)
	t.Execute(w, data)
}

func j0(w http.ResponseWriter, req *http.Request) {
	var views []pojo.View
	if len(stocks) < 100 {
		return
	}

	for _, stock := range stocks {

		if len(stock.J15) < 5 {
			continue
		}

		jv15 := stock.J15[len(stock.J15)-1]
		jj15 := stock.J15[len(stock.J15)-2]
		jj3 := stock.J15[len(stock.J15)-3]

		if jv15 < 21 || jj15 < 21 || jj3 < 21 {
			view := pojo.View{stock.Code, stock.Name, stock.Time, stock.Price, stock.Open, stock.Avg5, stock.Avg20, "", ""}
			view.Url = "http://finance.sina.com.cn/realstock/company/" + stock.Code + "/nc.shtml"
			views = append(views, view)
		}

	}
	t, err := template.New("webpage").Parse(pojo.Tmpl)

	check(err)
	//	data, _ := json.Marshal(views)
	data := make(map[string]interface{})
	data["stocks"] = views
	data["total"] = len(views)
	t.Execute(w, data)
}

func search(w http.ResponseWriter, req *http.Request) {

}
func update(w http.ResponseWriter, req *http.Request) {
	if running {
		return
	}
	running = true
	tmp := data.GetAllDayStockInfoFromQQ(threadNum)
	stocks = tmp
	tmp = nil
	running = false
}
func updateData() {

	for {

		tmp := time.Now()
		now := tmp.Hour()
		if len(stocks) > 100 {
			if now < 9 || now > 15 {
				fmt.Printf("~~~wtime=%d ;running=%t,now=%d\n", wtime, running, now)
				//time.Sleep(time.Duration(wtime*10) * time.Second)
				//continue
			}
		}
		if !running {
			update(nil, nil)
		}
		debug.FreeOSMemory()
		time.Sleep(time.Duration(wtime*60) * time.Second)
	}
}
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	port := flag.String("p", "9999", "启动端口号")
	targ := flag.Int("t", 10, "数据更新时间,单位为分钟,默认10分钟")
	threadNumArg := flag.Int("n", 30, "抓取时间线程数,默认30")
	flag.Parse()
	fmt.Printf("port=%s,wtime=%d,threadNum=%d\n", *port, *targ, *threadNumArg)
	wtime = *targ
	threadNum = *threadNumArg
	go updateData()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/avg", avg)
	mux.HandleFunc("/api/search", search)
	mux.HandleFunc("/api/update", update)
	mux.HandleFunc("/api/up", up)
	mux.HandleFunc("/api/low", low)
	mux.HandleFunc("/api/kdj", kdj)
	mux.HandleFunc("/api/j0", j0)
	mux.HandleFunc("/api/vup", vup)
	http.ListenAndServe(":"+*port, mux)
}
