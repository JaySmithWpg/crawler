package main

import (
	"crawler/downloader"
	"crawler/resolver"
	"crawler/utils"
	"fmt"
	"sync"
)

//just a throwaway main function
func main() {
	var wg sync.WaitGroup
	resolveRequests := make(chan resolver.Message)
	results, resolveErrors := resolver.Create(resolveRequests)

	downloadRequests := make(chan downloader.Request)
	downloads, downloadErrors := downloader.Downloader(downloadRequests)

	wg.Add(4)
	go func() {
		defer wg.Done()
		defer close(downloadRequests)
		for result := range results {
			fmt.Printf("%s: %s\n", result.HostName(), result.Address().String())
			downloadRequest, ok := result.(downloader.Request)
			if ok {
				downloadRequests <- downloadRequest
			}
		}
	}()

	go func() {
		defer wg.Done()
		for err := range resolveErrors {
			fmt.Printf("Error: %s\n", err.Error())
		}
	}()

	go func() {
		defer wg.Done()
		for downloaded := range downloads {
			fmt.Printf("%s\n", downloaded.Headers())
		}
	}()

	go func() {
		defer wg.Done()
		for downloadError := range downloadErrors {
			fmt.Printf("%s: %s\n", downloadError.HostName(), downloadError.Error())
		}
	}()

	resolveRequests <- utils.CreateCrawlerRecord("http://www.google.com")
	resolveRequests <- utils.CreateCrawlerRecord("http://www.gotogle.com")
	resolveRequests <- utils.CreateCrawlerRecord("https://www.google.com")
	resolveRequests <- utils.CreateCrawlerRecord("https://www.amazon.ca")
	resolveRequests <- utils.CreateCrawlerRecord("https://www.amazon.com")

	close(resolveRequests)
	wg.Wait()
}
