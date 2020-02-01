package board

import (
	"fmt"
	"os"
	"redis"
	"strconv"
	"tools"

	"github.com/fogleman/gg"
)

var (
	rate float64 = 0.5 // 画布尺寸调节比率
)

// 创建棋盘
func DrawBoard(insertX, insertY int, isBlack bool, creatorName, joinedName, groupDir, fileame string) {
	// 创建画布
	dc := gg.NewContext(int(1100*rate), int(1200*rate))
	// 棋盘底色
	dc.SetHexColor("f0b060")
	dc.DrawRectangle(0, 0, 1100*rate, 1200*rate)
	dc.Fill()
	dc.Stroke()
	// 绘制对局人员水印
	// dc.RotateAbout(-0.4363323129985824, 0, 1200*rate)
	// for x := 0; x < 48; x++ {
	// 	dc.SetRGBA(0, 0, 0, 0.2)
	// 	if err := dc.LoadFontFace("./static/font/SourceHanSans.ttf", 45*rate); err != nil {
	// 		panic(fmt.Sprintf("font file not found![%v]\n", err))
	// 	}
	// 	row := float64((x%4)*650) * rate
	// 	column := math.Floor(float64(x / 4))
	// 	columnSpace := column * 140 * rate
	// 	if int(column)%2 != 0 { // 水印隔行交叉
	// 		row -= 310 * rate
	// 	}
	// 	dc.DrawStringAnchored(creatorName+" - "+joinedName, row, columnSpace, 0.5, 0.5)
	// }
	// dc.Stroke()
	// 绘制棋盘
	// dc.RotateAbout(0.4363323129985824, 0, 1200*rate)
	lineSpacing := 50 * rate
	for x := 0; x < 19; x++ {
		dc.SetHexColor("81512f")
		dc.SetLineWidth(4 * rate)
		dc.DrawLine(float64(x*int(lineSpacing)+int(100*rate)), 100*rate, float64(x*int(lineSpacing)+int(100*rate)), 1000*rate) // 横线
		dc.DrawLine(100*rate, float64(x*int(lineSpacing)+int(100*rate)), 1000*rate, float64(x*int(lineSpacing)+int(100*rate))) // 竖线
		if err := dc.LoadFontFace("./static/font/SourceHanSans.ttf", 35*rate); err != nil {
			panic(fmt.Sprintf("font file not found![%v]\n", err))
		}
		dc.DrawStringAnchored(string(65+x), float64(50*(x+2))*rate, 50*rate, 0.5, 0.5)                     // 横轴坐标（上）
		dc.DrawStringAnchored(string(65+x), float64(50*(x+2))*rate, 1050*rate, 0.5, 0.5)                   // 横轴坐标（下）
		dc.DrawStringAnchored(string(fmt.Sprintf("%d", x+1)), 50*rate, float64(50*(x+2))*rate, 0.5, 0.5)   // 纵轴坐标（左）
		dc.DrawStringAnchored(string(fmt.Sprintf("%d", x+1)), 1050*rate, float64(50*(x+2))*rate, 0.5, 0.5) // 纵轴坐标（右）
	}
	dc.Stroke()
	// 绘制中心点及边角四点
	dc.DrawCircle(250*rate, 250*rate, 10*rate) // 左上
	dc.DrawCircle(250*rate, 850*rate, 10*rate) // 左下
	dc.DrawCircle(550*rate, 550*rate, 10*rate) // 中点
	dc.DrawCircle(850*rate, 250*rate, 10*rate) // 右上
	dc.DrawCircle(850*rate, 850*rate, 10*rate) // 右下
	dc.SetHexColor("81512f")
	dc.Fill()
	dc.Stroke()
	// 绘制棋子
	for x := 0; x < 19; x++ {
		res, _ := redis.LRANGE(groupDir+":battle-"+fileame+":chessboard:"+strconv.Itoa(x), 0, 18)
		for y := 0; y < len(res); y++ {
			if res[y] == "8" {
				continue
			}
			if res[y] == "0" {
				dc.SetHexColor("282828")
			} else if res[y] == "1" {
				dc.SetHexColor("f3f3f3")
			}
			dc.DrawCircle(float64((x+1)*50+50)*rate, float64((y+1)*50+50)*rate, 22*rate)
			dc.Fill()
			dc.Stroke()
			// 上次落子标记
			if x == insertX && y == insertY {
				dc.SetHexColor("f00")
				dc.DrawCircle(float64((x+1)*50+50)*rate, float64((y+1)*50+50)*rate, 12*rate)
				dc.Fill()
				dc.Stroke()
			}
		}
	}
	// 绘制分割线
	dc.SetHexColor("282828")
	dc.DrawLine(0, float64(1100*rate), 1100*rate, float64(1100*rate))
	dc.Stroke()
	// 绘制黑色说明棋子
	dc.SetHexColor("282828")
	dc.DrawCircle(60*rate, 1150*rate, 22*rate)
	dc.Fill()
	dc.DrawStringAnchored(creatorName, 110*rate, 1160*rate, 0, 0)
	dc.Stroke()
	// 绘制白色说明棋子
	dc.SetHexColor("f3f3f3")
	dc.DrawCircle(550*rate, 1150*rate, 22*rate)
	dc.Fill()
	dc.DrawStringAnchored(joinedName, 600*rate, 1160*rate, 0, 0)
	dc.Stroke()
	// 当前落子人标记
	dc.SetHexColor("f00")
	if isBlack {
		dc.DrawCircle(60*rate, 1150*rate, 12*rate)
	} else {
		dc.DrawCircle(550*rate, 1150*rate, 12*rate)
	}
	dc.Fill()
	dc.Stroke()
	// 输赢判定
	// if winner == "black" {
	// 	dc.SetRGB(186, 0, 31)
	// 	if err := dc.LoadFontFace("./static/font/SourceHanSans.ttf", 160*rate); err != nil {
	// 		panic(fmt.Sprintf("font file not found![%v]\n", err))
	// 	}
	// 	dc.DrawStringAnchored("黑棋获胜", 550*rate, 550*rate, 0.5, 0.5)
	// 	dc.Stroke()
	// } else if winner == "white" {
	// 	dc.SetRGB(186, 0, 31)
	// 	if err := dc.LoadFontFace("./static/font/SourceHanSans.ttf", 160*rate); err != nil {
	// 		panic(fmt.Sprintf("font file not found![%v]\n", err))
	// 	}
	// 	dc.DrawStringAnchored("白棋获胜", 550*rate, 550*rate, 0.5, 0.5)
	// 	dc.Stroke()
	// }
	// 检测文件夹是否需要被创建
	exist, err := tools.PathExists("./static/img/" + groupDir)
	if err != nil {
		panic(fmt.Sprintf("get dir error![%v]\n", err))
	}
	if !exist {
		err := os.Mkdir("./static/img/"+groupDir, os.ModePerm)
		if err != nil {
			panic(fmt.Sprintf("mkdir failed![%v]\n", err))
		}
	}
	// 导出图片
	dc.SavePNG("./static/img/" + groupDir + "/" + fileame + ".png")
}
