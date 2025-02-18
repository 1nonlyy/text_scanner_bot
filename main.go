package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	vision "cloud.google.com/go/vision/apiv1"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

func main() {
	//os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/Users/dauletzhumagali/Desktop/text_scanner_bot/your-service-account.json")
	er := godotenv.Load()
	if er != nil {
		log.Fatalf("Error loading .env file")
	}
	apikey := os.Getenv("API_KEY")
	bot, err := tgbotapi.NewBotAPI(apikey)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil && len(update.Message.Photo) > 0 { // If we got a message
			photo := update.Message.Photo[len(update.Message.Photo)-1]
			fileURL, err := bot.GetFileDirectURL(photo.FileID)
			if err != nil {
				log.Println("Error getting file URL:", err)
				continue
			}
			resText, err := scanText(fileURL)
			if err != nil {
				log.Printf("Image processing error: %v", err)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Error processing image"))
				continue
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Scanned text from image:\n"+resText)
			bot.Send(msg)
		}
	}
}

func scanText(imageURL string) (string, error) {
	ctx := context.Background()

	// Load Google Cloud Vision client
	client, err := vision.NewImageAnnotatorClient(ctx, option.WithCredentialsFile("config/textscanbot.json"))
	if err != nil {
		return "", err
	}
	defer client.Close()

	// Load the image from URL
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	// Step 2: Read the image as bytes
	imageData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %v", err)
	}

	img := &visionpb.Image{Content: imageData}
	annotations, err := client.DetectTexts(ctx, img, nil, 1)
	if err != nil {
		return "", fmt.Errorf("failed to detect text: %v", err)
	}

	if len(annotations) == 0 {
		return "No text found in the image", nil
	}

	// Return the first detected text
	return annotations[0].Description, nil
}
