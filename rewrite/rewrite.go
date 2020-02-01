package rewrite

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// 发送文字
func SendMsg(bot *tgbotapi.BotAPI, chatID int64, msgID int, content string) {
	var msg tgbotapi.MessageConfig
	msg = tgbotapi.NewMessage(chatID, content)
	if msgID != -1 {
		msg.ReplyToMessageID = msgID
	}
	if _, err := bot.Send(msg); err != nil {
		// log.Panic(err)
	}
}

// 发送图片
func SendImg(bot *tgbotapi.BotAPI, chatID int64, img interface{}) {
	upload := tgbotapi.NewPhotoUpload(chatID, img)
	if _, err := bot.Send(upload); err != nil {
		// log.Panic(err)
	}
}

// 删除信息
func SendDel(bot *tgbotapi.BotAPI, chatID int64, msgID int) {
	del := tgbotapi.NewDeleteMessage(chatID, msgID)
	if _, err := bot.Send(del); err != nil {
		// log.Panic(err)
	}
}
