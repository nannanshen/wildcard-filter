package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/panjf2000/ants/v2"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var Proxy_url = ""
var Domain string
var Dict string
var Pre string
var Suf string
var Filter string
var Schema string
var Output string
var Threads int
var File string



func main(){
	now := time.Now()
	domain := flag.String("d", "", "请输入域名")
	dict := flag.String("w", "", "请输入字典路径")
	file := flag.String("f", "", "请输入要检测的文件名")
	pre := flag.String("pre", "", "请输入前缀")
	suf := flag.String("suf", "", "请输入后缀")
	filter := flag.String("fil", "", "请输入要过滤的内容")
	schema := flag.String("p", "http", "请输入协议（http/https）")
	output := flag.String("o", "res.txt", "请输入要输出的文件名")
	threads := flag.Int("t", 50, "请输入并发数")
	flag.Parse()
	Domain = *domain
	Dict = *dict
	File = *file
	Pre = *pre
	Suf = *suf
	Filter = *filter
	Schema = *schema
	Output = *output
	Threads = *threads
	//fmt.Printf("%s:%s\n", *domain, Dict)
	CheckDomains(Output)
	slice := time.Since(now)
	fmt.Printf("程序运行%s秒",slice)
	//GetHeader("https://a.yiche.com",Proxy_url)

}







func GetHeader(url1 string,proxy_url string)(string){
	var httpclient *http.Client
	if proxy_url !=""{
		proxy,_ := url.Parse(proxy_url)
		httpclient = &http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyURL(proxy),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 10 * time.Second,
		}
	}else{
		httpclient = &http.Client{
			Transport: &http.Transport{
				//Proxy:           http.ProxyURL(proxy),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 10 * time.Second,
		}
	}

	httpclient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := httpclient.Get(url1)
	if err != nil{
		fmt.Println(err)
		return "err"
	}
	location := resp.Header.Get("Location")
	return location

}

func GetLinesInFile(fileName string) ([]string, error) {
	result := []string{}
	f, err := os.Open(fileName)
	if err != nil {
		return result, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

func Getdomains(domain string,filename string,pre string,suf string)([]string){
	result := []string{}
	subs,err := GetLinesInFile(filename)
	if err != nil{
		return result
	}
	for _,sub := range subs{
		dd := domain
		if suf != ""{
			sub = sub + suf
		}
		dd = sub + "." + dd
		if pre != ""{
			dd = pre + dd
		}
		result = append(result, dd)
	}
	return result
}

func CheckHeader(url1 string)(bool){
	filter := []string{}
	if strings.Contains(Filter,","){
		filter = strings.Split(Filter,",")
	}else{
		filter = append(filter,Filter)
	}
	header := GetHeader(url1,Proxy_url)
	if header == "err"{
		return false
	}
	if GetParse(header).Host == GetParse(url1).Host && len(GetParse(header).Scheme) > len(GetParse(url1).Scheme){
		header = GetHeader(header,Proxy_url)
	}
	for _,fil := range filter{
		if strings.Contains(header,fil) || header=="err"{
			//fmt.Println("no")
			return false
		}
	}
	return  true
}

func GetParse(url1 string)(*url.URL){
	u,_ :=url.Parse(url1)
	return u
}

func WriteLine(line string ,output string){
	f,err:=os.OpenFile(output, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err!=nil{
		fmt.Printf("can not open the file")
		return
	}
	//2、写入文件
	defer f.Close()
	f.WriteString(line+"\n")
}

func CheckDomains(output string){
	defer ants.Release()
	domains := []string{}
	if Domain == ""{
		domains,_ = GetLinesInFile(File)
	}else{
		domains = Getdomains(Domain,Dict,Pre,Suf)
	}

	var tunnel = make(chan string, 50)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	pool,_ := ants.NewPool(1000)
	for i:=0;i<Threads;i++{
		go func() {
			_ = pool.Submit(func() {
				for domain:= range tunnel {
					url1 := Schema + "://" + domain
					res := CheckHeader(url1)
					if res {
						if output != "" {
							mutex.Lock()
							WriteLine(domain, output)
							mutex.Unlock()
						}
						fmt.Println(domain)
					}
					wg.Done()
				}
			})
		}()
	}



	for _,domain := range domains{
		wg.Add(1)
		tunnel <- domain
	}

	wg.Wait()

}


