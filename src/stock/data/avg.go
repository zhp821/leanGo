package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"stock/pojo"
	"stock/util"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/axgle/mahonia"
)

func realStockPrice() []pojo.Stock {

	var data pojo.SinaStock

	for i := 1; i < 100; i++ {

		client := &http.Client{}

		//req, err := http.NewRequest("POST", "http://screener.finance.sina.com.cn/znxg/data/json.php/SSCore.doView?num=4000&sort=&asc=0&field0=stocktype&field1=sinahy&field2=diyu&value0=*&value1=*&value2=*&field3=trade&max3=752.13&min3=0&field4=open&max4=780.48&min4=0&field5=high&max5=788.61&min5=0&field6=low&max6=768&min6=0&field7=volume&max7=1105818634&min7=0&field8=dtsyl&max8=7379.21&min8=0", strings.NewReader(""))
		req, err := http.NewRequest("POST", "http://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData?page="+strconv.Itoa(i)+"&num=4000&sort=changepercent&asc=0&node=hs_a&symbol=&_s_r_a=page", strings.NewReader(""))
		if err != nil {
			// handle error
		}
		req.Header.Set("Content-Type", "application/json; charset=gbk")
		resp, err := client.Do(req)

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		html := string(body)
		dec := mahonia.NewDecoder("GBK")

		strr := dec.ConvertString(html)
		fmt.Printf(strconv.Itoa(i) + " ")
		if len(strr) < 500 {
			break
		}

		//	strr = strings.Replace(strr, "(", "", -1)
		//	strr = strings.Replace(strr, ")", "", -1)
		//strr = strings.Replace(strr, "items", "\"items\"", -1)
		strr = strings.Replace(strr, "symbol", "\"symbol\"", -1)
		strr = strings.Replace(strr, "name", "\"name\"", -1)
		strr = strings.Replace(strr, "code", "\"code\"", -1)
		strr = strings.Replace(strr, "pricechange", "\"pricechange\"", -1)
		strr = strings.Replace(strr, "changepercent", "\"changepercent\"", -1)
		strr = strings.Replace(strr, "sell", "\"sell\"", -1)
		strr = strings.Replace(strr, "buy", "\"buy\"", -1)
		strr = strings.Replace(strr, "settlement", "\"settlement\"", -1)
		strr = strings.Replace(strr, "amount", "\"amount\"", -1)
		strr = strings.Replace(strr, "ticktime", "\"ticktime\"", -1)
		strr = strings.Replace(strr, "per:", "\"per\":", -1)
		strr = strings.Replace(strr, "mktcap", "\"mktcap\"", -1)
		strr = strings.Replace(strr, "nmc", "\"nmc\"", -1)
		strr = strings.Replace(strr, "turnoverratio", "\"turnoverratio\"", -1)
		strr = strings.Replace(strr, "pb:", "\"pb\":", -1)

		strr = strings.Replace(strr, "dtsyl", "\"dtsyl\"", -1)
		strr = strings.Replace(strr, "trade", "\"trade\"", -1)
		strr = strings.Replace(strr, "open", "\"open\"", -1)
		strr = strings.Replace(strr, "high", "\"high\"", -1)
		strr = strings.Replace(strr, "low", "\"low\"", -1)
		strr = strings.Replace(strr, "volume", "\"volume\"", -1)
		strr = strings.Replace(strr, ",total", ",\"total\"", -1)
		strr = strings.Replace(strr, "page_total", "\"page_total\"", -1)
		strr = strings.Replace(strr, ",page", ",\"page\"", -1)
		strr = strings.Replace(strr, "num_per_page", "\"num_per_page\"", -1)

		//fmt.Println(strr)

		var stocks []pojo.Stock
		errr := json.Unmarshal([]byte(strr), &stocks)

		data.Items = append(data.Items, stocks...)

		if errr != nil {
			return data.Items
		}

	}
	return data.Items

}

//获取分钟级别的k线数据 m60,m30,m15
func getMinStocks(m string, stock *pojo.Stock) (k, d, j, rsv []float32) {
	r := strconv.FormatFloat(rand.Float64(), 'E', -1, 32)
	url := "http://ifzq.gtimg.cn/appstock/app/kline/mkline?param=" + stock.Code + "," + m + ",,50&r=" + r
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		// handle error
	}
	req.Header.Set("Content-Type", "application/json; charset=gbk")
	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if len(string(body)) < 500 {
		return
	}
	str := strings.Replace(string(body), "m60", "min", -1)
	str = strings.Replace(str, "m30", "min", -1)
	str = strings.Replace(str, "m15", "min", -1)
	str = strings.Replace(str, stock.Code, "list", -1)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(str), &data)
	if err != nil {
		// handle error
	}
	data = (data["data"].(map[string]interface{}))
	list := (data["list"].(map[string]interface{}))
	values := (list["min"].([]interface{}))
	var mstock []pojo.Stock
	for _, value := range values {
		vv, _ := value.([]interface{})
		var tmp pojo.Stock
		tmp.Time = vv[0].(string)
		tmp.Open = util.ParseFloat(vv[1].(string), 32)
		tmp.Close = util.ParseFloat(vv[2].(string), 32)
		tmp.High = util.ParseFloat(vv[3].(string), 32)
		tmp.Low = util.ParseFloat(vv[4].(string), 32)
		tmp.Price = tmp.Close
		mstock = append(mstock, tmp)
	}
	kdj := util.NewKdj(9, 3, 3)

	if m == "m60" {
		stock.M60 = mstock
	}
	if m == "m30" {
		stock.M30 = mstock
	}
	if m == "m15" {
		stock.M15 = mstock
	}

	return kdj.Kdj(mstock)
}
func getAllKdj(stock *pojo.Stock) {

	kdj := util.NewKdj(9, 3, 3)
	k, d, j, _ := kdj.Kdj(stock.Days)
	stock.K = k
	stock.D = d
	stock.J = j

	k, d, j, _ = getMinStocks("m60", stock)
	stock.K60 = k
	stock.D60 = d
	stock.J60 = j

	k, d, j, _ = getMinStocks("m30", stock)
	stock.K30 = k
	stock.D30 = d
	stock.J30 = j

	k, d, j, _ = getMinStocks("m15", stock)
	stock.K15 = k
	stock.D15 = d
	stock.J15 = j

}

func getStockInfoFromQQ(stock *pojo.Stock, ch chan int, num int, waitgroup sync.WaitGroup) {
	waitgroup.Add(1)
	url := "http://data.gtimg.cn/flashdata/hushen/latest/daily/" + stock.Code + ".js"
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, strings.NewReader(""))
	if err != nil {
		// handle error
	}
	req.Header.Set("Content-Type", "application/json; charset=gbk")
	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	s := strings.Split(string(body), "\\n\\")[2:]

	var parr []float32
	var varr []int
	var days []pojo.Stock

	for i, info := range s {

		if len(info) < 10 {
			continue
		}
		info = strings.Replace(info, "\n", "", -1)
		oneday := strings.Split(info, " ")

		parr = append(parr, util.ParseFloat(oneday[2], 32))
		varr = append(varr, util.ParseInt(oneday[5]))

		var one pojo.Stock
		one.Code = stock.Code
		one.Name = stock.Name
		one.Time = "20" + oneday[0]
		one.Open = util.ParseFloat(oneday[1], 32)
		one.Close = util.ParseFloat(oneday[2], 32)
		one.High = util.ParseFloat(oneday[3], 32)
		one.Low = util.ParseFloat(oneday[4], 32)
		one.Volume = util.ParseInt(oneday[5])
		le := len(parr)
		if i > 10 {
			one.Avg5 = util.Avg(parr[le-5 : le])
			one.Vavg5 = util.Avgint(varr[le-5 : le])
			one.Avg10 = util.Avg(parr[le-10 : le])

			min, max := util.MinAndMax(days[le-11 : le-1])
			one.Max10 = max
			one.Min10 = min

		}
		if i > 20 {
			one.Avg20 = util.Avg(parr[le-20 : le])
			one.Vavg20 = util.Avgint(varr[le-20 : le])
			min, max := util.MinAndMax(days[le-21 : le-1])
			one.Max20 = max
			one.Min20 = min
		}

		days = append(days, one)
	}
	stock.Days = days
	lev := len(varr)
	le := len(parr)

	if le < 5 {
		//fmt.Printf("%s-%s stock is  empty...\n", stock.Code, stock.Name)
		<-ch
		waitgroup.Done()
		return
	}

	//此处算均价的时候加上当前的实时价格,但成交量以抓到的数据为准;

	if stock.Price != parr[le-1] {
		parr = append(parr, stock.Price)
		le = len(parr)
	}

	if le > 21 {

		stock.Avg5 = util.Avg(parr[le-5 : le])
		stock.Avg20 = util.Avg(parr[le-20 : le])

		stock.Vavg5 = util.Avgint(varr[lev-5 : lev])
		stock.Vavg20 = util.Avgint(varr[lev-20 : lev])
	}
	if le > 60 {
		tmp60 := parr[le-60 : le]
		stock.Avg60 = util.Avg(tmp60)
	}
	getAllKdj(stock)

	//	avg := (stock.Avg5 - stock.Avg20) / stock.Price
	//	vavg := (stock.Vavg5 - stock.Vavg20) / stock.Vavg5
	//	change := (stock.Price - stock.Open) * 100 / stock.Open

	//	kv60 := stock.K60[len(stock.K60)-1]
	//	kd60 := kv60 - stock.K60[len(stock.K60)-2]

	//	if le > 21 && kv60 < 80 && kd60 > 5 && avg > -0.2 && vavg > 0.3 && change > 6 {
	//		fmt.Printf("%d--name:%s, code:%s,price:%f,change:%f,avg5: %f ,avg20:%f,vavg5:%f,vavg20:%f \n", num, stock.Name, stock.Code, stock.Price, change, stock.Avg5, stock.Avg20, stock.Vavg5, stock.Vavg20)
	//	}

	<-ch
	waitgroup.Done()
}
func GetAllDayStockInfoFromQQ(num int) (stocks []pojo.Stock) {
	start := time.Now()
	stocks = realStockPrice()
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	ch := make(chan int, num)
	defer close(ch)
	fmt.Printf("一共将抓取%d条stock\n", len(stocks))
	var waitgroup sync.WaitGroup
	for i := 0; i < len(stocks); i++ {
		stocks[i].Time = timeStr
		ch <- 1
		go getStockInfoFromQQ(&stocks[i], ch, i, waitgroup)
	}
	waitgroup.Wait()
	end := time.Now()
	delta := end.Sub(start)
	fmt.Printf("~~it fetch all sina stock sucess cost %s\n~~", delta)
	return stocks
}
