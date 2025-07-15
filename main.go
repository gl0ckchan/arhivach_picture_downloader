package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var baseURL string
var semaphore = make(chan struct{})
var wg sync.WaitGroup

func main() {
	link := flag.String("link", "", "link to thread")
	semaphore_flag := flag.Int("goroutines", 3, "number of goroutines running")
	flag.Parse()

	if *link == "" {
		fmt.Println("ERROR: -link is required")
		flag.Usage()
		os.Exit(1)
	}
	semaphore = make(chan struct{}, *semaphore_flag)

	u, _ := url.Parse(*link)
	url := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	baseURL = url
	fmt.Printf("\t%s\n", baseURL)

	resp, _ := http.Get(*link)
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("ERROR: ", err)
		return
	}

	if err := os.Mkdir("downloaded", 0755); err != nil && !os.IsExist(err) {
		fmt.Println("ERROR: Cannot create folder 'downloaded': ", err)
	}

	processAllProduct(doc)

	wg.Wait()
	fmt.Println("All downloads completed")
}

func processAllProduct(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "a" {
		processNode(n)

	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processAllProduct(c)
	}
}

func processNode(n *html.Node) {
	switch n.Data {
	case "a":
		isImageOrVideo := false
		HrefURL := ""
		for _, a := range n.Attr {
			if a.Key == "class" && a.Val == "img_filename" {
				isImageOrVideo = true
			}
			if a.Key == "href" {
				HrefURL = a.Val
			}
		}
		if isImageOrVideo && HrefURL != "" {
			if HrefURL[0] == '/' {
				HrefURL = baseURL + HrefURL
			}
			fmt.Println("Href URL: ", HrefURL)

			semaphore <- struct{}{}
			wg.Add(1)
			go func(url string) {
				defer wg.Done()
				defer func() { <-semaphore }()
				downloadFile(url)
			}(HrefURL)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processNode(c)
	}
}

func downloadFile(fileURL string) {
	resp, err := http.Get(fileURL)
	if err != nil {
		fmt.Println("ERROR downloading: ", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("ERROR: Failed to download", fileURL, "Status:", resp.Status)
		return
	}

	segments := strings.Split(fileURL, "/")
	fileName := segments[len(segments)-1]
	if fileName == "" {
		fmt.Println("ERROR: Invalid file name extracted from URL")
		return
	}

	file, err := os.Create("downloaded/" + fileName)
	if err != nil {
		fmt.Println("ERROR creating file:", err)
		return
	}
	defer file.Close()

	n, err := io.Copy(file, resp.Body)
	if err != nil {
		fmt.Println("ERROR saving file:", fileName, err)
		return
	}

	fmt.Println("File saved as", fileName, "size:", n)
}
