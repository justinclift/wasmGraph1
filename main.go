//Wasming
// compile: GOOS=js GOARCH=wasm go build -o main.wasm ./main.go
package main

import (
	"fmt"
	"math"
	"math/rand"
	"syscall/js"
)

type matrix []float64

type Point struct {
	Colour [3]int
	Num    int
	X      float64
	Y      float64
	Z      float64
}

type Object []Point

var (
	// The empty world space
	worldSpace   map[string]Object
	pointCounter = int(1)

	// The point objects
	object1 = Object{
		{X: 0, Y: 1.75, Z: 1.0},
		{X: 1.5, Y: -1.75, Z: 1.0},
		{X: -1.5, Y: -1.75, Z: 1.0},
		{X: 0, Y: 0, Z: 1.75},
	}
	object2 = Object{
		{X: 1.5, Y: 1.5, Z: -1.0},
		{X: 1.5, Y: -1.5, Z: -1.0},
		{X: -1.5, Y: -1.5, Z: -1.0},
	}
	object3 = Object{
		{X: 2, Y: -2, Z: 1.0},
		{X: 2, Y: -4, Z: 1.0},
		{X: -2, Y: -4, Z: 1.0},
		{X: -2, Y: -2, Z: 1.0},
		{X: 0, Y: -3, Z: 2.5},
	}

	width, height      float64
	rCall              js.Callback
	ctx, doc, canvasEl js.Value
	//debug  = true // If true, some debugging info is printed to the javascript console
)

func main() {
	// Initialise canvas
	doc = js.Global().Get("document")
	canvasEl = doc.Call("getElementById", "mycanvas")
	width = doc.Get("body").Get("clientWidth").Float()
	height = doc.Get("body").Get("clientHeight").Float()
	canvasEl.Call("setAttribute", "width", width)
	canvasEl.Call("setAttribute", "height", height)
	canvasEl.Set("tabIndex", 0) // Not sure if this is needed
	ctx = canvasEl.Call("getContext", "2d")

	// Seed the random generator (only used for colour generation), with a value that generates a "known ok" colour set
	rand.Seed(3)

	// Add some objects to the world space
	worldSpace = make(map[string]Object, 1)
	worldSpace["ob1"] = importObject(object1, 3.0, 3.0, 0.0)
	worldSpace["ob1 copy"] = importObject(object1, -3.0, 3.0, 0.0)
	worldSpace["ob2"] = importObject(object2, 3.0, -3.0, 1.0)
	worldSpace["ob3"] = importObject(object3, -3.0, 0.0, -1.0)

	//// Simple keyboard handler for catching the arrow, WASD, and numpad keys
	//// Key value info can be found here: https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/key/Key_Values
	//keypressEvt := js.NewCallback(func(args []js.Value) {
	//	event := args[0]
	//	key := event.Get("key").String()
	//	if debug {
	//		fmt.Printf("Key is: %v\n", key)
	//	}
	//	// TODO: Use key presses to rotate the view around the world space origin
	//	//switch key {
	//	//case "ArrowLeft", "a", "A", "4":
	//	//	rectX -= step
	//	//case "ArrowRight", "d", "D", "6":
	//	//	rectX += step
	//	//case "ArrowUp", "w", "W", "8":
	//	//	rectY -= step
	//	//case "ArrowDown", "s", "S", "2":
	//	//	rectY += step
	//	//case "7", "Home":
	//	//	rectX -= step
	//	//	rectY -= step
	//	//case "9", "PageUp":
	//	//	rectX += step
	//	//	rectY -= step
	//	//case "1", "End":
	//	//	rectX -= step
	//	//	rectY += step
	//	//case "3", "PageDown":
	//	//	rectX += step
	//	//	rectY += step
	//	//}
	//})
	//defer keypressEvt.Release()
	//doc.Call("addEventListener", "keydown", keypressEvt)

	// Set the frame renderer going
	rCall = js.NewCallback(renderFrame)
	js.Global().Call("requestAnimationFrame", rCall)
	defer rCall.Release()

	// Keeps the application running
	done := make(chan struct{}, 0)
	<-done
}

// Renders one frame of the animation
func renderFrame(args []js.Value) {
	{
		// Handle window resizing
		curBodyW := doc.Get("body").Get("clientWidth").Float()
		curBodyH := doc.Get("body").Get("clientHeight").Float()
		if curBodyW != width || curBodyH != height {
			width, height = curBodyW, curBodyH
			canvasEl.Set("width", width)
			canvasEl.Set("height", height)
		}

		// ** Draw the graph area **

		// Grey background
		ctx.Set("fillStyle", "lightgrey")
		ctx.Call("fillRect", 1, 1, width-1, height-1)

		// Draw border around the graph area
		border := float64(2)
		gap := float64(3)
		left := border + gap
		top := border + gap
		graphWidth := width * 0.75
		graphHeight := height - 1
		centerX := graphWidth / 2
		centerY := graphHeight / 2
		ctx.Call("beginPath")
		ctx.Set("strokeStyle", "black")
		ctx.Set("lineWidth", "2")
		ctx.Call("moveTo", border, border)
		ctx.Call("lineTo", graphWidth, border)
		ctx.Call("lineTo", graphWidth, graphHeight)
		ctx.Call("lineTo", border, graphHeight)
		ctx.Call("closePath")
		ctx.Call("stroke")

		// Draw horizontal axis
		ctx.Call("beginPath")
		ctx.Set("strokeStyle", "black")
		ctx.Set("lineWidth", "2")
		ctx.Call("moveTo", left, centerY)
		ctx.Call("lineTo", graphWidth-gap, centerY)
		ctx.Call("stroke")

		// Draw vertical axis
		ctx.Call("beginPath")
		ctx.Set("strokeStyle", "black")
		ctx.Set("lineWidth", "2")
		ctx.Call("moveTo", centerX, top)
		ctx.Call("lineTo", centerX, graphHeight-gap)
		ctx.Call("stroke")

		// Draw horizontal markers
		step := math.Min(width, height) / 30
		markerPoint := graphHeight/2 - 5
		for i := graphWidth / 2; i < graphWidth-step; i += step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", i+step, markerPoint)
			ctx.Call("lineTo", i+step, markerPoint+10)
			ctx.Call("stroke")
		}
		for i := graphWidth / 2; i > left; i -= step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", i-step, markerPoint)
			ctx.Call("lineTo", i-step, markerPoint+10)
			ctx.Call("stroke")
		}

		// Draw vertical markers
		markerPoint = graphWidth/2 - 5
		for i := graphHeight / 2; i < graphHeight-step; i += step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", markerPoint, i+step)
			ctx.Call("lineTo", markerPoint+10, i+step)
			ctx.Call("stroke")
		}
		for i := graphHeight / 2; i > top; i -= step {
			ctx.Call("beginPath")
			ctx.Set("strokeStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("moveTo", markerPoint, i-step)
			ctx.Call("lineTo", markerPoint+10, i-step)
			ctx.Call("stroke")
		}

		// TODO: Maybe add axis label text every few points?

		// Draw the points
		var pointNum int
		for _, o := range worldSpace {
			for k, l := range o {

				// Draw the coloured dot for the point
				px := centerX + (l.X * step)
				py := centerY + ((l.Y * step) * -1)
				ctx.Set("fillStyle", fmt.Sprintf("rgb(%d, %d, %d)", l.Colour[0], l.Colour[1], l.Colour[2]))
				ctx.Call("beginPath")
				ctx.Call("arc", px, py, 5, 0, 2*math.Pi)
				ctx.Call("fill")
				ctx.Set("fillStyle", "black")
				ctx.Set("font", "12px sans-serif")
				ctx.Call("fillText", fmt.Sprintf("Point %d", l.Num), px+5, py+15)

				// Draw lines between the points
				var tx, ty float64
				if k == 0 {
					if len(o) >= 4 {
						// TODO: This is just a dodgy workaround while testing.  Would be good to figure out a better approach.
						//       Maybe instead of using just Points, use some kind of structure that also defines edges?  Or would
						//       code to automatically work out the edges between points be better instead?
						tx = centerX + (o[len(o)-2].X * step)
						ty = centerY + ((o[len(o)-2].Y * step) * -1)
						ctx.Call("beginPath")
						ctx.Set("strokeStyle", "black")
						ctx.Set("lineWidth", "1")
						ctx.Call("moveTo", px, py)
						ctx.Call("lineTo", tx, ty)
						ctx.Call("stroke")

						px := centerX + (o[1].X * step)
						py := centerY + ((o[1].Y * step) * -1)
						tx = centerX + (o[len(o)-1].X * step)
						ty = centerY + ((o[len(o)-1].Y * step) * -1)
						ctx.Call("beginPath")
						ctx.Set("strokeStyle", "black")
						ctx.Set("lineWidth", "1")
						ctx.Call("moveTo", px, py)
						ctx.Call("lineTo", tx, ty)
						ctx.Call("stroke")
					}
					if len(o) == 5 {
						px := centerX + (o[2].X * step)
						py := centerY + ((o[2].Y * step) * -1)
						tx = centerX + (o[len(o)-1].X * step)
						ty = centerY + ((o[len(o)-1].Y * step) * -1)
						ctx.Call("beginPath")
						ctx.Set("strokeStyle", "black")
						ctx.Set("lineWidth", "1")
						ctx.Call("moveTo", px, py)
						ctx.Call("lineTo", tx, ty)
						ctx.Call("stroke")
					}
					tx = centerX + (o[len(o)-1].X * step)
					ty = centerY + ((o[len(o)-1].Y * step) * -1)
				} else {
					tx = centerX + (o[k-1].X * step)
					ty = centerY + ((o[k-1].Y * step) * -1)
				}
				ctx.Call("beginPath")
				ctx.Set("strokeStyle", "black")
				ctx.Set("lineWidth", "1")
				// setLineDash doesn't seem to work.  TODO: Figure out the right way to call this, as ctx.Call() outright fails
				// https://developer.mozilla.org/en-US/docs/Web/API/CanvasRenderingContext2D/setLineDash
				//ctx.Set("setLineDash", "{5, 15}")
				//ctx.Set("setLineDash", "5, 15")
				//ctx.Set("setLineDash", "[5, 15]")
				//ctx.Set("setLineDash", "\"[5, 15]\"")
				ctx.Call("moveTo", px, py)
				ctx.Call("lineTo", tx, ty)
				ctx.Call("stroke")

				// Draw darker coloured legend text
				ctx.Set("fillStyle", "black")
				ctx.Set("font", "bold 14px serif")
				ctx.Call("fillText", fmt.Sprintf("Point %d:", l.Num), graphWidth+20, l.Num*25)

				// Draw lighter coloured legend text
				ctx.Set("font", "12px sans-serif")
				ctx.Call("fillText", fmt.Sprintf("(%0.1f, %0.1f, %0.1f)", l.X, l.Y, l.Z), graphWidth+100, l.Num*25)
				pointNum++
			}
		}

		// It seems kind of weird (to me) to recursively call itself here, instead of using a timer approach, but
		// apparently this is best practise (at least in web environments: https://css-tricks.com/using-requestanimationframe)
		js.Global().Call("requestAnimationFrame", rCall)
	}
}

// Returns an object who's points have been transformed into 3D world space XYZ co-ordinates.  Also assigns numbers
// and colours to each point
func importObject(ob Object, x float64, y float64, z float64) (translatedObject Object) {
	// X and Y translation matrix.  Translates the objects into the world space at the given X and Y co-ordinates
	translateMatrix := matrix{
		1, 0, 0, x,
		0, 1, 0, y,
		0, 0, 1, z,
		0, 0, 0, 1,
	}

	for _, j := range ob {
		translatedObject = append(translatedObject, Point{
			Colour: [3]int{rand.Intn(255), rand.Intn(255), rand.Intn(255)},
			Num:    pointCounter,
			X:      (translateMatrix[0] * j.X) + (translateMatrix[1] * j.Y) + (translateMatrix[2] * j.Z) + (translateMatrix[3] * 1),   // 1st col, top
			Y:      (translateMatrix[4] * j.X) + (translateMatrix[5] * j.Y) + (translateMatrix[6] * j.Z) + (translateMatrix[7] * 1),   // 1st col, upper middle
			Z:      (translateMatrix[8] * j.X) + (translateMatrix[9] * j.Y) + (translateMatrix[10] * j.Z) + (translateMatrix[11] * 1), // 1st col, lower middle
		})
		pointCounter++
	}
	return translatedObject
}

// Multiplies one matrix by another
func matrixMult(opMatrix matrix, m matrix) (resultMatrix matrix) {
	top0 := m[0]
	top1 := m[1]
	top2 := m[2]
	top3 := m[3]
	upperMid0 := m[4]
	upperMid1 := m[5]
	upperMid2 := m[6]
	upperMid3 := m[7]
	lowerMid0 := m[8]
	lowerMid1 := m[9]
	lowerMid2 := m[10]
	lowerMid3 := m[11]
	bot0 := m[12]
	bot1 := m[13]
	bot2 := m[14]
	bot3 := m[15]

	resultMatrix = matrix{
		(opMatrix[0] * top0) + (opMatrix[1] * upperMid0) + (opMatrix[2] * lowerMid0) + (opMatrix[3] * bot0), // 1st col, top
		(opMatrix[0] * top1) + (opMatrix[1] * upperMid1) + (opMatrix[2] * lowerMid1) + (opMatrix[3] * bot1), // 2nd col, top
		(opMatrix[0] * top2) + (opMatrix[1] * upperMid2) + (opMatrix[2] * lowerMid2) + (opMatrix[3] * bot2), // 3rd col, top
		(opMatrix[0] * top3) + (opMatrix[1] * upperMid3) + (opMatrix[2] * lowerMid3) + (opMatrix[3] * bot3), // 4th col, top

		(opMatrix[4] * top0) + (opMatrix[5] * upperMid0) + (opMatrix[6] * lowerMid0) + (opMatrix[7] * bot0), // 1st col, upper middle
		(opMatrix[4] * top1) + (opMatrix[5] * upperMid1) + (opMatrix[6] * lowerMid1) + (opMatrix[7] * bot1), // 2nd col, upper middle
		(opMatrix[4] * top2) + (opMatrix[5] * upperMid2) + (opMatrix[6] * lowerMid2) + (opMatrix[7] * bot2), // 3rd col, upper middle
		(opMatrix[4] * top3) + (opMatrix[5] * upperMid3) + (opMatrix[6] * lowerMid3) + (opMatrix[7] * bot3), // 4th col, upper middle

		(opMatrix[8] * top0) + (opMatrix[9] * upperMid0) + (opMatrix[10] * lowerMid0) + (opMatrix[11] * bot0), // 1st col, lower middle
		(opMatrix[8] * top1) + (opMatrix[9] * upperMid1) + (opMatrix[10] * lowerMid1) + (opMatrix[11] * bot1), // 2nd col, lower middle
		(opMatrix[8] * top2) + (opMatrix[9] * upperMid2) + (opMatrix[10] * lowerMid2) + (opMatrix[11] * bot2), // 3rd col, lower middle
		(opMatrix[8] * top3) + (opMatrix[9] * upperMid3) + (opMatrix[10] * lowerMid3) + (opMatrix[11] * bot3), // 4th col, lower middle

		(opMatrix[12] * top0) + (opMatrix[13] * upperMid0) + (opMatrix[14] * lowerMid0) + (opMatrix[15] * bot0), // 1st col, bottom
		(opMatrix[12] * top1) + (opMatrix[13] * upperMid1) + (opMatrix[14] * lowerMid1) + (opMatrix[15] * bot1), // 2nd col, bottom
		(opMatrix[12] * top2) + (opMatrix[13] * upperMid2) + (opMatrix[14] * lowerMid2) + (opMatrix[15] * bot2), // 3rd col, bottom
		(opMatrix[12] * top3) + (opMatrix[13] * upperMid3) + (opMatrix[14] * lowerMid3) + (opMatrix[15] * bot3), // 4th col, bottom
	}
	return resultMatrix
}

// Rotates a transformation matrix around the X axis by the given degrees
func rotateAroundX(m matrix, degrees float64) matrix {
	rad := (math.Pi / 180) * degrees // The Go math functions use radians, so we convert degrees to radians
	rotateXMatrix := matrix{
		1, 0, 0, 0,
		math.Cos(rad), 0, -math.Sin(rad), 0,
		math.Sin(rad), 0, math.Cos(rad), 0,
		0, 0, 0, 1,
	}
	return matrixMult(rotateXMatrix, m)
}

// Rotates a transformation matrix around the Y axis by the given degrees
func rotateAroundY(m matrix, degrees float64) matrix {
	rad := (math.Pi / 180) * degrees // The Go math functions use radians, so we convert degrees to radians
	rotateYMatrix := matrix{
		math.Cos(rad), 0, math.Sin(rad), 0,
		0, 1, 0, 0,
		-math.Sin(rad), 0, math.Cos(rad), 0,
		0, 0, 0, 1,
	}
	return matrixMult(rotateYMatrix, m)
}

// Rotates a transformation matrix around the Z axis by the given degrees
func rotateAroundZ(m matrix, degrees float64) matrix {
	rad := (math.Pi / 180) * degrees // The Go math functions use radians, so we convert degrees to radians
	rotateZMatrix := matrix{
		math.Cos(rad), -math.Sin(rad), 0, 0,
		math.Sin(rad), math.Cos(rad), 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	return matrixMult(rotateZMatrix, m)
}

// Scales a transformation matrix by the given X, Y, and Z values
func scale(m matrix, x float64, y float64, z float64) matrix {
	scaleMatrix := matrix{
		x, 0, 0, 0,
		0, y, 0, 0,
		0, 0, z, 0,
		0, 0, 0, 1,
	}
	return matrixMult(scaleMatrix, m)
}

// Translates (moves) a transformation matrix by the given X, Y and Z values
func translate(m matrix, translateX float64, translateY float64, translateZ float64) matrix {
	translateMatrix := matrix{
		1, 0, 0, translateX,
		0, 1, 0, translateY,
		0, 0, 1, translateZ,
		0, 0, 0, 1,
	}
	return matrixMult(translateMatrix, m)
}

// Transform the XYZ co-ordinates using the values from the transformation matrix
func transform(m matrix, p Point) (t Point) {
	top0 := m[0]
	top1 := m[1]
	top2 := m[2]
	top3 := m[3]
	upperMid0 := m[4]
	upperMid1 := m[5]
	upperMid2 := m[6]
	upperMid3 := m[7]
	lowerMid0 := m[8]
	lowerMid1 := m[9]
	lowerMid2 := m[10]
	lowerMid3 := m[11]
	//bot0 := m[12] // The fourth row values can be ignored for 3D matrices
	//bot1 := m[13]
	//bot2 := m[14]
	//bot3 := m[15]

	t.Colour = p.Colour
	t.Num = p.Num
	t.X = (top0 * p.X) + (top1 * p.Y) + (top2 * p.Z) + top3
	t.Y = (upperMid0 * p.X) + (upperMid1 * p.Y) + (upperMid2 * p.Z) + upperMid3
	t.Z = (lowerMid0 * p.X) + (lowerMid1 * p.Y) + (lowerMid2 * p.Z) + lowerMid3
	return
}
