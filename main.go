package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var BaseURL string

func init() {
	// .env dosyasını yükleyin
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// BOT_TOKEN çevresel değişkenini al
	BotToken := os.Getenv("BOT_TOKEN")
	if BotToken == "" {
		fmt.Println("BOT_TOKEN is not set in the .env file")
	}

	// BaseURL'yi oluştur
	BaseURL = "https://api.telegram.org/bot" + BotToken + "/"
}

type Update struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		Chat struct {
			ID int `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

var chatID int
var isRunning bool
var monitoredURL = "https://www.sibervatan.org/"

func sendMessage(chatID int, text string) {
	url := BaseURL + "sendMessage"
	data := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	jsonData, _ := json.Marshal(data)
	_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}

func getUpdates(offset int) ([]Update, error) {
	url := fmt.Sprintf("%sgetUpdates?offset=%d", BaseURL, offset)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Ok     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Result, nil
}

func getWebsite(url string) (string, time.Duration) {
	start := time.Now()
	resp, err := http.Get(url)
	responseTime := time.Since(start)

	if err != nil {
		fmt.Printf("Error fetching the page: %s\n", err)
		return "There is an issue with the website", responseTime
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Website %s is UP (Status Code: %d)\n", url, resp.StatusCode)
		return fmt.Sprintf("Website %s is UP (Status Code: %d, Response Time: %d ms)", url, resp.StatusCode, responseTime.Milliseconds()), responseTime
	} else {
		fmt.Printf("Website %s is DOWN (Status Code: %d)\n", url, resp.StatusCode)
		return fmt.Sprintf("Website %s is DOWN (Status Code: %d, Response Time: %d ms)", url, resp.StatusCode, responseTime.Milliseconds()), responseTime
	}
}

func main() {
	offset := 0

	for {
		updates, err := getUpdates(offset)

		if err != nil {
			fmt.Println("Error getting updates:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		if len(updates) > 0 {
			for _, update := range updates {
				chatID = update.Message.Chat.ID
				text := update.Message.Text

				if strings.HasPrefix(text, "/url") {
					parts := strings.Split(text, " ")
					if len(parts) == 2 {
						monitoredURL = parts[1]
						sendMessage(chatID, "Takip edilecek URL güncellendi: "+monitoredURL)
					} else {
						sendMessage(chatID, "Lütfen geçerli bir URL girin: /url <yeni-url>")
					}
				}

				switch text {
				case "/start":
					sendMessage(chatID, "Komutlar:\n/begin - Web sitesi takibini başlat\n/stop - Web sitesi takibini durdur\n/url <url> - Takip edilecek URL'yi değiştir\n/merhaba - ?\n/start - Komutların özetini gösterir")
				case "/begin":
					if !isRunning {
						isRunning = true
						sendMessage(chatID, "Takip başlatıldı. Web sitesi durumu izleniyor.")
					} else {
						sendMessage(chatID, "Zaten aktif durumda.")
					}
				case "/stop":
					if isRunning {
						isRunning = false
						sendMessage(chatID, "Takip durduruldu.")
					} else {
						sendMessage(chatID, "Zaten durdurulmuş.")
					}
				case "/merhaba":
					sendMessage(chatID, "Galatasaray!")
				}
				offset = update.UpdateID + 1
			}
		}

		if isRunning {
			statusMessage, responseTime := getWebsite(monitoredURL)
			sendMessage(chatID, statusMessage)

			if responseTime.Milliseconds() > 100 {
				sendMessage(chatID, "UYARI: Web sitesi yanıt süresi çok yavaş! (Response Time: "+fmt.Sprintf("%d", responseTime.Milliseconds())+" ms)")
			}
		}

		time.Sleep(5 * time.Second)
	}
}
