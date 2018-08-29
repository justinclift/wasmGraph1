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
	rectX  = 20
	rectY  = 60
	step   = 20
	debug  = true // If true, some debugging info is printed to the javascript console
)

func main() {
	// Initialise canvas
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
	width = doc.Get("body").Get("clientWidth").Float()
	height = doc.Get("body").Get("clientHeight").Float()
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	canvasEl.Set("tabIndex", 0) // Not sure if this is needed
	ctx := canvasEl.Call("getContext", "2d")

	// Simple keyboard handler for catching the arrow, WASD, and numpad keys
	// Key value info can be found here: https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/key/Key_Values
	keypressEvt := js.NewCallback(func(args []js.Value) {
		event := args[0]
		key := event.Get("key").String()
		if debug {
			fmt.Printf("Key is: %v\n", key)
		}
		switch key {
		case "ArrowLeft":
			rectX -= step
		case "ArrowRight":
			rectX += step
		case "ArrowUp":
			rectY -= step
		case "ArrowDown":
			rectY += step
		case "a":
			rectX -= step
		case "d":
			rectX += step
		case "w":
			rectY -= step
		case "s":
			rectY += step
		case "A":
			rectX -= step
		case "D":
			rectX += step
		case "W":
			rectY -= step
		case "S":
			rectY += step
		case "4":
			rectX -= step
		case "6":
			rectX += step
		case "8":
			rectY -= step
		case "2":
			rectY += step
		case "7":
			rectX -= step
			rectY -= step
		case "9":
			rectX += step
			rectY -= step
		case "1":
			rectX -= step
			rectY += step
		case "3":
			rectX += step
			rectY += step
		}
	})
	defer keypressEvt.Release()
	doc.Call("addEventListener", "keydown", keypressEvt)

	done := make(chan struct{}, 0)

	var renderFrame js.Callback
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

		// Write a line about using the keyboard keys
		ctx.Set("fillStyle", "blue")
		ctx.Set("font", "24px serif")
		ctx.Call("fillText", "Use the wasd, arrow, or numpad keys to move the square around.", 20, 40)

		// Draw a simple square
		ctx.Call("beginPath")
		ctx.Set("strokeStyle", "black")
		ctx.Set("lineWidth", "2")
		ctx.Call("moveTo", rectX, rectY)
		ctx.Call("lineTo", rectX+rectWidth, rectY)
		ctx.Call("lineTo", rectX+rectWidth, rectY+rectHeight)
		ctx.Call("lineTo", rectX, rectY+rectHeight)
		ctx.Call("closePath")
		ctx.Call("stroke")

		// TODO: Whatever else

		// It seems kind of weird (to me) to recursively call itself here, instead of using a timer approach, but
		// apparently this is best practise (at least in web environments: https://css-tricks.com/using-requestanimationframe)
		js.Global().Call("requestAnimationFrame", renderFrame)
	})
	defer renderFrame.Release()

	// Start running
	js.Global().Call("requestAnimationFrame", renderFrame)

	// Keeps the application running
	<-done
}
