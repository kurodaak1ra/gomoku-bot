package play

import (
	"board"
	"redis"
	"regexp"
	"rewrite"
	"strconv"
	"strings"
	"tools"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	uuid "github.com/satori/go.uuid"
)

// 创建游戏
func Create(bot *tgbotapi.BotAPI, chatID int64, msgID int, creatorID int, creatorName string) {
	wait := strconv.FormatInt(chatID, 10) + ":waiting:" + strconv.Itoa(creatorID)
	if ex, _ := redis.EXISTS(wait); ex == int64(1) {
		rewrite.SendMsg(bot, chatID, msgID, "您已创建游戏，不能重复创建")
		return
	}
	if arr, _ := redis.KEYS(strconv.FormatInt(chatID, 10) + ":*" + strconv.Itoa(creatorID) + "*"); len(arr) != 0 {
		rewrite.SendMsg(bot, chatID, msgID, "您加入的游戏尚未结束，不能创建新的游戏")
		return
	}
	redis.SET([][]string{
		[]string{wait, creatorName, "300"},
	})
	// 送信
	rewrite.SendMsg(bot, chatID, msgID, "游戏创建成功，等待玩家加入游戏(5分钟内有效)")
}

// 匹配对手
func JoinPrev(bot *tgbotapi.BotAPI, chatID int64, msgID int, joinedID int, joinedName string) {
	// 游戏中不可加入其他局
	d, _ := redis.KEYS(strconv.FormatInt(chatID, 10) + ":battle-*" + strconv.Itoa(joinedID) + "*")
	if len(d) != 0 {
		rewrite.SendMsg(bot, chatID, msgID, "您已在游戏中，不能加入其他场次")
		return
	}
	// 从数据库中读取等待区名单
	s, _ := redis.KEYS(strconv.FormatInt(chatID, 10) + ":waiting:*")
	if len(s) == 0 {
		rewrite.SendMsg(bot, chatID, msgID, "无人创建游戏，请使用 /create 命令创建游戏")
		return
	}
	msg := tgbotapi.NewMessage(chatID, "请从下面等待玩家中选择一个开始游戏(20秒内有效)")
	var list tgbotapi.InlineKeyboardMarkup
	for _, val := range s {
		username, _ := redis.GET(val)
		path := strconv.FormatInt(chatID, 10) + ":waiting:"
		// UUID
		uuid := uuid.Must(uuid.NewV4()).String()
		// 数据入库
		redis.SET([][]string{
			[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":from", "join", "20"},
			[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":ownerID", strconv.Itoa(joinedID), "20"},
			[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":uid", strings.Split(val, path)[1], "20"},
			[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid + ":username", username.(string), "20"},
		})
		// 填充选项列表
		list.InlineKeyboard = append(list.InlineKeyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(username.(string), uuid),
		))
	}
	msg.ReplyMarkup = list
	msg.ReplyToMessageID = msgID
	bot.Send(msg)
}

// 加入游戏
func JoinNext(bot *tgbotapi.BotAPI, chatID int64, msgID int, creatorID, creatorName, joinedID, joinedName string) {
	redis.DEL([]string{
		strconv.FormatInt(chatID, 10) + ":waiting:" + creatorID, // 清除等待中
	})
	key := strconv.FormatInt(chatID, 10) + ":battle-" + creatorID + "-" + joinedID
	// 生成数据棋盘
	for x := 0; x < 19; x++ {
		for y := 0; y < 19; y++ {
			redis.LPUSH(key+":chessboard:"+strconv.Itoa(x), "8", 900)
		}
	}
	// 设置 玩家棋子颜色 及 当前棋子标志
	redis.SET([][]string{
		[]string{key + ":isBlack", "true", "900"},
		[]string{key + ":" + creatorID + ":color", "black", "900"},
		[]string{key + ":" + creatorID + ":username", creatorName, "900"},
		[]string{key + ":" + joinedID + ":color", "white", "900"},
		[]string{key + ":" + joinedID + ":username", joinedName, "900"},
	})
	// 拼接文件名
	filename := creatorID + "-" + joinedID
	// 创建空棋盘
	board.DrawBoard(0, 0, true, creatorName, joinedName, strconv.FormatInt(chatID, 10), filename)
	// 发送空棋盘
	rewrite.SendMsg(bot, chatID, msgID, "加入成功，游戏正式开始\n庄家持黑棋，先出子\n落子格式，例如：/n H5")
	rewrite.SendImg(bot, chatID, "./static/img/"+strconv.FormatInt(chatID, 10)+"/"+filename+".png")
}

// 落子
func Next(bot *tgbotapi.BotAPI, chatID int64, msgID int, UID int, coordinate string) {
	// 非空棋盘逻辑判断、绘制开始
	if coordinate == "" {
		rewrite.SendMsg(bot, chatID, msgID, "输入格式错误（例如：/n H5）")
		return
	}
	regNUM := regexp.MustCompile(`^(?:1[0-9]|[1-9])$`) // 正则匹配 数字
	regABC := regexp.MustCompile(`[a-sA-S]+`)          // 正则匹配 大小写字母
	// 根据 uid 查询当前对局数据
	d, _ := redis.KEYS(strconv.FormatInt(chatID, 10) + ":battle-*" + strconv.Itoa(UID) + "*")
	if len(d) == 0 {
		rewrite.SendMsg(bot, chatID, msgID, "您尚未加入任何一局游戏，或您所在的棋局由于超过15分钟无操作，已被销毁")
		return
	}
	dataPathArr := strings.Split(d[0], ":")           // path 切割
	uidDirSplit := strings.Split(dataPathArr[1], "-") // 两个人 UID 切割
	cName, _ := redis.GET(dataPathArr[0] + ":" + dataPathArr[1] + ":" + uidDirSplit[1] + ":username")
	jName, _ := redis.GET(dataPathArr[0] + ":" + dataPathArr[1] + ":" + uidDirSplit[2] + ":username")
	userChessColor, _ := redis.GET(dataPathArr[0] + ":" + dataPathArr[1] + ":" + strconv.Itoa(UID) + ":color")
	blackFlagStr, _ := redis.GET(dataPathArr[0] + ":" + dataPathArr[1] + ":isBlack")
	isBlack, _ := strconv.ParseBool(blackFlagStr.(string))
	var isBlackBin int
	var currentChessColor string
	if isBlack {
		isBlackBin = 0
		currentChessColor = "black"
	} else {
		isBlackBin = 1
		currentChessColor = "white"
	}
	// 判断输入数据有效性
	if !regABC.MatchString(coordinate[0:1]) || !regNUM.MatchString(coordinate[1:len(coordinate)]) {
		rewrite.SendMsg(bot, chatID, msgID, "输入格式错误（例如：/n H5），或坐标已超出棋盘范围")
		return
	}
	insertX, insertY := tools.CoordinateEx(coordinate)
	// 检测到谁落子
	if userChessColor != currentChessColor {
		rewrite.SendMsg(bot, chatID, msgID, "当前为对手落子")
		return
	}
	// 检测落子位置是否有效
	res, _ := redis.LRANGE(dataPathArr[0]+":"+dataPathArr[1]+":chessboard:"+strconv.Itoa(insertX), insertY, insertY)
	if res[0] != "8" {
		rewrite.SendMsg(bot, chatID, msgID, "落子坐标无效")
		return
	}
	// 重置 TTL
	var gameDataPath [][]string
	for x := 0; x < 19; x++ {
		for y := 0; y < 19; y++ {
			gameDataPath = append(gameDataPath, []string{dataPathArr[0] + ":" + dataPathArr[1] + ":chessboard:" + strconv.Itoa(x), "900"})
		}
	}
	gameDataPath = append(gameDataPath, []string{dataPathArr[0] + ":" + dataPathArr[1] + ":" + uidDirSplit[1] + ":color", "900"})
	gameDataPath = append(gameDataPath, []string{dataPathArr[0] + ":" + dataPathArr[1] + ":" + uidDirSplit[1] + ":username", "900"})
	gameDataPath = append(gameDataPath, []string{dataPathArr[0] + ":" + dataPathArr[1] + ":" + uidDirSplit[2] + ":color", "900"})
	gameDataPath = append(gameDataPath, []string{dataPathArr[0] + ":" + dataPathArr[1] + ":" + uidDirSplit[2] + ":username", "900"})
	redis.EXPIRE(gameDataPath)
	// 数据入库（redis）
	redis.LPUSH(dataPathArr[0]+":"+dataPathArr[1]+":"+strconv.Itoa(UID)+":step", coordinate, 900)
	redis.LSET(dataPathArr[0]+":"+dataPathArr[1]+":chessboard:"+strconv.Itoa(insertX), insertY, isBlackBin)
	redis.SET([][]string{
		[]string{dataPathArr[0] + ":" + dataPathArr[1] + ":isBlack", strconv.FormatBool(!isBlack), "900"},
	})
	// 绘制棋盘
	board.DrawBoard(insertX, insertY, !isBlack, cName.(string), jName.(string), strconv.FormatInt(chatID, 10), uidDirSplit[1]+"-"+uidDirSplit[2])
	// 发送新棋盘
	rewrite.SendImg(bot, chatID, "./static/img/"+strconv.FormatInt(chatID, 10)+"/"+uidDirSplit[1]+"-"+uidDirSplit[2]+".png")
	// 判定
	if winner := collectArr(19, insertX, insertY, dataPathArr[0]+":"+dataPathArr[1]+":chessboard:", bot, chatID, msgID); winner == "" {
		// rewrite.SendDel(bot, chatID, msgID) // 删除玩家发送的坐标消息
	} else {
		creatorTheWorld, _ := redis.LLEN(dataPathArr[0] + ":" + dataPathArr[1] + ":" + uidDirSplit[1] + ":theWorld")
		joinorTheWorld, _ := redis.LLEN(dataPathArr[0] + ":" + dataPathArr[1] + ":" + uidDirSplit[2] + ":theWorld")
		if winner == "black" {
			rewrite.SendMsg(bot, chatID, -1, "@"+cName.(string)+" @"+jName.(string)+"\n本局游戏结束，黑棋获胜！\n黑棋："+cName.(string)+" 悔棋"+strconv.FormatInt(creatorTheWorld.(int64), 10)+"次\n白棋："+jName.(string)+" 悔棋"+strconv.FormatInt(joinorTheWorld.(int64), 10)+"次")
		} else if winner == "white" {
			rewrite.SendMsg(bot, chatID, -1, "@"+cName.(string)+" @"+jName.(string)+"\n本局游戏结束，白棋获胜！\n黑棋："+cName.(string)+" 悔棋"+strconv.FormatInt(creatorTheWorld.(int64), 10)+"次\n白棋："+jName.(string)+" 悔棋"+strconv.FormatInt(joinorTheWorld.(int64), 10)+"次")
		}
		redis.CLEAR(chatID, UID)
	}
}

// 悔棋
func TheWorld(bot *tgbotapi.BotAPI, chatID int64, msgID int, UID int, username string) {
	res, _ := redis.KEYS(strconv.FormatInt(chatID, 10) + ":battle-*" + strconv.Itoa(UID) + "*")
	if len(res) == 0 {
		rewrite.SendMsg(bot, chatID, msgID, "您尚未加入任何一局游戏，或您所在的棋局由于超过15分钟无人操作，已被销毁")
		return
	}
	dataPathArr := strings.Split(res[0], ":") // path 切割
	stepLen, _ := redis.LLEN(dataPathArr[0] + ":" + dataPathArr[1] + ":" + strconv.Itoa(UID) + ":step")
	if stepLen.(int64) < 2 {
		rewrite.SendMsg(bot, chatID, msgID, "两步以内禁止悔棋")
		return
	}
	uidDirSplit := strings.Split(dataPathArr[1], "-") // 两个人 UID 切割
	var anotherUID string
	if uid, _ := strconv.Atoi(uidDirSplit[1]); UID == uid {
		anotherUID = uidDirSplit[2]
	} else {
		anotherUID = uidDirSplit[1]
	}
	anotherName, _ := redis.GET(dataPathArr[0] + ":" + dataPathArr[1] + ":" + anotherUID + ":username")
	uuid1 := uuid.Must(uuid.NewV4()).String()
	uuid2 := uuid.Must(uuid.NewV4()).String()
	redis.SET([][]string{
		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid1 + ":from", "the-world", "10"},
		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid1 + ":ownerID", anotherUID, "10"},
		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid1 + ":result", "true", "10"},
		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid1 + ":path", dataPathArr[0] + ":" + dataPathArr[1], "10"},
		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid1 + ":applicantUID", strconv.Itoa(UID), "10"},
		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid1 + ":applicantUsername", username, "10"},

		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid2 + ":from", "the-world", "10"},
		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid2 + ":ownerID", anotherUID, "10"},
		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid2 + ":result", "false", "10"},
		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid2 + ":path", dataPathArr[0] + ":" + dataPathArr[1], "10"},
		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid2 + ":applicantUID", strconv.Itoa(UID), "10"},
		[]string{strconv.FormatInt(chatID, 10) + ":buttons:" + uuid2 + ":applicantUsername", username, "10"},
	})
	msg := tgbotapi.NewMessage(chatID, "@"+anotherName.(string)+" 对手触发了悔棋指令，是否允许(10秒内有效，超时默认不允许)")
	list := tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{
			[]tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData("いいよ〜", uuid1),
				tgbotapi.NewInlineKeyboardButtonData("だめです", uuid2),
			},
		},
	}
	msg.ReplyMarkup = list
	bot.Send(msg)
}

// 手动结束
func GameOver(bot *tgbotapi.BotAPI, chatID int64, msgID int, UID int) {
	if cb, _ := redis.CLEAR(chatID, UID); cb == nil {
		rewrite.SendMsg(bot, chatID, msgID, "您未加入任何一场游戏")
		return
	}
	rewrite.SendMsg(bot, chatID, msgID, "游戏结束，不计输赢")
}

// 收集 横竖左右 数组
func collectArr(c, insertX, insertY int, path string, bot *tgbotapi.BotAPI, chatID int64, msgID int) string {
	var (
		rowArr    []string
		columnArr []string
		leftArr   []string
		rightArr  []string
		startX    int = insertX - 4
		startY    int = insertY - 4
		endX      int = insertX + 4
		endY      int = insertY + 4
	)
	for x := startY; x <= endY; x++ {
		if x < 0 || x > 18 {
			continue
		}
		step := x - startY
		// 横
		if x == insertY {
			for y := startX; y <= endX; y++ {
				if y >= 0 && y <= 18 {
					row, _ := redis.LRANGE(path+strconv.Itoa(y), x, x)
					rowArr = append(rowArr, row[0])
				}
			}
		}
		// 竖
		column, _ := redis.LRANGE(path+strconv.Itoa(insertX), x, x)
		columnArr = append(columnArr, column[0])
		// 左
		if startX+step >= 0 && startX+step <= 18 {
			left, _ := redis.LRANGE(path+strconv.Itoa(startX+step), x, x)
			leftArr = append(leftArr, left[0])
		}
		// 右
		if endX-step >= 0 && endX-step <= 18 {
			right, _ := redis.LRANGE(path+strconv.Itoa(endX-step), x, x)
			rightArr = append(rightArr, right[0])
		}
	}
	return determinationPrev(rowArr, columnArr, leftArr, rightArr, bot, chatID, msgID)
}

// 判定
func determinationPrev(row, column, left, right []string, bot *tgbotapi.BotAPI, chatID int64, msgID int) string {

	// 横
	if res := determinationFinal(row); res == "black" {
		return "black"
	} else if res == "white" {
		return "white"
	}
	// 竖
	if res := determinationFinal(column); res == "black" {
		return "black"
	} else if res == "white" {
		return "white"
	}
	// 左
	if res := determinationFinal(left); res == "black" {
		return "black"

	} else if res == "white" {
		return "white"
	}
	// 右
	if res := determinationFinal(right); res == "black" {
		return "black"

	} else if res == "white" {
		return "white"
	}
	return ""
}

var (
	determinationBlack       int
	determinationWhite       int
	determinationTargetChess string
)

// 五子棋判定
func determinationFinal(arr []string) string {
	determinationBlack = 0
	determinationWhite = 0
	determinationTargetChess = ""
	for _, val := range arr {
		if val == "8" {
			continue
		}
		if val == "0" {
			determinationBlack += 1
		}
		if val == "1" {
			determinationWhite += 1
		}
		// fmt.Println(determinationBlack, determinationWhite)
		if determinationBlack == 5 {
			return "black"
		}
		if determinationWhite == 5 {
			return "white"
		}
		if determinationTargetChess == "" {
			determinationTargetChess = val
		} else {
			if val != determinationTargetChess {
				determinationBlack = 0
				determinationWhite = 0
				determinationTargetChess = val
			}
		}
	}
	return ""
}
