package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/cors"
)

type Application struct {
	Title string `json:"title"`
	Data  struct {
		Name        string `json:"name"`
		Phone       string `json:"phone"`
		Company     string `json:"company,omitempty"`
		Email       string `json:"email,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"data"`
}

var bot *tgbotapi.BotAPI
var chatIDs = make(map[int64]bool)
var mu sync.Mutex

func main() {
	var err error
	bot, err = tgbotapi.NewBotAPI("6714752858:AAFzc2TWPi3pJxVgZinFLj63wBW0TggVmjk")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	mux := http.NewServeMux()
	mux.HandleFunc("/tgbot/add", handleApplication)

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://numbers-agency.ru/"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(mux)

	go func() {
		log.Println("Server is running on port 3000")
		log.Fatal(http.ListenAndServe(":3000", handler))
	}()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			handleBotCommand(update.Message)
		}
	}
}

func handleApplication(w http.ResponseWriter, r *http.Request) {
	var app Application

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &app)
	if err != nil {
		http.Error(w, "Неверный формат JSON", http.StatusBadRequest)
		return
	}

	message := fmt.Sprintf("Новая заявка\n\nТип: %s\nИмя: %s\nТелефон: %s\nКомпания: %s\nПочта: %s\nОписание: %s",
		app.Title, app.Data.Name, app.Data.Phone, app.Data.Company, app.Data.Email, app.Data.Description)

	mu.Lock()
	for chatID := range chatIDs {
		msg := tgbotapi.NewMessage(chatID, message)
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Не удалось отправить сообщение в Telegram для chatID %d: %v", chatID, err)
		}
	}
	mu.Unlock()

	response := map[string]string{"message": "Заявка успешно отправлена в Telegram"}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func handleBotCommand(message *tgbotapi.Message) {
	mu.Lock()
	chatIDs[message.Chat.ID] = true
	mu.Unlock()

	msg := tgbotapi.NewMessage(message.Chat.ID, "")

	switch message.Command() {
	case "start":
		msg.Text = "Привет! Я готов принимать ваши заявки."

	default:
		msg.Text = "Извините, я не знаю такую команду."
	}

	if _, err := bot.Send(msg); err != nil {
		log.Panic(err)
	}
}
