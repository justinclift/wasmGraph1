//Wasming
// compile: GOOS=js GOARCH=wasm go build -o main.wasm ./main.go
package main

import (
	"fmt"
	"syscall/js"
)

var (
	width  float64
	height float64
)

func main() {
	// Initialise canvas
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
	width = doc.Get("body").Get("clientWidth").Float()
	height = doc.Get("body").Get("clientHeight").Float()
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	ctx := canvasEl.Call("getContext", "2d")

	var offset int
	keypressEvt := js.NewCallback(func(args []js.Value) {
		e := args[0]
		//fmt.Printf("Stuff here: %v", args[0])
		println("Hello, WebAssembly!")
		key := e.Get("key").String()
		//fmt.Printf("Event is of type: %v\n", e.Type())
		//for i, j := range key {
		//	fmt.Printf("Index %v, Value %v\n", i, j)
		//}
		fmt.Printf("Key is: %v\n", key)
		//kbdType := kbd.InstanceOf("KeyboardObject")
		//if kbdType {
		//	println("Kbd is of type KeyboardObject")
		//}
		//fmt.Printf("stuff: %v\n", args[0])
		//println(e)
		offset+=1
	})
	defer keypressEvt.Release()
	doc.Call("addEventListener", "keydown", keypressEvt)


	done := make(chan struct{}, 0)

	var renderFrame js.Callback
	rectX := 20
	rectY := 20
	rectHeight := 30
	rectWidth := 30
	renderFrame = js.NewCallback(func(args []js.Value) {
		// Handle window resizing
		curBodyW := doc.Get("body").Get("clientWidth").Float()
		curBodyH := doc.Get("body").Get("clientHeight").Float()
		if curBodyW != width || curBodyH != height {
			width, height = curBodyW, curBodyH
			canvasEl.Set("width", width)
			canvasEl.Set("height", height)
		}

		// ** Draw the frame content **

		// Grey background
		ctx.Set("fillStyle", "lightgrey")
		ctx.Call("fillRect", 1, 1, width-1, height-1)

		// TODO: Other stuff

		// Draw a simple square
		ctx.Set("fillStyle", "black")
		ctx.Call("fillRect", offset + rectX, rectY, rectWidth, 2)
		ctx.Call("fillRect", offset + rectX, rectY, 2, rectHeight)
		ctx.Call("fillRect", offset + rectX+rectWidth, rectY, 2, rectHeight+2)
		ctx.Call("fillRect", offset + rectX, rectY+rectHeight, rectWidth, 2)


		// It seems kind of weird to recursively call itself here, instead of using a timer approach, but apparently
		// this is best practise (at least in web environments: https://css-tricks.com/using-requestanimationframe)
		js.Global().Call("requestAnimationFrame", renderFrame)
	})
	defer renderFrame.Release()

	// Start running
	js.Global().Call("requestAnimationFrame", renderFrame)

	// Keeps the application running
	<-done
}
