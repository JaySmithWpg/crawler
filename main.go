package main

import (
	"fmt"
	"github.com/JaySmithWpg/crawler/downloader"
	"github.com/JaySmithWpg/crawler/filter"
	"github.com/JaySmithWpg/crawler/parser"
	//"github.com/JaySmithWpg/crawler/utils"
	"net/url"
	"sync"
)

//just a throwaway main function
func main() {
	var wg sync.WaitGroup

	d := downloader.Create()
	p := parser.Create()
	f := filter.Create()

	u, _ := url.Parse("http://www.google.com")
	d.Request(u)
	//	d.Close()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer p.Close()
		for resp := range d.Completed() {
			p.Parse(resp)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for u := range p.Urls() {
			f.Test(u)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for u := range f.Results() {
			fmt.Printf("%s\n", u.String())
			d.Request(u)
		}
	}()
	wg.Wait()
}
