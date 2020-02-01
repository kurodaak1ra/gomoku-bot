package button

import (
	"play"
	"redis"
	"rewrite"
	"strconv"
	"tools"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// 选择对手 Button
func SelectBtn(uuid, callbackQueryID string, bot *tgbotapi.BotAPI, chatID int64, msgID, fromID int, fromName string) {
	uid, _ := redis.GET(strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":uid")
	username, _ := redis.GET(strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":username")
	// 获取等待池中的人是否存在
	waitingExists, _ := redis.EXISTS(strconv.FormatInt(chatID, 10) + ":waiting:" + uid.(string))
	if waitingExists == int64(0) {
		bot.AnswerCallbackQuery(tgbotapi.NewCallbackWithAlert(callbackQueryID, "此人已过期")) // 回调应答
		return
	}
	// 禁止自娱自乐
	if strconv.Itoa(fromID) == uid {
		bot.AnswerCallbackQuery(tgbotapi.NewCallbackWithAlert(callbackQueryID, "禁止自娱自乐")) // 回调应答
		return
	}
	// 通过
	rewrite.SendDel(bot, chatID, msgID)                                                                // 删除选项消息
	bot.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQueryID, ""))                                 // 回调应答
	play.JoinNext(bot, chatID, msgID, uid.(string), username.(string), strconv.Itoa(fromID), fromName) // 加入
}

// 悔棋 Button
func TheWorldBtn(uuid, callbackQueryID string, bot *tgbotapi.BotAPI, chatID int64, msgID int, UID int) {
	bot.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQueryID, ""))                                           // 回调应答
	applicantUsername, _ := redis.GET(strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":applicantUsername") // 悔棋人 username
	if val, _ := redis.GET(strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":result"); val == "false" {
		rewrite.SendMsg(bot, chatID, -1, "@"+applicantUsername.(string)+" 对方拒绝了您的悔棋请求，请继续落子")
		return
	}
	rewrite.SendMsg(bot, chatID, -1, "@"+applicantUsername.(string)+" 对方同意了您的悔棋请求，请继续落子")
	path, _ := redis.GET(strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":path")
	blackFlagStr, _ := redis.GET(path.(string) + ":isBlack")
	isBlack, _ := strconv.ParseBool(blackFlagStr.(string))
	applicantUID, _ := redis.GET(strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":applicantUID") // 悔棋人 UID
	applicantColor, _ := redis.GET(path.(string) + ":" + applicantUID.(string) + ":color")             // 悔棋人棋子颜色
	var currentColor string
	if isBlack {
		currentColor = "black"
	} else {
		currentColor = "white"
	}
	redis.LPUSH(path.(string)+":"+strconv.Itoa(UID)+":theWorld", "1", 300)
	if currentColor == applicantColor {
		anotherPopRes, _ := redis.LPOP(path.(string) + ":" + strconv.Itoa(UID) + ":step")
		furuiX, furuiY := tools.CoordinateEx(anotherPopRes.(string))
		redis.LSET(path.(string)+":chessboard:"+strconv.Itoa(furuiX), furuiY, 8)
		redis.SET([][]string{
			[]string{path.(string) + ":isBlack", strconv.FormatBool(!isBlack), "900"},
		})
	}
	applicantPopRes, _ := redis.LPOP(path.(string) + ":" + applicantUID.(string) + ":step") // 悔棋人 最后一次 落子位置
	furuiX, furuiY := tools.CoordinateEx(applicantPopRes.(string))
	redis.LSET(path.(string)+":chessboard:"+strconv.Itoa(furuiX), furuiY, 8)
	anotherSecPopRes, _ := redis.LPOP(path.(string) + ":" + strconv.Itoa(UID) + ":step") // 对手 最后第一次 落子位置
	prevX, prevY := tools.CoordinateEx(anotherSecPopRes.(string))
	redis.LSET(path.(string)+":chessboard:"+strconv.Itoa(prevX), prevY, 8)
	play.Next(bot, chatID, msgID, UID, anotherSecPopRes.(string))
}
