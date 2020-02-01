package tools

import (
	"os"
	"strconv"
	"strings"
)

// uint8 to string
// func U2S(num interface{}) string {
// 	return string(num.([]uint8))
// }

// 检测文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 坐标转换
func CoordinateEx(coordinate string) (int, int) {
	x := int([]rune(strings.ToUpper(coordinate[0:1]))[0] - 65)
	y, _ := strconv.Atoi(coordinate[1:len(coordinate)])
	y = y - 1
	return x, y
}
