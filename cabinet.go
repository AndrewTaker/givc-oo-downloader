package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Cabinet struct {
	Client     *http.Client
	BaseUrl    string
	ReportYear string
	ReportName string
}

func NewCabinet(cabinetCookies map[string]string) *Cabinet {
	var c Cabinet
	c.BaseUrl = "https://cabinet.miccedu.ru"

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal("cabinet.NewCabinet: could not initiate cookies")
	}

	rawUrl, err := url.Parse(c.BaseUrl)
	if err != nil {
		log.Fatal("cabinet.NewCabinet: could not parse url")
	}

	var cookies []*http.Cookie
	for k, v := range cabinetCookies {
		cookies = append(cookies, &http.Cookie{Name: k, Value: v})
	}

	jar.SetCookies(rawUrl, cookies)
	c.Client = &http.Client{Jar: jar}

	c.ReportName = report
	c.ReportYear = year

	return &c
}

func (c *Cabinet) request(url string) (*http.Response, error) {
	var req *http.Request
	var err error

	if !strings.HasPrefix(url, c.BaseUrl) {
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", c.BaseUrl, url), nil)
	} else {
		req, err = http.NewRequest(http.MethodGet, url, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("cabinet.request: request err for %s %s", url, err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cabinet.request: response err for %s %s", url, err)
	}

	return resp, nil
}

func (c *Cabinet) LoadReportPage() (*http.Response, error) {
	resp, err := c.request(fmt.Sprintf("%s/object", c.BaseUrl))
	if err != nil {
		return nil, fmt.Errorf("cabinet.LoadReportPage err: %s", err)
	}

	return resp, err
}

func (c *Cabinet) doStuff(u string, wg *sync.WaitGroup, fileChan chan<- string, folder string) error {
	defer wg.Done()
	parsedURL, err := url.Parse(u)
	if err != nil {
		return fmt.Errorf("cabinet.doStuff: invalid URL %s: %s", u, err)
	}

	resp, err := c.request(u)
	if err != nil {
		return fmt.Errorf("cabinet.doStuff: err %s", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return fmt.Errorf("cabinet.doStuff: failed to parse document: %s", err)
	}

	container := parsedURL.Query().Get("container")

	switch container {
	case "jslisttr":
		c.processJSListTR(doc, wg, fileChan, folder)

	case "juploadtr":
		href := c.extractUploadHref(doc)
		if href != "" {
			fileChan <- href
		}
	}

	return nil
}

func (c *Cabinet) processJSListTR(doc *goquery.Document, wg *sync.WaitGroup, fileChan chan<- string, folder string) {
	doc.Find(`tr[id^="tr"]`).Each(func(i int, s *goquery.Selection) {
		jumper := s.Children().Find(`button[onclick^="reopenJumper"]`).AttrOr("onclick", "")
		if jumper != "" {
			wg.Add(1)
			go c.doStuff(c.parseURL(jumper), wg, fileChan, folder)
		}
	})
}

func (c *Cabinet) extractUploadHref(doc *goquery.Document) string {
	return doc.Find(`a[href^="/excel/"]`).AttrOr("href", "")
}

func (c *Cabinet) createFileName(s string) string {
	ext := path.Ext(s)
	args := strings.Split(s, "/")
	org := strings.ReplaceAll(args[8], c.ReportName, "")
	org = strings.ReplaceAll(org, ext, "")
	return fmt.Sprintf("%s%s%s", args[7], org, ext)
}

func (c *Cabinet) parseURL(jumperVal string) string {
	v := parseArgs(jumperVal)

	u, err := url.Parse(fmt.Sprintf("%s/object/ajax/edit.php", c.BaseUrl))
	if err != nil {
		log.Fatal(err)
	}

	q := u.Query()
	q.Set("id", v[0])
	q.Set("pid", v[1])
	q.Set("type", v[2])
	q.Set("form", v[3])
	q.Set("reqtype", v[4])
	q.Set("container", v[5]+"tr")
	q.Set("edulevel", "2")

	u.RawQuery = q.Encode()

	return u.String()
}
