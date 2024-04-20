package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"sync"

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

	var wg sync.WaitGroup

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	numPages := 250
	pageChan := make(chan int, numPages)
	resultChan := make(chan []Product, numPages)

	// Отправка номеров страниц в канал
	for i := 1; i <= numPages; i++ {
		pageChan <- i
	}
	close(pageChan)

	// Запуск горутин для обработки каждой страницы
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for page := range pageChan {
				products := getProducts(ctx, page)
				if len(products) > 0 {
					resultChan <- products
				}
			}
		}()
	}

	// Закрытие resultChan после завершения всех горутин
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Получение результатов из горутин и добавление в общий список продуктов
	for products := range resultChan {
		for _, p := range products {
			writer.Write([]string{"https://poizonshop.ru" + p.Url, p.Name, p.Price})
		}
	}
}

// Функция для получения данных о продуктах с одной страницы
func getProducts(ctx context.Context, page int) []Product {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("blink-settings", "scriptEnabled=false"),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36"),
	)

	ctx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	pageStr := strconv.Itoa(page)
	var productNodes []*cdp.Node
	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://poizonshop.ru/products?brands=Nike&page="+pageStr+"&perPage=40&sizeType=EU&sizeValue=42.5&sort=by-relevance"),
		chromedp.Nodes(".product-card_product_card__5aPyG", &productNodes, chromedp.ByQueryAll),
	); err != nil {
		log.Printf("Error getting products from page %s: %s", pageStr, err)
		return nil
	}

	var products []Product
	for _, node := range productNodes {
		var name, price, url string
		url, _ = node.Attribute("href")
		if err := chromedp.Run(ctx,
			chromedp.Text(".product-card_name__amzGC", &name, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(".product-card-price_product_card_price__ei89N", &price, chromedp.ByQuery, chromedp.FromNode(node)),
		); err != nil {
			log.Printf("Error getting product details from page %s: %s", pageStr, err)
			continue
		}
		products = append(products, Product{Name: name, Price: price, Url: url})
	}
	return products
}
