package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type stringSlice []string

func (s *stringSlice) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func (s stringSlice) String() string {
	return strings.Join(s, "\n")
}

func (s *stringSlice) Values() []string {
	return *s
}

func main() {
	domains := flag.String("d", "", "file containing list of domains or ip addresses seperated by newlines, uses stdin if left empty")
	workers := flag.Uint("c", uint(runtime.NumCPU()), "number of concurrent requests")
	timeout := flag.Uint("t", 5, "timeout in seconds")
	portList := flag.String("p", "443,80", "comma seperated list of ports to probe")
	insecure := flag.Bool("k", true, "ignore SSL errors")
	userAgent := flag.String("a", "pokehttp: https://github.com/bruston/pokehttp", "user-agent header to use")
	redirects := flag.Bool("f", true, "follow redirects")
	headers := &stringSlice{}
	flag.Var(headers, "H", "add a header to the request, eg: \"Foo: bar\", can be specified multiple times")
	flag.Parse()

	var input io.ReadCloser
	if *domains == "" {
		input = os.Stdin
	} else {
		f, err := os.Open(*domains)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening domain/url list: %v", err)
			os.Exit(1)
		}
		input = f
	}
	defer input.Close()

	work := make(chan string)
	s := bufio.NewScanner(input)
	go func() {
		for s.Scan() {
			work <- s.Text()
		}
		close(work)
	}()

	client := &http.Client{
		Timeout: time.Second * time.Duration(*timeout),
	}
	if !*redirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	transport := &http.Transport{
		MaxIdleConns:      30,
		IdleConnTimeout:   time.Second,
		DisableKeepAlives: true,
	}
	if *insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client.Transport = transport

	ports := cleanPorts(*portList)
	wg := &sync.WaitGroup{}
	for i := 0; i < int(*workers); i++ {
		wg.Add(1)
		go func() {
			for v := range work {
				if strings.HasPrefix(v, "http") {
					code, size, title, err := doReq(client, v, headers.Values(), *userAgent)
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
						code, size, title, err := doReq(client, url, headers.Values(), *userAgent)
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

func doReq(client *http.Client, url string, headers []string, userAgent string) (int, int, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, 0, "", err
	}
	for _, v := range headers {
		pair := strings.Split(v, ":")
		if len(pair) == 1 {
			req.Header.Add(pair[0], "")
			continue
		}
		pair[1] = strings.TrimLeft(pair[1], " ")
		if strings.ToLower(pair[0]) == "host" {
			req.Host = pair[1]
			continue
		}
		req.Header.Add(pair[0], strings.Join(pair[1:], ":"))
	}
	req.Header.Set("User-Agent", userAgent)
	req.Close = true
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
