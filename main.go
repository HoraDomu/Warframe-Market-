package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

const war_api = "https://api.warframe.market/v1"

type OrdersResponse struct {
	Payload struct {
		Item struct {
			ItemName string `json:"item_name"`
		} `json:"item"`
		Orders []struct {
			OrderType string  `json:"order_type"`
			Platinum  float64 `json:"platinum"`
			Visible   bool    `json:"visible"`
			// user.status can indicate online/offline; keep raw to allow future filtering
			User struct {
				Status string `json:"status"`
			} `json:"user"`
		} `json:"orders"`
	} `json:"payload"`
}

func sanitizeInput(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// common replacements
	replacer := strings.NewReplacer(
		" ", "_",
		"'", "",
		"\"", "",
		",", "",
		":", "",
		"/", "_",
		"(`", "",
		")", "",
		"(", "",
		"-", "_",
	)
	s = replacer.Replace(s)
	// collapse multiple underscores
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	return s
}

func fetchOrders(slug string) (*OrdersResponse, error) {
	url := fmt.Sprintf("%s/items/%s/orders", war_api, slug)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// polite user-agent
	req.Header.Set("User-Agent", "warframe-market-cli/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("item not found (API returned 404 for '%s')", slug)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var or OrdersResponse
	if err := json.Unmarshal(body, &or); err != nil {
		return nil, err
	}
	return &or, nil
}

func main() {
	var item string
	if len(os.Args) > 1 {
		// accept item as command line arguments
		item = strings.Join(os.Args[1:], " ")
	} else {
		fmt.Print("Item: ")
		// read whole line
		var input string
		if _, err := fmt.Scanln(&input); err != nil {
			// fallback to reading from STDIN manually (for names with spaces)
			data, err2 := ioutil.ReadAll(os.Stdin)
			if err2 != nil {
				log.Fatalf("failed to read input: %v", err2)
			}
			input = strings.TrimSpace(string(data))
		}
		item = input
	}

	if item == "" {
		log.Fatalln("no item provided")
	}

	slug := sanitizeInput(item)

	or, err := fetchOrders(slug)
	if err != nil {
		log.Fatalf("error fetching orders: %v", err)
	}

	// compute lowest sell and highest buy from visible orders
	var lowestSell *float64
	var highestBuy *float64
	for _, o := range or.Payload.Orders {
		if !o.Visible {
			continue
		}
		switch strings.ToLower(o.OrderType) {
		case "sell":
			p := o.Platinum
			if lowestSell == nil || p < *lowestSell {
				lowestSell = &p
			}
		case "buy":
			p := o.Platinum
			if highestBuy == nil || p > *highestBuy {
				highestBuy = &p
			}
		}
	}

	fmt.Printf("Item: %s\n", or.Payload.Item.ItemName)
	if lowestSell != nil {
		// print without trailing .0 when integer
		if float64(int64(*lowestSell)) == *lowestSell {
			fmt.Printf("Lowest sell: %d platinum\n", int64(*lowestSell))
		} else {
			fmt.Printf("Lowest sell: %.2f platinum\n", *lowestSell)
		}
	} else {
		fmt.Println("Lowest sell: (no visible sell orders)")
	}
	if highestBuy != nil {
		if float64(int64(*highestBuy)) == *highestBuy {
			fmt.Printf("Highest buy: %d platinum\n", int64(*highestBuy))
		} else {
			fmt.Printf("Highest buy: %.2f platinum\n", *highestBuy)
		}
	} else {
		fmt.Println("Highest buy: (no visible buy orders)")
	}
}
