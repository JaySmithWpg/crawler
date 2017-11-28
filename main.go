package main

import (
	"fmt"
	"github.com/JaySmithWpg/crawler/downloader"
	//"github.com/JaySmithWpg/crawler/filter"
	//"github.com/JaySmithWpg/crawler/parser"
	//"github.com/JaySmithWpg/crawler/utils"
	"net/url"
	"sync"
)

//just a throwaway main function
func main() {
	var wg sync.WaitGroup

	d := downloader.Create()

	u, _ := url.Parse("http://www.google.com")
	d.Request(u)
	d.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for resp := range d.Completed() {
			fmt.Printf(resp.Status)
		}
	}()
	wg.Wait()
}
