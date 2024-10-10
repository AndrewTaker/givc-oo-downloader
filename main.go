package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var (
	report, year, login, password string
	cabinet                       *Cabinet
)

func main() {
	flag.StringVar(&report, "r", "", "report type [oo1, oo2]")
	flag.StringVar(&year, "y", "", "report year 2016-2024")
	flag.StringVar(&login, "l", "", "miccedu login")
	flag.StringVar(&password, "p", "", "miccedu password")
	flag.Parse()

	if len(os.Args[1:]) != 8 { // mandatory args are 4
		flag.PrintDefaults()
		log.Fatalf("must provide all args")
	}

	cabinet = NewCabinet(map[string]string{
		"ologin":    login,
		"opassword": password,
		"source":    "direct",
		"ltype":     "default",
		"ocel":      "2",
		"ocf":       report,
		"ocy":       year,
	})

	folder := fmt.Sprintf("./%s_%s", cabinet.ReportName, cabinet.ReportYear)

	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		log.Fatalf("Failed to create folder: %v", err)
	}

	resp, err := cabinet.LoadReportPage()
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println(err)
	}

	log.Println(doc.Find(`div p[id="youare"]`).Text())
	log.Println(doc.Find("h1").Text())

	var wg sync.WaitGroup
	fileChan := make(chan string)

	go func() {
		for href := range fileChan {
			fileUrl := cabinet.BaseUrl + href
			err := saveFile(fileUrl, folder, cabinet.ReportName)
			if err != nil {
				log.Printf("failed to save file from %s: %v", fileUrl, err)
			} else {
				log.Printf("successfully saved file from %s", fileUrl)
			}
		}
	}()

	doc.Find(`tr[id^="tr"]`).Each(func(i int, s *goquery.Selection) {
		jumper := s.Children().Find(`button[onclick^="reopenJumper"]`).AttrOr("onclick", "")
		if jumper != "" {
			wg.Add(1)
			go cabinet.doStuff(cabinet.parseURL(jumper), &wg, fileChan, folder)
		}
	})

	wg.Wait()
	close(fileChan)

	log.Println("done do done")
}

func saveFile(url, folder, report string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file from %s: %w", url, err)
	}
	defer resp.Body.Close()

	parts := strings.Split(url, "/")
	if len(parts) < 7 {
		return fmt.Errorf("invalid URL structure: %s", url)
	}

	fileName := strings.ReplaceAll(fmt.Sprintf("%s_%s", parts[len(parts)-2], parts[len(parts)-1]), report+"_", "")
	filePath := filepath.Join(folder, fileName)

	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %w", filePath, err)
	}

	return nil
}

func parseArgs(s string) []string {
	cleared := strings.ReplaceAll(s, ", ", ",")
	cleared = strings.Trim(cleared, "reopenJumper()")
	cleared = strings.ReplaceAll(cleared, "\"", "")
	return strings.Split(cleared, ",")
}
