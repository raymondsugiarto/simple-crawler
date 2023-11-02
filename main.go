package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
)

type WebPage struct {
	Url       string
	NumLinks  int
	NumImages int
	LastFetch time.Time
}

type Asset struct {
	OldUrl string
	NewUrl string
}

func main() {
	pwd, _ := os.Getwd()

	metadata := flag.Bool("metadata", false, "show metadata")
	var urls string
	flag.StringVar(&urls, "urls", "bar", "a string var")
	flag.Parse()

	if len(flag.Args()) > 0 {
		var wg sync.WaitGroup
		for i := 0; i < len(flag.Args()); i++ {
			url := flag.Args()[i]
			// Increment the wait group counter
			wg.Add(1)
			go func(metadata bool, pwd, url string) {
				// Decrement the counter when the go routine completes
				defer wg.Done()

				// Call the function check
				GetHtml(metadata, pwd, url)
			}(*metadata, pwd, url)
		}
		// Wait for all the checkWebsite calls to finish
		wg.Wait()
	}
}

func PrintMetadata(webPage WebPage) {
	fmt.Println("site: ", webPage.Url)
	fmt.Println("num_links: ", webPage.NumLinks)
	fmt.Println("images: ", webPage.NumImages)
	fmt.Println("last_fetch: ", webPage.LastFetch)
	fmt.Println("=====================================")
}

func GetHtml(metadata bool, pwd, url string) {
	webPage := WebPage{
		Url:       url,
		NumLinks:  0,
		NumImages: 0,
	}

	assetUrls := &[]Asset{}

	c := colly.NewCollector(
		colly.MaxDepth(1),
	)

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		webPage.NumLinks++

	})

	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		webPage.NumImages++

		link := e.Attr("src")
		if !strings.HasPrefix(link, "src") {
			link = e.Request.AbsoluteURL(link)
		}
		CreateDir(pwd + "/html/asset-" + e.Request.URL.Host)
		DownloadFile(pwd+"/html/asset-"+e.Request.URL.Host, link, assetUrls)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("OnResponse", r.Request.URL)

		CreateDir(pwd + "/html")
		err := os.WriteFile(pwd+"/html/"+r.Request.URL.Host+".html", []byte(r.Body), 0644)
		if err != nil {
			fmt.Println("Error when save file", err)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		url = r.URL.Host
		webPage.LastFetch = time.Now()
	})

	// Set error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit(url)

	htmlPath := pwd + "/html/" + url + ".html"
	ReplaceAssetUrls(htmlPath, pwd, assetUrls)

	if metadata {
		PrintMetadata(webPage)
	}
}

func ReplaceAssetUrls(filePath, pwd string, assetUrls *[]Asset) {
	// Read the file contents
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	// Convert the file's contents to a string
	fileContents := string(data)

	for i := 0; i < len(*assetUrls); i++ {
		fileContents = strings.Replace(fileContents, (*assetUrls)[i].OldUrl, (*assetUrls)[i].NewUrl, -1)
	}

	// Write the modified contents back to the file
	err = ioutil.WriteFile(filePath, []byte(fileContents), 0644)
	if err != nil {
		return
	}
}

func DownloadFile(pwd, url string, assetUrls *[]Asset) {
	// Extract the file name from the URL
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	targetFile := pwd + "/" + fileName
	*assetUrls = append(*assetUrls, Asset{
		OldUrl: url,
		NewUrl: targetFile,
	})
	fmt.Println("Downloading file:", url)
	// Create a local file to save the asset
	localFile, err := os.Create(targetFile)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer localFile.Close()

	// Download the asset from the web and save it locally
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}
	defer response.Body.Close()

	_, err = io.Copy(localFile, response.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
	}

}

func CreateDir(directoryPath string) {

	// Check if the directory exists
	if _, err := os.Stat(directoryPath); os.IsNotExist(err) {
		// Directory doesn't exist, so create it
		err := os.MkdirAll(directoryPath, os.ModePerm)
		if err != nil {
			// fmt.Printf("Error creating directory: %v\n", err)
		} else {
			fmt.Printf("Directory created: %s\n", directoryPath)
		}
	}
}
