package main

import (
	"button"
	"log"
	"play"
	"redis"
	"regexp"
	"rewrite"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	chatID   int64
	msgID    int
	fromID   int
	fromName string
)

func main() {
	bot, err := tgbotapi.NewBotAPI("token") // token
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		// 监听 callback，匹配对手
		if update.CallbackQuery != nil {
			chatID = update.CallbackQuery.Message.Chat.ID
			msgID = update.CallbackQuery.Message.MessageID
			fromID = update.CallbackQuery.From.ID
			fromName = update.CallbackQuery.From.UserName
			uuid := update.CallbackQuery.Data
			// 根据来源分发
			fromPath, _ := redis.GET(strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":from")
			if fromPath == nil {
				bot.AnswerCallbackQuery(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "选项已过期")) // 回调应答
				continue
			}
			if ownerID, _ := redis.GET(strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":ownerID"); strconv.Itoa(fromID) != ownerID {
				bot.AnswerCallbackQuery(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "雨女无瓜")) // 回调应答
				continue
			}
			if fromPath == "join" {
				button.SelectBtn(uuid, update.CallbackQuery.ID, bot, chatID, msgID, fromID, fromName)
			} else if fromPath == "the-world" {
				button.TheWorldBtn(uuid, update.CallbackQuery.ID, bot, chatID, msgID, fromID)
			}
		}

		// 没消息 continue
		if update.Message == nil {
			continue
		}

		// 变量赋值
		chatID = update.Message.Chat.ID
		msgID = update.Message.MessageID
		fromID = update.Message.From.ID

		// 指令切割
		reg := regexp.MustCompile(`\s+`)
		splitStrArr := reg.Split(strings.TrimSpace(update.Message.Text), -1)
		var ins [2]string
		ins[0] = splitStrArr[0]
		if len(splitStrArr) != 1 {
			ins[1] = splitStrArr[1]
		}

		// 匹配指令分发
		if ins[0] == "/start@ka_dev_bot" || ins[0] == "/start@ka_wzq_game_bot" { // 欢迎页
			rewrite.SendMsg(bot, chatID, msgID, "这是一个后端基于 Go 语言编写的五子棋游戏 Bot")
		} else if ins[0] == "/create@ka_dev_bot" || ins[0] == "/create@ka_wzq_game_bot" { // 创建
			play.Create(bot, chatID, msgID, fromID, update.Message.From.UserName)
		} else if ins[0] == "/join@ka_dev_bot" || ins[0] == "/join@ka_wzq_game_bot" { // 加入 (prev
			play.JoinPrev(bot, chatID, msgID, fromID, update.Message.From.UserName)
		} else if ins[0] == "/n" { // 下棋
			play.Next(bot, chatID, msgID, fromID, ins[1])
		} else if ins[0] == "/game_over@ka_dev_bot" || ins[0] == "/game_over@ka_wzq_game_bot" { // 手动结束
			play.GameOver(bot, chatID, msgID, fromID)
		} else if ins[0] == "/the_world@ka_dev_bot" || ins[0] == "/the_world@ka_wzq_game_bot" { // 悔棋
			play.TheWorld(bot, chatID, msgID, fromID, fromName)
		}
	}
}
