package avg

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

	client := &http.Client{}

	req, err := http.NewRequest("POST", "http://screener.finance.sina.com.cn/znxg/data/json.php/SSCore.doView?num=4000&sort=&asc=0&field0=stocktype&field1=sinahy&field2=diyu&value0=*&value1=*&value2=*&field3=dtsyl&max3=100&min3=1&field4=trade&max4=752.13&min4=0", strings.NewReader(""))
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

	strr = strings.Replace(strr, "(", "", -1)
	strr = strings.Replace(strr, ")", "", -1)
	strr = strings.Replace(strr, "items", "\"items\"", -1)
	strr = strings.Replace(strr, "symbol", "\"symbol\"", -1)
	strr = strings.Replace(strr, "name", "\"name\"", -1)
	strr = strings.Replace(strr, "dtsyl", "\"dtsyl\"", -1)
	strr = strings.Replace(strr, "trade", "\"trade\"", -1)
	strr = strings.Replace(strr, ",total", ",\"total\"", -1)
	strr = strings.Replace(strr, "page_total", "\"page_total\"", -1)
	strr = strings.Replace(strr, ",page", ",\"page\"", -1)
	strr = strings.Replace(strr, "num_per_page", "\"num_per_page\"", -1)

	var data pojo.SinaStock
	errr := json.Unmarshal([]byte(strr), &data)
	if errr == nil {

		return data.Items
	} else {
		fmt.Println(errr)
		return nil
	}

}

//获取分钟级别的k线数据 m5,m60,m30
func getMinStocks(m string, stock pojo.Stock) {
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
		tmp.Low = util.ParseFloat(vv[3].(string), 32)
		mstock = append(mstock, tmp)
	}
	kdj := util.NewKdj(9, 3, 3)
	k, d, j := kdj.Kdj(mstock)
	lev := len(mstock)
	stock.K = k[lev-1]
	stock.D = d[lev-1]
	stock.J = j[lev-1]
}

func getStockInfoFromQQ(stock pojo.Stock, ch chan int, num int, waitgroup sync.WaitGroup) {
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

	for _, info := range s {

		if len(info) < 10 {
			continue
		}
		info = strings.Replace(info, "\n", "", -1)
		oneday := strings.Split(info, " ")

		parr = append(parr, util.ParseFloat(oneday[2], 32))
		varr = append(varr, util.ParseInt(oneday[5]))

		one := stock
		one.Code = stock.Code
		one.Name = stock.Name
		one.Time = "20" + oneday[0]
		one.Open = util.ParseFloat(oneday[1], 32)
		one.Close = util.ParseFloat(oneday[2], 32)
		one.High = util.ParseFloat(oneday[3], 32)
		one.Low = util.ParseFloat(oneday[4], 32)
		one.Volume = util.ParseInt(oneday[5])
		stock.Days = append(stock.Days, one)
	}

	lev := len(varr)
	le := len(parr)

	//此处算均价的时候加上当前的实时价格,但成交量以抓到的数据为准;

	if stock.Price != parr[le-1] {
		parr = append(parr, stock.Price)
		le = len(parr)
	}
	//此处计算日的kdj
	//	kdj := util.NewKdj(9, 3, 3)
	//	k, d, j := kdj.Kdj(stock.Days)

	//	stock.K = k[lev-1]
	//	stock.D = d[lev-1]
	//	stock.J = j[lev-1]

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
	avg := (stock.Avg5 - stock.Avg20) / stock.Price
	vavg := (stock.Vavg5 - stock.Vavg20) / stock.Vavg5

	if le > 21 && avg > -0.01 && avg < 0.05 && vavg > 0.2 && stock.Price > stock.Avg20 && stock.Price > stock.Avg5 {
		getMinStocks("m30", stock)
		kd := stock.K - stock.D
		fmt.Printf("%d--name:%s, code:%s,time:%s,price:%f,change:%f,avg5: %f ,avg20:%f,vavg5:%f,vavg20:%f,k:%f,d:%f,j:%f,kd:%f \n", num, stock.Name, stock.Code, stock.Days[le-1].Time, stock.Price, (stock.Price - parr[le-2]), stock.Avg5, stock.Avg20, stock.Vavg5, stock.Vavg20, stock.K, stock.D, stock.J, kd)
	}

	//fmt.Printf("%d--%s %d  end\n", num, stock.Name, len(stock.Days))
	<-ch
	waitgroup.Done()
}
func GetAllDayStockInfoFromQQ(num int32) {
	start := time.Now()
	stocks := realStockPrice()
	timeStr := time.Now().Format("2006-01-02 15:04:05")
	ch := make(chan int, num)
	defer close(ch)
	println(len(stocks))
	var waitgroup sync.WaitGroup
	for i := 0; i < len(stocks); i++ {
		stocks[i].Time = timeStr
		ch <- 1
		go getStockInfoFromQQ(stocks[i], ch, i, waitgroup)
	}
	waitgroup.Wait()
	end := time.Now()
	delta := end.Sub(start)
	fmt.Printf("~~it fetch all sina stock sucess cost %s~~", delta)
}
