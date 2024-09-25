// src/main.go
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"syscall/js"
)

func processImage(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		fmt.Println("未提供文件")
		return nil
	}

	file := args[0]

	// 获取 Promise 构造函数
	promiseConstructor := js.Global().Get("Promise")

	// 创建一个新的 Promise
	promise := promiseConstructor.New(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		resolve := args[0]
		reject := args[1]

		// 调用 JavaScript 的 arrayBuffer 方法，返回一个 Promise
		arrayBufferPromise := file.Call("arrayBuffer")
		arrayBufferPromise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			arrayBuffer := args[0]
			uint8Array := js.Global().Get("Uint8Array").New(arrayBuffer)
			length := uint8Array.Length()
			byteSlice := make([]byte, length)
			js.CopyBytesToGo(byteSlice, uint8Array)

			// 解码图片
			img, _, err := image.Decode(bytes.NewReader(byteSlice))
			if err != nil {
				fmt.Println("解码图片失败:", err)
				reject.Invoke(js.ValueOf(err.Error()))
				return nil
			}

			// 将图片转换为灰度图
			grayImg := image.NewGray(img.Bounds())
			for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
				for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
					originalColor := img.At(x, y)
					grayColor := color.GrayModel.Convert(originalColor)
					grayImg.Set(x, y, grayColor)
				}
			}

			// 编码处理后的图片为 PNG
			var buf bytes.Buffer
			if err := png.Encode(&buf, grayImg); err != nil {
				fmt.Println("编码图片失败:", err)
				reject.Invoke(js.ValueOf(err.Error()))
				return nil
			}

			// 将处理后的字节数据复制回 JavaScript 的 Uint8Array
			resultUint8Array := js.Global().Get("Uint8Array").New(len(buf.Bytes()))
			js.CopyBytesToJS(resultUint8Array, buf.Bytes())

			// 解析成功，调用 resolve
			resolve.Invoke(resultUint8Array)
			return nil
		}))
		return nil
	}))

	return promise
}

func main() {
	// 注册 processImage 函数到 JavaScript 的全局对象
	js.Global().Set("processImage", js.FuncOf(processImage))

	// 防止 Go 程序退出
	c := make(chan struct{}, 0)
	<-c
}
