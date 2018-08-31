//Wasming
// compile: GOOS=js GOARCH=wasm go build -o main.wasm ./main.go
package main

import (
	"fmt"
	"math"
	"syscall/js"
	"time"

	"go.uber.org/atomic"
)

type matrix []float64

type Point struct {
	Num int
	X   float64
	Y   float64
	Z   float64
}

type Edge []int
type Surface []int

type Object struct {
	C   string // Colour of the object
	P   []Point
	E   []Edge    // List of points to connect by edges
	S   []Surface // List of points to connect in order, to create a surface
	Mid Point     // The mid point of the object.  Used for calculating object draw order in a very simple way
}

type OperationType int

const (
	ROTATE OperationType = iota
	SCALE
	TRANSLATE
)

type Operation struct {
	op OperationType
	t  int32 // Number of milliseconds the operation should take
	f  int32 // Number of display frames the operation should be broken into
	X  float64
	Y  float64
	Z  float64
}

type paintOrder struct {
	midZ float64 // Z depth of an object's mid point
	name string
}

var (
	// The empty world space
	worldSpace   map[string]Object
	pointCounter = int(1)

	// The point objects
	object1 = Object{
		C: "lightblue",
		P: []Point{
			{X: 0, Y: 1.75, Z: 1.0},    // Point 0 for this object
			{X: 1.5, Y: -1.75, Z: 1.0}, // Point 1 for this object
			{X: -1.5, Y: -1.75, Z: 1.0},
			{X: 0, Y: 0, Z: 1.75},
		},
		E: []Edge{
			{0, 1}, // Connect point 0 to point 1
			{0, 2}, // Connect point 0 to point 2
			{1, 2}, // Connect point 1 to point 2
			{0, 3}, // etc
			{1, 3},
			{2, 3},
		},
		S: []Surface{
			{0, 1, 3},
			{0, 2, 3},
			{0, 1, 2},
			{1, 2, 3},
		},
	}
	object2 = Object{
		C: "lightgreen",
		P: []Point{
			{X: 1.5, Y: 1.5, Z: -1.0},  // Point 0 for this object
			{X: 1.5, Y: -1.5, Z: -1.0}, // Point 1 for this object
			{X: -1.5, Y: -1.5, Z: -1.0},
		},
		E: []Edge{
			{0, 1}, // Connect point 0 to point 1
			{1, 2}, // Connect point 1 to point 2
			{2, 0}, // etc
		},
		S: []Surface{
			{0, 1, 2},
		},
	}
	object3 = Object{
		C: "indianred",
		P: []Point{
			{X: 2, Y: -2, Z: 1.0},
			{X: 2, Y: -4, Z: 1.0},
			{X: -2, Y: -4, Z: 1.0},
			{X: -2, Y: -2, Z: 1.0},
			{X: 0, Y: -3, Z: 2.5},
		},
		E: []Edge{
			{0, 1},
			{1, 2},
			{2, 3},
			{3, 0},
			{0, 4},
			{1, 4},
			{2, 4},
			{3, 4},
		},
		S: []Surface{
			{0, 1, 4},
			{1, 2, 4},
			{2, 3, 4},
			{3, 0, 4},
			{0, 1, 2, 3},
		},
	}

	// The 4x4 identity matrix
	identityMatrix = matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}

	// Initialise the transform matrix with the identity matrix
	transformMatrix = identityMatrix

	// FIFO queue
	queue        chan Operation
	renderActive *atomic.Bool

	width, height      float64
	kCall, rCall       js.Callback
	ctx, doc, canvasEl js.Value
	opText             string
	debug              = false // If true, some debugging info is printed to the javascript console
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

	// Set up the keypress handler
	renderActive = atomic.NewBool(false)
	kCall = js.NewCallback(keypressHandler)
	doc.Call("addEventListener", "keydown", kCall)
	defer kCall.Release()

	// Set the frame renderer going
	rCall = js.NewCallback(renderFrame)
	js.Global().Call("requestAnimationFrame", rCall)
	defer rCall.Release()

	// Set the operations processor going
	queue = make(chan Operation)
	go processOperations(queue)

	// TODO: Look into clip regions, so things outside the graph area aren't drawn
	//       This probably means we have to draw the point info table on the right differently too

	// Add some objects to the world space
	worldSpace = make(map[string]Object, 1)
	worldSpace["ob1"] = importObject(object1, 3.0, 3.0, 0.0)
	worldSpace["ob1 copy"] = importObject(object1, -3.0, 3.0, 0.0)
	worldSpace["ob2"] = importObject(object2, 3.0, -3.0, 1.0)
	worldSpace["ob3"] = importObject(object3, -3.0, 0.0, -1.0)

	// Add some transformation operations to the queue
	queue <- Operation{op: ROTATE, t: 1000, f: 60, X: 0, Y: 0, Z: 90}
	queue <- Operation{op: SCALE, t: 1000, f: 60, X: 2.0, Y: 2.0, Z: 2.0}
	queue <- Operation{op: ROTATE, t: 1000, f: 60, X: 0, Y: 360, Z: 0}
	queue <- Operation{op: SCALE, t: 1000, f: 60, X: 0.5, Y: 0.5, Z: 0.5}
	queue <- Operation{op: ROTATE, t: 1000, f: 60, X: 45, Y: 0, Z: -240}
	queue <- Operation{op: SCALE, t: 1000, f: 60, X: 1.5, Y: 1.5, Z: 1.52}

	// Keep the application running
	done := make(chan struct{}, 0)
	<-done
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

	// Translate the points
	var midX, midY, midZ float64
	var pt Point
	for _, j := range ob.P {
		pt = Point{
			Num: pointCounter,
			X:   (translateMatrix[0] * j.X) + (translateMatrix[1] * j.Y) + (translateMatrix[2] * j.Z) + (translateMatrix[3] * 1),   // 1st col, top
			Y:   (translateMatrix[4] * j.X) + (translateMatrix[5] * j.Y) + (translateMatrix[6] * j.Z) + (translateMatrix[7] * 1),   // 1st col, upper middle
			Z:   (translateMatrix[8] * j.X) + (translateMatrix[9] * j.Y) + (translateMatrix[10] * j.Z) + (translateMatrix[11] * 1), // 1st col, lower middle
		}
		translatedObject.P = append(translatedObject.P, pt)
		midX = pt.X
		midY = pt.Y
		midZ = pt.Z
		pointCounter++
	}

	// Determine the mid point for the object
	numPts := float64(len(ob.P))
	translatedObject.Mid.X = midX / numPts
	translatedObject.Mid.Y = midY / numPts
	translatedObject.Mid.Z = midZ / numPts

	// Copy the colour, edge, and surface definitions across
	translatedObject.C = ob.C
	for _, j := range ob.E {
		translatedObject.E = append(translatedObject.E, j)
	}
	for _, j := range ob.S {
		translatedObject.S = append(translatedObject.S, j)
	}

	return translatedObject
}

// Simple keyboard handler for catching the arrow, WASD, and numpad keys
// Key value info can be found here: https://developer.mozilla.org/en-US/docs/Web/API/KeyboardEvent/key/Key_Values
// TODO: See if it's feasible to catch mouse wheel events.  If so, add scale operations for enabling zoom in/out
func keypressHandler(args []js.Value) {
	event := args[0]
	key := event.Get("key").String()
	if debug {
		fmt.Printf("Key is: %v\n", key)
	}

	// Don't add operations if one is already in progress
	stepSize := float64(25)
	if !renderActive.Load() {
		switch key {
		case "ArrowLeft", "a", "A", "4":
			queue <- Operation{op: ROTATE, t: 50, f: 12, X: 0, Y: -stepSize, Z: 0}
		case "ArrowRight", "d", "D", "6":
			queue <- Operation{op: ROTATE, t: 50, f: 12, X: 0, Y: stepSize, Z: 0}
		case "ArrowUp", "w", "W", "8":
			queue <- Operation{op: ROTATE, t: 50, f: 12, X: -stepSize, Y: 0, Z: 0}
		case "ArrowDown", "s", "S", "2":
			queue <- Operation{op: ROTATE, t: 50, f: 12, X: stepSize, Y: 0, Z: 0}
		case "7", "Home":
			queue <- Operation{op: ROTATE, t: 50, f: 12, X: -stepSize, Y: -stepSize, Z: 0}
		case "9", "PageUp":
			queue <- Operation{op: ROTATE, t: 50, f: 12, X: -stepSize, Y: stepSize, Z: 0}
		case "1", "End":
			queue <- Operation{op: ROTATE, t: 50, f: 12, X: stepSize, Y: -stepSize, Z: 0}
		case "3", "PageDown":
			queue <- Operation{op: ROTATE, t: 50, f: 12, X: stepSize, Y: stepSize, Z: 0}
		case "-":
			queue <- Operation{op: ROTATE, t: 50, f: 12, X: 0, Y: 0, Z: -stepSize}
		case "+":
			queue <- Operation{op: ROTATE, t: 50, f: 12, X: 0, Y: 0, Z: stepSize}
		}
	}
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

// Animates the transformation operations
func processOperations(queue <-chan Operation) {
	for i := range queue {
		renderActive.Store(true)         // Mark rendering as now in progress
		parts := i.f                     // Number of parts to break each transformation into
		transformMatrix = identityMatrix // Reset the transform matrix
		switch i.op {
		case ROTATE: // Rotate the objects in world space
			// Divide the desired angle into a small number of parts
			if i.X != 0 {
				transformMatrix = rotateAroundX(transformMatrix, i.X/float64(parts))
			}
			if i.Y != 0 {
				transformMatrix = rotateAroundY(transformMatrix, i.Y/float64(parts))
			}
			if i.Z != 0 {
				transformMatrix = rotateAroundZ(transformMatrix, i.Z/float64(parts))
			}
			opText = fmt.Sprintf("Rotation. X: %0.2f Y: %0.2f Z: %0.2f", i.X, i.Y, i.Z)

		case SCALE:
			// Scale the objects in world space
			var xPart, yPart, zPart float64
			if i.X != 1 {
				xPart = ((i.X - 1) / float64(parts)) + 1
			}
			if i.Y != 1 {
				yPart = ((i.Y - 1) / float64(parts)) + 1
			}
			if i.Z != 1 {
				zPart = ((i.Z - 1) / float64(parts)) + 1
			}
			transformMatrix = scale(transformMatrix, xPart, yPart, zPart)
			opText = fmt.Sprintf("Scale. X: %0.2f Y: %0.2f Z: %0.2f", i.X, i.Y, i.Z)

		case TRANSLATE:
			// Translate (move) the objects in world space
			transformMatrix = translate(transformMatrix, i.X/float64(parts), i.Y/float64(parts), i.Z/float64(parts))
			opText = fmt.Sprintf("Translate (move). X: %0.2f Y: %0.2f Z: %0.2f", i.X, i.Y, i.Z)
		}

		// Apply each transformation, one small part at a time (this gives the animation effect)
		timeSlice := time.Millisecond * time.Duration(i.t/parts)
		for t := 0; t < int(parts); t++ {
			time.Sleep(timeSlice)
			for j, o := range worldSpace {
				var newPoints []Point

				// Transform each point of in the object
				for _, j := range o.P {
					newPoints = append(newPoints, transform(transformMatrix, j))
				}
				o.P = newPoints

				// Transform the mid point of the object.  In theory, this should mean the mid point can always be used
				// for a simple (not-cpu-intensive) way to sort the objects in Z depth order
				o.Mid = transform(transformMatrix, o.Mid)

				// Update the object in world space
				worldSpace[j] = o
			}
		}
		renderActive.Store(false)
		opText = "Complete. Rotate with WASD or numpad."
	}
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

		// Setup useful variables
		border := float64(2)
		gap := float64(3)
		left := border + gap
		top := border + gap
		graphWidth := width * 0.75
		graphHeight := height - 1
		centerX := graphWidth / 2
		centerY := graphHeight / 2

		// Clear the background
		ctx.Set("fillStyle", "white")
		ctx.Call("fillRect", 0, 0, width, height)

		// Draw grid lines
		step := math.Min(width, height) / 30
		ctx.Set("strokeStyle", "rgb(220, 220, 220)")
		ctx.Call("setLineDash", []interface{}{1, 3})
		for i := left; i < graphWidth-step; i += step {
			// Vertical dashed lines
			ctx.Call("beginPath")
			ctx.Call("moveTo", i+step, top)
			ctx.Call("lineTo", i+step, graphHeight)
			ctx.Call("stroke")
		}
		for i := top; i < graphHeight-step; i += step {
			// Horizontal dashed lines
			ctx.Call("beginPath")
			ctx.Call("moveTo", left, i+step)
			ctx.Call("lineTo", graphWidth-border, i+step)
			ctx.Call("stroke")
		}

		// Sort the objects by mid point Z depth order
		order := make(map[int]paintOrder, len(worldSpace))
		var tmpOrder paintOrder
		for i, j := range worldSpace {
			if len(order) == 0 {
				// Add the first order item
				order[0] = paintOrder{name: i, midZ: j.Mid.Z}
			} else {
				tmpOrder = paintOrder{name: i, midZ: j.Mid.Z}
				for k := 0; k < len(order); k++ {
					if order[k].midZ > tmpOrder.midZ {
						// Swap the items
						a := paintOrder{name: order[k].name, midZ: order[k].midZ}
						order[k] = tmpOrder
						tmpOrder = a
					}
				}
				order[len(order)] = tmpOrder // Add the new item (should be the largest Z value so far) to the end
			}
		}

		// Draw the objects, in Z depth order
		var pointX, pointY float64
		var pointNum int
		numWld := len(worldSpace)
		for i := 0; i < numWld; i++ {
			o := worldSpace[order[i].name]

			// Draw the surfaces
			ctx.Set("fillStyle", o.C)
			for _, l := range o.S {
				for m, n := range l {
					pointX = o.P[n].X
					pointY = o.P[n].Y
					if m == 0 {
						ctx.Call("beginPath")
						ctx.Call("moveTo", centerX+(pointX*step), centerY+((pointY*step)*-1))
					} else {
						ctx.Call("lineTo", centerX+(pointX*step), centerY+((pointY*step)*-1))
					}
				}
				ctx.Call("closePath")
				ctx.Call("fill")
			}

			// Draw the edges
			ctx.Set("strokeStyle", "black")
			ctx.Set("fillStyle", "black")
			ctx.Set("lineWidth", "1")
			ctx.Call("setLineDash", []interface{}{2, 4})
			var point1X, point1Y, point2X, point2Y float64
			for _, l := range o.E {
				point1X = o.P[l[0]].X
				point1Y = o.P[l[0]].Y
				point2X = o.P[l[1]].X
				point2Y = o.P[l[1]].Y
				ctx.Call("beginPath")
				ctx.Call("moveTo", centerX+(point1X*step), centerY+((point1Y*step)*-1))
				ctx.Call("lineTo", centerX+(point2X*step), centerY+((point2Y*step)*-1))
				ctx.Call("stroke")
			}

			// Draw the points on the graph
			ctx.Call("setLineDash", []interface{}{})
			var px, py float64
			for _, l := range o.P {
				// Draw a dot for the point
				px = centerX + (l.X * step)
				py = centerY + ((l.Y * step) * -1)
				ctx.Call("beginPath")
				ctx.Call("arc", px, py, 1, 0, 2*math.Pi)
				ctx.Call("fill")

				// Label the point on the graph
				ctx.Set("font", "12px sans-serif")
				ctx.Call("fillText", fmt.Sprintf("Point %d", l.Num), px+5, py+15)
			}
		}

		// Clear the information area (right side)
		ctx.Set("fillStyle", "white")
		ctx.Call("fillRect", graphWidth+1, 0, width, height)

		// Draw the information area (right side) text
		ctx.Set("fillStyle", "black")
		for _, o := range worldSpace {
			for _, l := range o.P {
				// Draw darker coloured legend text
				ctx.Set("font", "bold 14px serif")
				ctx.Call("fillText", fmt.Sprintf("Point %d:", l.Num), graphWidth+20, l.Num*25)

				// Draw lighter coloured legend text
				ctx.Set("font", "12px sans-serif")
				ctx.Call("fillText", fmt.Sprintf("(%0.1f, %0.1f, %0.1f)", l.X, l.Y, l.Z), graphWidth+100, l.Num*25)
				pointNum++
			}
		}

		// Draw a border around the graph area
		ctx.Call("setLineDash", []interface{}{})
		ctx.Set("lineWidth", "2")
		ctx.Set("strokeStyle", "white")
		ctx.Call("beginPath")
		ctx.Call("moveTo", 0, 0)
		ctx.Call("lineTo", width, 0)
		ctx.Call("lineTo", width, height)
		ctx.Call("lineTo", 0, height)
		ctx.Call("closePath")
		ctx.Call("stroke")
		ctx.Set("lineWidth", "2")
		ctx.Set("strokeStyle", "black")
		ctx.Call("beginPath")
		ctx.Call("moveTo", border, border)
		ctx.Call("lineTo", graphWidth, border)
		ctx.Call("lineTo", graphWidth, graphHeight)
		ctx.Call("lineTo", border, graphHeight)
		ctx.Call("closePath")
		ctx.Call("stroke")

		// Draw the text describing the current operation
		// TODO: Figure out better Y placement for this
		ctx.Set("font", "bold 14px serif")
		ctx.Call("fillText", "Operation:", graphWidth+20, graphHeight-40)
		ctx.Set("font", "14px sans-serif")
		ctx.Call("fillText", opText, graphWidth+20, graphHeight-20)

		// Schedule the next frame render call
		js.Global().Call("requestAnimationFrame", rCall)
	}
}

// Rotates a transformation matrix around the X axis by the given degrees
func rotateAroundX(m matrix, degrees float64) matrix {
	rad := (math.Pi / 180) * degrees // The Go math functions use radians, so we convert degrees to radians
	rotateXMatrix := matrix{
		1, 0, 0, 0,
		0, math.Cos(rad), -math.Sin(rad), 0,
		0, math.Sin(rad), math.Cos(rad), 0,
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

	t.Num = p.Num
	t.X = (top0 * p.X) + (top1 * p.Y) + (top2 * p.Z) + top3
	t.Y = (upperMid0 * p.X) + (upperMid1 * p.Y) + (upperMid2 * p.Z) + upperMid3
	t.Z = (lowerMid0 * p.X) + (lowerMid1 * p.Y) + (lowerMid2 * p.Z) + lowerMid3
	return
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
