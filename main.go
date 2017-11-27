package main

import (
	"fmt"
	"github.com/JaySmithWpg/crawler/downloader"
	"github.com/JaySmithWpg/crawler/parser"
	"github.com/JaySmithWpg/crawler/resolver"
	"github.com/JaySmithWpg/crawler/utils"
	"sync"
)

//just a throwaway main function
func main() {
	var wg sync.WaitGroup
	resolveRequests := make(chan resolver.Message)
	results, resolveErrors := resolver.Create(resolveRequests)

	downloadRequests := make(chan downloader.Message)
	downloads, downloadErrors := downloader.Create(downloadRequests)

	parseRequests := make(chan parser.Request)
	urls := parser.Create(parseRequests)

	wg.Add(5)
	go func() {
		defer wg.Done()
		defer close(downloadRequests)
		for result := range results {
			fmt.Printf("Resolved %s: %s\n", result.HostName(), result.Address())
			downloadRequest, ok := result.(downloader.Message)
			if ok {
				downloadRequests <- downloadRequest
			}
		}
	}()

	go func() {
		defer wg.Done()
		for err := range resolveErrors {
			fmt.Printf("Resolver Error: %s\n", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		defer close(parseRequests)
		for downloaded := range downloads {
			parseRequest, ok := downloaded.(parser.Request)
			if ok {
				fmt.Printf("File Downloaded: %s\n", parseRequest.Url())
				parseRequests <- parseRequest
			} else {
				fmt.Printf("Could not cast parser request\n")
			}
		}
	}()

	go func() {
		defer wg.Done()
		for downloadError := range downloadErrors {
			fmt.Printf("Download Error: %s - %s\n", downloadError.HostName(), downloadError.Error())
		}
	}()

	go func() {
		defer wg.Done()
		for url := range urls {
			fmt.Printf("Url Parsed: %s\n", url.String())
		}
	}()

	resolveRequests <- utils.CreateCrawlerRecord("http://www.google.com:80")
	resolveRequests <- utils.CreateCrawlerRecord("http://www.gotogle.com")
	resolveRequests <- utils.CreateCrawlerRecord("https://www.cnn.com")
	resolveRequests <- utils.CreateCrawlerRecord("https://www.amazon.ca")
	resolveRequests <- utils.CreateCrawlerRecord("https://www.amazon.com")

	close(resolveRequests)
	wg.Wait()
}
