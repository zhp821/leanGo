package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"stock/data"
	"stock/pojo"
	"time"
)

var stocks []pojo.Stock
var running bool
var threadNum int
var wtime int

func avg(w http.ResponseWriter, req *http.Request) {

	//w.Header().Set("Content-Type", "application/json")

	var views []pojo.View
	if len(stocks) < 100 {
		return
	}
	for _, stock := range stocks {

		avg := (stock.Avg5 - stock.Avg20) / stock.Price
		vavg := (stock.Vavg5 - stock.Vavg20) / stock.Vavg5
		le := len(stock.Days)

		if le > 21 && avg > -0.01 && vavg > 0.2 {

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

		if le > 21 && stock.Open > 0 && change > 6 {

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

func search(w http.ResponseWriter, req *http.Request) {

}
func update(w http.ResponseWriter, req *http.Request) {
	if running {
		return
	}
	running = true
	tmp := data.GetAllDayStockInfoFromQQ(threadNum)
	stocks = tmp
	running = false
}
func updateData() {
	for {
		if !running {
			update(nil, nil)
		}
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
	wtime = *flag.Int("t", 10, "数据更新时间,单位为分钟,默认10分钟")
	threadNum = *flag.Int("n", 30, "抓取时间线程数,默认30")
	flag.Parse()
	fmt.Printf("listen port is %s\n", *port)
	go updateData()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/avg", avg)
	mux.HandleFunc("/api/search", search)
	mux.HandleFunc("/api/sys/update", update)
	mux.HandleFunc("/api/up", up)
	mux.HandleFunc("/api/low", low)
	http.ListenAndServe(":"+*port, mux)

}
