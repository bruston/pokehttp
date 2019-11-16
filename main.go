package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	domains := flag.String("d", "", "file containing list of domains or ip addresses seperated by newlines")
	workers := flag.Uint("c", uint(runtime.NumCPU()), "number of concurrent requests")
	timeout := flag.Uint("t", 5, "timeout in seconds")
	portList := flag.String("p", "80,433", "comma seperated list of ports to probe")
	host := flag.String("h", "", "host header")
	forwarded := flag.String("x", "", "X-Forwarded-For header")
	customKey := flag.String("ck", "", "add a custom header to all requests (key)")
	customVal := flag.String("cv", "", "add a custom header to all requests (value)")
	insecure := flag.Bool("k", false, "ignore SSL errors")
	userAgent := flag.String("a", "pokehttp: https://github.com/bruston/pokehttp", "user-agent header to use")
	flag.Parse()

	if *domains == "" {
		fmt.Println("Please specify a domain list with the -d flag.\n")
		flag.PrintDefaults()
		return
	}

	f, err := os.Open(*domains)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	work := make(chan string)
	s := bufio.NewScanner(f)
	go func() {
		for s.Scan() {
			work <- s.Text()
		}
		close(work)
	}()

	ports := cleanPorts(*portList)
	wg := &sync.WaitGroup{}

	client := &http.Client{Timeout: time.Second * time.Duration(*timeout)}
	if *insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	for i := 0; i < int(*workers); i++ {
		wg.Add(1)
		go func() {
			for v := range work {
				if strings.HasPrefix(v, "http") {
					code, size, title, err := doReq(client, v, *host, *forwarded, *customKey, *customVal, *userAgent)
					if err != nil {
						continue
					}
					fmt.Printf("%s %d %d %s\n", v, code, size, title)
					continue
				}
				for _, port := range ports {
					dst := make([]string, 0, 2)
					switch port {
					case "80":
						dst = append(dst, "http://"+v)
					case "443":
						dst = append(dst, "https://"+v)
					default:
						dst = append(dst, []string{fmt.Sprintf("http://%s:%s", v, port), fmt.Sprintf("https://%s:%s", v, port)}...)
					}
					for _, url := range dst {
						code, size, title, err := doReq(client, url, *host, *forwarded, *customKey, *customVal, *userAgent)
						if err != nil {
							continue
						}
						fmt.Printf("%s %d %d %s\n", url, code, size, title)
					}
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func cleanPorts(p string) []string {
	p = strings.TrimSuffix(p, ",")
	ports := strings.Split(p, ",")
	for i, v := range ports {
		ports[i] = strings.TrimSpace(v)
	}
	return ports
}

func doReq(client *http.Client, url, host, forwarded, customKey, customVal, userAgent string) (int, int, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, "", err
	}
	if host != "" {
		req.Host = host
	}
	if forwarded != "" {
		req.Header.Set("X-Forwarded-For", forwarded)
	}
	if customKey != "" {
		req.Header.Set(customKey, customVal)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()
	status := resp.StatusCode
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, "", err
	}
	size := len(b)
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(b))
	if err != nil {
		return 0, 0, "", err
	}
	title := doc.Find("title").Text()
	title = strings.Replace(title, "\r\n", "", -1)
	title = strings.Replace(title, "\n", "", -1)
	return status, size, title, nil
}
