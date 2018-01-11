package avg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"stock/pojo"
	"stock/util"
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

		//		one := stock
		//		one.Code = stock.Code
		//		one.Name = stock.Name
		//		one.Time = "20" + oneday[0]
		//		one.Open = util.ParseFloat(oneday[1], 32)
		//		one.Close = util.ParseFloat(oneday[2], 32)
		//		one.High = util.ParseFloat(oneday[3], 32)
		//		one.Low = util.ParseFloat(oneday[4], 32)
		//		one.Volume = util.ParseInt(oneday[5])
		//		stock.Days = append(stock.Days, one)
	}

	//	le := len(parr)
	//	parr = parr[0 : le-1]
	//	varr = varr[0 : le-1]
	lev := len(varr)
	le := len(parr)

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
	avg := (stock.Avg5 - stock.Avg20) / stock.Price
	vavg := (stock.Vavg5 - stock.Vavg20) / stock.Vavg5

	if le > 21 && avg > -0.01 && avg < 0.05 && vavg > 0.2 && stock.Price > stock.Avg20 && stock.Price > stock.Avg5 {
		fmt.Printf("%d--name:%s, code:\n%s\n,price:%f,change:%f,avg5: %f ,avg20:%f,vavg5:%f,vavg20:%f \n", num, stock.Name, stock.Code, stock.Price, (stock.Price - parr[le-1]), stock.Avg5, stock.Avg20, stock.Vavg5, stock.Vavg20)
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
