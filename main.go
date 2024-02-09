package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type Product struct {
	Name  string
	Price string
	Url   string
}

func main() {

	file, err := os.Create("products.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var products []Product

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.WindowSize(1280, 1080),
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36"),
		chromedp.Flag("blink-settings", "scriptEnabled=false"),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	for i := 1; i < 250; i++ {
		page := strconv.Itoa(i)
		var productNodes []*cdp.Node
		if err := chromedp.Run(ctx,
			chromedp.Navigate("https://poizonshop.ru/products?brands=Nike&page="+page+"&perPage=40&sizeType=EU&sizeValue=42.5&sort=by-relevance"),

			chromedp.Evaluate(`window.scrollTo(0, document.documentElement.scrollHeight)`, nil),
			// Slow down the action so we can see what happen.
			chromedp.Sleep(1*time.Second),
			chromedp.Nodes(".product-card_product_card__5aPyG", &productNodes, chromedp.ByQueryAll),
		); err != nil {
			panic(err)
		}

		for _, node := range productNodes {
			var name, price, url string
			url, _ = (node.Attribute("href"))
			err := chromedp.Run(ctx,
				chromedp.Text(".product-card_name__amzGC", &name, chromedp.ByQuery, chromedp.FromNode(node)),
				chromedp.Text(".product-card-price_product_card_price__ei89N", &price, chromedp.ByQuery, chromedp.FromNode(node)),
			)

			if err != nil {
				log.Fatal("Error:", err)
			}

			product := Product{
				Name:  name,
				Price: price,
				Url:   url,
			}

			products = append(products, product)
		}
	}

	var data [][]string
	for _, p := range products {
		data = append(data, []string{"https://poizonshop.ru" + p.Url, p.Name, p.Price})
	}
	writer.WriteAll(data)
}
