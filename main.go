// SPDX-License-Identifier: GPL-3.0-or-later
// Copyright (C) 2025 Ummuys

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	botCMD "github.com/Ummuys/TG_W/bot"
	dbCMD "github.com/Ummuys/TG_W/database"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initLogFile() (string, error) {
	now := time.Now().Format("2006-01-02")
	path := "logs/"
	fileName := path + now + ".log"

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		_, err = os.Create(fileName)
		if err != nil {
			return "", fmt.Errorf("сan't create a file: %w", err)
		}
	}
	return fileName, nil
}

func initZapLogger(file *os.File) (*zap.Logger, error) {
	cfg := zap.Config{
		Encoding:         "console", // Используем консольный кодировщик
		Level:            zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths:      []string{file.Name()},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey: "message",
			LevelKey:   "level",
			TimeKey:    "time",
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("[2006-01-02 15:04:05 [MSK]")) // Формат времени
			},
			EncodeLevel: zapcore.CapitalLevelEncoder, // Формат для уровня
		},
	}
	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("сan't init Zap logger: %w", err)
	}
	return logger, nil
}

func initBotNTools(botApi string) (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel, error) {
	bot, err := tgbotapi.NewBotAPI(botApi)
	if err != nil {
		return nil, nil, fmt.Errorf("can't create a bot: %w", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	updates.Clear()
	return bot, updates, nil
}

func chatManagement(wg *sync.WaitGroup, logger *zap.Logger, cancel context.CancelFunc, endCicle chan struct{}) {
	wg.Done()
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Write 'exit' to turn-off the bot.")
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if strings.ToLower(text) == "exit" {
			logger.Info("Alert: Starting to turn off the bot!")
			fmt.Println("Alert: Starting to turn off the bot!")
			cancel()
			logger.Info("Alert: Timer for 5 sec started!")
			fmt.Println("Alert: Timer for 5 sec started!")
			timer := time.NewTimer(5 * time.Second)
			<-timer.C
			logger.Info("Alert: Times out!")
			endCicle <- struct{}{}
			return
		} else {
			fmt.Println(" wrong command!")
		}
	}
}

func startBotLoop(bot *tgbotapi.BotAPI, updates tgbotapi.UpdatesChannel, logger *zap.Logger, ctx context.Context, endCicle chan struct{}, chatInfo map[int64]botCMD.ChatInfo, chatsChans map[int64]chan tgbotapi.Update) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	logger.Info("Bot started")
	for {
		select {
		case update := <-updates:
			{
				if update.Message == nil {
					continue
				}

				var (
					ch     chan tgbotapi.Update
					exists bool
				)

				chatId := update.Message.Chat.ID
				mu.Lock() // Need to process this information before other rutins read it.
				ch, exists = chatsChans[chatId]

				if !exists {

					ch = make(chan tgbotapi.Update, 5) // Make chan with buff. Max message - 5
					chatsChans[chatId] = ch
					logger.Info("Message: Started new chat\tID: " + strconv.FormatInt(chatId, 10))
					wg.Add(1)
					go botCMD.InitChat(bot, chatsChans, ch, &wg, ctx, logger, chatInfo, chatId)
				}
				mu.Unlock() // Unlock to read

				wg.Add(1)
				go func(wg *sync.WaitGroup, ctx context.Context, ch chan tgbotapi.Update, upd tgbotapi.Update) {
					defer wg.Done()
					select {
					case <-ctx.Done():
						return
					case ch <- upd:
						return
					}
				}(&wg, ctx, ch, update)

			} // End first case
		case <-endCicle:
			{
				wg.Wait()
				return
			} // End second case

		} //End select

	} // End for
}

func main() {

	// init file for logs
	fileName, err := initLogFile()
	if err != nil {
		panic(err)
	}

	// open file to read and write logs
	file, err := os.Open(fileName)
	if err != nil {
		panic(fmt.Sprintf("Err on open file: %v", err))
	}
	defer file.Close()

	// init logger
	logger, err := initZapLogger(file)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	logger.Info("Start logs")

	// init sync
	var wg sync.WaitGroup
	err = godotenv.Load() // find ".env"
	if err != nil {
		logger.Error(err.Error())
		panic(fmt.Errorf("can't open env: %w", err))
	}

	// init ChatInfo and take info from db
	ChatInfo := make(map[int64]botCMD.ChatInfo, 100)
	pool, pCtx, err := dbCMD.Connect(os.Getenv("CONNECT"))
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}
	defer pool.Close()

	err = dbCMD.CreateItems(logger, pool, pCtx)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	err = dbCMD.GetInfoFromTable(logger, pool, pCtx, ChatInfo)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	// init bot, chan for updates
	botApi := os.Getenv("BOT_API") // develop
	bot, updates, err := initBotNTools(botApi)
	if err != nil {
		logger.Error(err.Error())
		panic(err)
	}

	// map, where will be saved chatId and they chan
	chatsChans := make(map[int64]chan tgbotapi.Update)

	// sync rutins
	ctx, cancel := context.WithCancel(context.Background())
	endCicle := make(chan struct{})

	//function for neatly completing all routines and exiting the program
	wg.Add(1)
	go chatManagement(&wg, logger, cancel, endCicle)

	startBotLoop(bot, updates, logger, ctx, endCicle, ChatInfo, chatsChans)
	bot.StopReceivingUpdates()
	logger.Info("Message: Bot shut down")

	wg.Wait()
	if err = dbCMD.FillInfoFromMap(pCtx, pool, ChatInfo); err != nil {
		logger.Error(err.Error())
	}

	fmt.Println("END CHAT COUNT ROOTINS -->", runtime.NumGoroutine())
	buf := make([]byte, 1<<20) // 1MB буфера
	n := runtime.Stack(buf, true)
	fmt.Printf("=== Goroutines dump ===\n%s\n", buf[:n]) // Check: is all rutins shot down

}
