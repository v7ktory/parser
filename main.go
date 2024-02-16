package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type Scraper interface {
	Scrape(ctx context.Context, pageNum string, bootItem chan<- []BootItem)
}
type BootItem struct {
	Name  string
	Price string
	URL   string
}

func NewBootItem() *BootItem {
	return &BootItem{}
}

func (b *BootItem) Scrape(ctx context.Context, pageNum string, bootItem chan<- []BootItem) {
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("blink-settings", "scriptEnabled=false"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36"),
	}

	context, cancel := chromedp.NewExecAllocator(ctx, options...)
	defer cancel()

	context, cancel = chromedp.NewContext(context)
	defer cancel()

	var itemNode []*cdp.Node
	if err := chromedp.Run(context,
		chromedp.Navigate("https://poizonshop.ru/products?brands=Nike&page="+pageNum+"&perPage=40&sizeType=EU&sizeValue=42.5&sort=by-relevance"),
		chromedp.Nodes(".product-card_product_card__5aPyG", &itemNode, chromedp.ByQueryAll),
	); err != nil {
		log.Fatalf("failed getting nodes: %s", err)
	}

	var bootItems []BootItem
	for _, n := range itemNode {
		var name, price string
		URL, ok := n.Attribute("href")
		if !ok {
			log.Fatal("failed getting nodes")
		}
		if err := chromedp.Run(context,
			chromedp.Text(".product-card_name__amzGC", &name, chromedp.ByQuery, chromedp.FromNode(n)),
			chromedp.Text(".product-card-price_product_card_price__ei89N", &price, chromedp.ByQuery, chromedp.FromNode(n)),
		); err != nil {
			log.Fatalf("failed getting nodes: %s", err)
		}

		bootItems = append(bootItems, BootItem{
			Name:  name,
			Price: price,
			URL:   URL,
		})
	}
	bootItem <- bootItems
}

func main() {
	file := CreateFile("test.csv")
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	bootItem := make(chan []BootItem)

	scraper := NewBootItem()
	go scraper.Scrape(context.Background(), "1", bootItem)

	for item := range bootItem {
		err := writer.Write([]string{item[0].Name, item[0].Price, item[0].URL})
		if err != nil {
			log.Fatalf("failed writing csv: %s", err)
		}
	}
}

func CreateFile(fileName string) *os.File {
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	return file
}
