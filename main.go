package main

import (
	"bufio"
	"crypto/tls"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	_ "io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/fatih/color"
	"golang.org/x/net/html"
)

var file = flag.String("r", "", "Input a url files")

func getTitle(body io.ReadCloser) string {
	if body != nil {
		tokenizer := html.NewTokenizer(body)
		title := "<title> tag missing"
		for {
			tokenType := tokenizer.Next()
			if tokenType == html.ErrorToken {
				err := tokenizer.Err()
				if err == io.EOF {
					break
				} else {
					title = err.Error()
				}
			}
			if tokenType == html.StartTagToken {
				token := tokenizer.Token()
				if "title" == token.Data {
					_ = tokenizer.Next()
					title = tokenizer.Token().Data
					break
				}
			}
		}
		title = strings.Join(strings.Fields(strings.TrimSpace(title)), " ")
		return title
	}
	return ""
}
func Url() {
	flag.Parse()
	if strings.HasSuffix(*file, ".txt") {
		file1, _ := os.Open(*file)
		defer file1.Close()
		f, err := os.Create("data.csv") //写文件
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
		f.WriteString("\xEF\xBB\xBF")
		writer := csv.NewWriter(f)
		writer.Write([]string{"地址", "状态", "server", "title"})
		var wg sync.WaitGroup
		scanner := bufio.NewScanner(file1)
		for scanner.Scan() {
			wg.Add(1) //计数器加1
			go func(url string) {
				defer wg.Done() //计数器减1
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				client := &http.Client{
					Timeout:   40 * time.Second,
					Transport: tr,
				}
				resp, rr := http.NewRequest("GET", url, nil)
				resp.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 Safari/537.36")
				if rr != nil {
					panic(rr)
				}
				resp1, err := client.Do(resp)
				if err != nil {
					b := time.Now().Format("2006-01-02 15:04:05")
					fmt.Printf("[\033[35m%s\033[0m][\033[35m%s\033[0m][\033[31m无法访问\033[0m]\n", b, url)
					headline := []string{url, "无法访问", "", ""}
					writer.Write(headline)
					//writer.Flush()
					return
				}
				defer resp1.Body.Close()
				title := getTitle(resp1.Body)
				resp1.Body.Close()
				//color.Set(color.FgBlue)
				ser, dd := resp1.Header["Server"]
				if !dd {
					ser = []string{"None"}
				}
				b := time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("[\033[34m%s\033[0m][\033[34m%s\033[0m] [\033[34m%d\033[0m] [\033[34m%s\033[0m][\033[36m%s\033[0m]\n", b, url, resp1.StatusCode, ser, title)
				//color.Unset()
				s := fmt.Sprint(resp1.StatusCode)
				headline := []string{url, s, ser[0], title}
				writer.Write(headline)
				writer.Flush()
			}(scanner.Text())

		}
		wg.Wait() //阻塞
	}

}

func main() {
	start := time.Now().Unix()
	Url()
	end := time.Now().Unix()
	fmt.Printf("运行花费时间为%v秒", end-start)

}
