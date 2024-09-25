// src/main.go
package main

import (
	"math"
	"syscall/js"
)

func processImage(this js.Value, args []js.Value) interface{} {
	// 检查是否有足够的参数
	if len(args) < 3 {
		js.Global().Get("console").Call("error", "processImage requires 3 arguments: data, width, height")
		return js.Null()
	}

	// 获取参数
	data := args[0]

	// 将 js.Value 类型的 data 转换为 []uint8
	byteLength := data.Get("byteLength").Int()
	byteSlice := make([]uint8, byteLength)
	js.CopyBytesToGo(byteSlice, data)

	// 简单的图像处理：将图像转换为灰度
	for i := 0; i < byteLength; i += 4 {
		r := float64(byteSlice[i])
		g := float64(byteSlice[i+1])
		b := float64(byteSlice[i+2])

		// 使用标准亮度公式计算灰度值
		gray := byte(math.Round(0.299*r + 0.587*g + 0.114*b))

		// 设置 RGB 为灰度值，保持 Alpha 不变
		byteSlice[i] = gray
		byteSlice[i+1] = gray
		byteSlice[i+2] = gray
		// byteSlice[i+3] 保持不变
	}

	// 将处理后的字节切片复制回 JavaScript 的 Uint8ClampedArray
	processedData := js.Global().Get("Uint8ClampedArray").New(byteLength)
	js.CopyBytesToJS(processedData, byteSlice)

	return processedData
}

func main() {
	c := make(chan struct{}, 0)

	js.Global().Set("processImage", js.FuncOf(processImage))

	<-c
}
