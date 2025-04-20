package bot

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	weather "github.com/Ummuys/TWB/weather"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type ChatInfo struct {
	Id            int64
	UserState     string
	FavoriteCitys [3]string
}

func InitChat(bot *tgbotapi.BotAPI, usersChans map[int64]chan tgbotapi.Update, ch chan tgbotapi.Update, wg *sync.WaitGroup, ctx context.Context, logger *zap.Logger, usersInfo map[int64]ChatInfo, id int64) {
	defer wg.Done()
	defer delete(usersChans, id)
	chatInfo, exists := usersInfo[id]
	if !exists {
		chatInfo = ChatInfo{
			Id:            id,
			UserState:     "",
			FavoriteCitys: [3]string{"", "", ""},
		}
	}
	usersInfo[id] = Chat(ch, ctx, bot, logger, chatInfo)
}

func Chat(ch chan tgbotapi.Update, ctx context.Context, bot *tgbotapi.BotAPI, logger *zap.Logger, chatInfo ChatInfo) ChatInfo {
	chooseOption := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Погоду 🌧️"),
		tgbotapi.NewKeyboardButton("?"),
		tgbotapi.NewKeyboardButton("Об авторе"),
	))

	getBack := tgbotapi.NewReplyKeyboard(tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Вернуться обратно")))

	var resp string
	var update tgbotapi.Update
	var err error
	author :=
		`
	Автор: Евгений Егоров (Ummuys)
	Github: https://github.com/Ummuys
	Telegram: @Ummuys
	`

	chatID := chatInfo.Id
	userState := chatInfo.UserState
	favoriteCitys := chatInfo.FavoriteCitys
	pos := 0
	for pos = 0; pos < len(favoriteCitys); pos++ {
		if favoriteCitys[pos] == "" {
			break
		}
	}

	for {
		if userState == "END" {
			userState = ""
			resp = "/menu"
		} else {
			update, err = waitMessage(ch, ctx, logger, chatID)
			if err != nil {
				return ChatInfo{Id: chatID, UserState: userState, FavoriteCitys: favoriteCitys}
			}
			resp = update.Message.Text
			logger.Info("Message: New response\t" + "\tID: " + strconv.FormatInt(chatID, 10) + "\tuserState: \"" + userState + "\"" + "\tRESP: \"" + resp + "\"")
		}
		switch {
		case resp == "/start":
			{
				msg := tgbotapi.NewMessage(chatID, "Приветствую тебя в тестовом телеграмм боте от Ummuys!\n"+
					"Надеюсь ты сможешь найти удобные функции для себя!\n"+
					"Чтобы перейти к основному функционалу, напиши /menu")
				msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				bot.Send(msg)
			}
		case resp == "/menu":
			{
				msg := tgbotapi.NewMessage(chatID, "Что хотите узнать?")
				msg.ReplyMarkup = chooseOption
				bot.Send(msg)
				userState = "CHS"
			}
		case resp == "Вернуться обратно":
			{
				userState = "END"
			}
		case resp == "Очистить список":
			{
				userState = "END"
				for i := range favoriteCitys {
					favoriteCitys[i] = ""
				}
			}
		case userState == "CHS":
			{
				switch {
				case resp == "Погоду 🌧️" || resp == "Погоду":
					{
						userState = "WEA"
						ChooseCity(bot, chatID, favoriteCitys, logger)
					}
				case resp == "?":
					{
						msg := tgbotapi.NewMessage(chatID, "Новые функции будут, когда-нибудь . . .")
						msg.ReplyMarkup = getBack
						bot.Send(msg)
					}
				case resp == "Об авторе" || resp == "/about":
					{
						msg := tgbotapi.NewMessage(chatID, author)
						msg.ReplyMarkup = getBack
						bot.Send(msg)
					}
				default:
					{
						userState = "END"
						msg := tgbotapi.NewMessage(chatID, "Не найден такой функционал. . .")
						bot.Send(msg)
					}
				}
			}
		case userState == "WEA":
			{
				cityState, cityName, err := weather.GetWeather(resp)
				if err != nil {
					logger.Error(err.Error())
					bot.Send(tgbotapi.NewMessage(chatID, "Похоже проблемы с сайтом погоды, попробуйте позже. . . "))
				} else {
					if cityState == "" {
						logger.Error("Message: City doens't exists\t" + "\tID: " + strconv.FormatInt(chatID, 10) + "\tCity: \"" + cityName + "\"")
						bot.Send(tgbotapi.NewMessage(chatID, "Такой город не найден/не существует!"))
					} else {
						bot.Send(tgbotapi.NewMessage(chatID, "Нашел твой город, держи!"))
						bot.Send(tgbotapi.NewMessage(chatID, cityState))
						if !Сontains(favoriteCitys, cityName) {
							logger.Info("Message: Added city\t" + "City: " + cityName + "\tID: " + strconv.FormatInt(chatID, 10))
							favoriteCitys[pos%3] = cityName
							pos++
						}
					}
				}
				userState = "END"
			}

		default:
			logger.Error("Message: Command doens't exists\t" + "\tID: " + strconv.FormatInt(chatID, 10) + "\tRESP: \"" + resp + "\"")
			msg := tgbotapi.NewMessage(chatID, "Не найден такой функционал. . .")
			msg.ReplyMarkup = getBack
			bot.Send(msg)
		}
	}
}

func waitMessage(ch chan tgbotapi.Update, ctx context.Context, logger *zap.Logger, id int64) (tgbotapi.Update, error) {
	timer := time.NewTimer(1 * time.Minute)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()

	select {
	case <-ctx.Done():
		{
			logger.Debug("Cancel catched! Chat go to off\tID: " + strconv.FormatInt(id, 10))
			return tgbotapi.Update{}, fmt.Errorf("cancel catched")
		}
	case <-timer.C:
		{
			logger.Info("Time's out! Chat go to AFK\tID: " + strconv.FormatInt(id, 10))
			return tgbotapi.Update{}, fmt.Errorf("time out")
		}
	case update := <-ch:
		{
			return update, nil
		}
	}
}

func ChooseCity(bot *tgbotapi.BotAPI, chatID int64, priority_city [3]string, logger *zap.Logger) {
	var rows [][]tgbotapi.KeyboardButton
	for _, city := range priority_city {
		if city == "" {
			continue
		}
		row := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(city),
		)
		rows = append(rows, row)
	}
	if len(rows) > 0 {
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Очистить список")))
		logger.Info("Message: Delete all city\t" + "\tID: " + strconv.FormatInt(chatID, 10))
	}
	rows = append(rows, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("Вернуться обратно")))
	keyboard := tgbotapi.NewReplyKeyboard(rows...)
	msg := tgbotapi.NewMessage(chatID, "Введите свой город на русском или английском языке")
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func Сontains(priority_city [3]string, city string) bool {
	for _, ex_city := range priority_city {
		if ex_city == city {
			return true
		}
	}
	return false
}
