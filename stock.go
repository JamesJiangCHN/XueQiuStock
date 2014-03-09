package main

import (
	"bufio"
	"code.google.com/p/mahonia"
	"compress/gzip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type XQUser struct {
	username      string
	password      string
	telephone     string
	cookies       []*http.Cookie
	Refresh_token string
	Access_token  string
}

type Stock struct {
	Code     string
	Name     string
	EnName   string
	Hasexist string
	deatil   StockDetail
}

type StockDeal struct {
	date      string
	time      string
	code      string
	name      string
	deal      string
	callPrice string
}

type StockDetail struct {
	Symbol     string
	Name       string
	Current    string //现价
	Percentage string //涨幅
	Change     string //涨跌额
	Open       string //今开
	Last_close string //昨收
}
type Quotes struct {
	Quotes []StockDetail
}

type Stocks struct {
	Stocks []Stock
}

func NewXQUser() *XQUser {
	x := new(XQUser)
	return x
}

func main() {
	fmt.Println("Hello Xueqiu!")
	deals := readFile("D:\\20140309 历史成交查询.txt")
	//fmt.Println(deals)
	xqUser := NewXQUser()
	xqUser.username = "********"
	//xqUser.telephone = "********"
	xqUser.password = "********"
	if xqUser.login() == true {
		for _, deal := range deals {
			st := xqUser.getStock(deal.code)
			if st.Code == "" || st.Name == "" {
				continue
			}
			postStr := "#雪球助手# 于"
			postStr += deal.date + " " + deal.time
			postStr += " " + deal.deal
			postStr += " $" + st.Name + "(" + st.Code + ")$"
			postStr += " 买入价" + deal.callPrice
			postStr += " 现价为:" + st.deatil.Current
			//postStr += " 涨幅:" + st.deatil.Percentage + "%"
			//postStr += " 涨跌额" + st.deatil.Change
			xqUser.post(postStr)
			time.Sleep(6 * time.Second)
		}
	}

}

func readFile(fileName string) (deals []StockDeal) {
	s := []StockDeal{}
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("open File Error!")
		return s
	}
	defer file.Close()
	rd := bufio.NewReader(file)
	decoder := mahonia.NewDecoder("gb18030")
	if decoder == nil {
		fmt.Println("编码不存在!")
		return s
	}

	for {
		str, err := rd.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}
		str = decoder.ConvertString(str)

		if strings.HasPrefix(str, "20") {
			strs := strings.Fields(str)
			fmt.Println(strs)
			var deal StockDeal
			deal.date = strs[0]
			deal.time = strs[1]
			deal.code = strs[2]
			deal.name = strs[3]
			deal.deal = strs[4]
			deal.callPrice = strs[5]
			s = append(s, deal)

		}
	}
	return s
}

func (user *XQUser) login() (b bool) {
	currentTime := time.Now().UnixNano() / 1000000
	loginUrl := "http://xueqiu.com/service/poster"
	postUrl := "/provider/oauth/token"
	tmd5 := md5.New()
	tmd5.Write([]byte(user.password))
	user.password = hex.EncodeToString(tmd5.Sum(nil))
	//fmt.Println("Md5 : " + user.password)
	post_arg := url.Values{
		"url":                {postUrl},
		"data[username]":     {user.username},
		"data[areacode]":     {"86"},
		"data[password]":     {user.password},
		"data[telephone]":    {user.telephone},
		"data[remember_me]":  {"1"},
		"data[access_token]": {""}, //Pm0HVhQUxPHTd0osDjwJ4Y
		"data[_]":            {strconv.FormatInt(currentTime, 10)}}
	//client := &http.Client{nil, nil, jar}
	client := new(http.Client)
	reqest, err := http.NewRequest("POST", loginUrl,
		strings.NewReader(post_arg.Encode()))
	//req.Header.Set("Referer", loginUrl)

	if err != nil {
		log.Fatal(err)
		fmt.Println("Error")
	}

	reqest.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	reqest.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	reqest.Header.Add("Accept-Encoding", "gzip,deflate")
	reqest.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	reqest.Header.Add("Connection", "keep-alive")
	reqest.Header.Add("Host", "login.sina.com.cn")
	reqest.Header.Add("Referer", "http://weibo.com/")
	reqest.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")
	response, err := client.Do(reqest)
	defer response.Body.Close()

	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(0)
	}
	//fmt.Println("Resp : " + strconv.Itoa(response.StatusCode))

	if response.StatusCode == 200 {
		body := getData(response)
		err = json.Unmarshal(body, user)
		if err != nil {
			fmt.Println("Error:", err)
		}
		user.cookies = response.Cookies()
		return true
	}
	return false

}

func (user *XQUser) post(postTxt string) {
	postUrl := "http://xueqiu.com/service/poster"
	jsonUrl := "/statuses/update.json"
	currentTime := time.Now().UnixNano() / 1000000
	post_arg := url.Values{
		"url":                {jsonUrl},
		"data[status]":       {postTxt},
		"data[title]":        {""},
		"data[access_token]": {user.Access_token},
		"data[_]":            {strconv.FormatInt(currentTime, 10)}}
	fmt.Println(postUrl + postTxt)
	client := new(http.Client)
	reqest, err := http.NewRequest("POST", postUrl,
		strings.NewReader(post_arg.Encode()))

	if err != nil {
		log.Fatal(err)
		fmt.Println("Error")
	}

	reqest.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	reqest.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	reqest.Header.Add("Accept-Encoding", "deflate")
	reqest.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	reqest.Header.Add("Connection", "keep-alive")
	reqest.Header.Add("Host", "login.sina.com.cn")
	reqest.Header.Add("Referer", "http://weibo.com/")
	reqest.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")
	for i := range user.cookies {
		reqest.AddCookie(user.cookies[i])
	}

	response, err := client.Do(reqest)
	defer response.Body.Close()

	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(0)
	}

	if response.StatusCode == 200 {
		getData(response)
	} else {
		fmt.Println(response.Status)
	}

}

func (user *XQUser) getStock(code string) (st Stock) {
	currentTime := time.Now().UnixNano() / 1000000
	sendUrl := "http://xueqiu.com/stock/search.json?size=10"
	sendUrl += "&code="
	sendUrl += code
	sendUrl += "&access_token="
	sendUrl += user.Access_token
	sendUrl += "&_="
	sendUrl += strconv.FormatInt(currentTime, 10)
	fmt.Println("SendUrl : " + sendUrl)
	response, err := http.Get(sendUrl)
	fmt.Println("Resp : " + strconv.Itoa(response.StatusCode))
	if err != nil {
		panic(err.Error())
	}
	body := getData(response)
	var stocks Stocks
	json.Unmarshal(body, &stocks)
	if len(stocks.Stocks) > 0 {
		st = stocks.Stocks[0]
		st.deatil = user.getDetail(st.Code)
	}
	return st
}

func (user *XQUser) getDetail(code string) (sd StockDetail) {
	currentTime := time.Now().UnixNano() / 1000000
	sendUrl := "http://xueqiu.com/stock/quote.json?"
	sendUrl += "code="
	sendUrl += code
	sendUrl += "&access_token="
	sendUrl += user.Access_token
	sendUrl += "&_="
	sendUrl += strconv.FormatInt(currentTime, 10)
	fmt.Println("SendUrl : " + sendUrl)
	response, err := http.Get(sendUrl)
	fmt.Println("Resp : " + strconv.Itoa(response.StatusCode))
	if err != nil {
		panic(err.Error())
	}
	body := getData(response)
	var qu Quotes
	json.Unmarshal(body, &qu)
	if len(qu.Quotes) > 0 {
		sd = qu.Quotes[0]
	}
	return sd

}

func getData(response *http.Response) (data []byte) {
	if response.StatusCode == 200 {
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			fz, err := gzip.NewReader(response.Body)
			if err != nil {
				fmt.Println("Gzip error.") //解压失败（还是读取原来文件）gz文件还是读取原始文件
			} else {
				r := bufio.NewReader(fz) //解压成功后读取解压后的文件
				data, _, _ = r.ReadLine()
			}
		default:
			dataByte, _ := ioutil.ReadAll(response.Body)
			data = dataByte
		}
		response.Body.Close()

		//fmt.Println(string(body))
	} else {
		fmt.Println(response.Status)
	}
	return data
}
