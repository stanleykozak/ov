package oviewer

import (
	"log"
	"strings"
)

// Go to the top line.
func (root *Root) moveTop() {
	root.moveLine(0)
}

// Go to the bottom line.
func (root *Root) moveBottom() {
	root.moveLine(root.Doc.endNum + 1)
}

// Move to the specified line.
func (root *Root) moveLine(lN int) {
	root.resetSelect()
	root.Doc.topLN = lN
	root.Doc.topLX = 0
}

// Move up one screen.
func (root *Root) movePgUp() {
	root.resetSelect()
	root.moveNumUp(root.statusPos - root.headerLen())
}

// Moves down one screen.
func (root *Root) movePgDn() {
	root.resetSelect()
	root.Doc.topLN = root.bottomLN - root.Doc.Header
	root.Doc.topLX = root.bottomLX
}

// Moves up half a screen.
func (root *Root) moveHfUp() {
	root.resetSelect()
	root.moveNumUp((root.statusPos - root.headerLen()) / 2)
}

// Moves down half a screen.
func (root *Root) moveHfDn() {
	root.resetSelect()
	root.moveNumDown((root.statusPos - root.headerLen()) / 2)
}

// numOfSlice returns what number x is in slice.
func numOfSlice(listX []int, x int) int {
	for n, v := range listX {
		if v >= x {
			return n
		}
	}
	return len(listX) - 1
}

// numOfReverseSlice returns what number x is from the back of slice.
func numOfReverseSlice(listX []int, x int) int {
	for n := len(listX) - 1; n >= 0; n-- {
		if listX[n] <= x {
			return n
		}
	}
	return 0
}

// Moves up by the specified number of y.
func (root *Root) moveNumUp(moveY int) {
	if !root.Doc.WrapMode {
		root.Doc.topLN -= moveY
		return
	}

	// WrapMode
	num := root.Doc.topLN + root.Doc.Header
	root.Doc.topLX, num = root.findNumUp(root.Doc.topLX, num, moveY)
	root.Doc.topLN = num - root.Doc.Header
}

// Moves down by the specified number of y.
func (root *Root) moveNumDown(moveY int) {
	if !root.Doc.WrapMode {
		root.Doc.topLN += moveY
		return
	}

	// WrapMode
	num := root.Doc.topLN + root.Doc.Header
	x := root.Doc.topLX

	listX, err := root.leftMostX(num)
	if err != nil {
		log.Println(err, num)
		return
	}
	n := numOfReverseSlice(listX, x)

	for y := 0; y < moveY; y++ {
		if n >= len(listX) {
			num++
			if num > root.Doc.endNum {
				break
			}
			listX, err = root.leftMostX(num)
			if err != nil {
				log.Println(err, num)
				return
			}
			n = 0
		}
		x = 0
		if len(listX) > 0 && n < len(listX) {
			x = listX[n]
		}
		n++
	}
	root.Doc.topLN = num - root.Doc.Header
	root.Doc.topLX = x
}

// Move up one line.
func (root *Root) moveUp() {
	root.resetSelect()

	if root.Doc.topLN == 0 && root.Doc.topLX == 0 {
		return
	}

	if !root.Doc.WrapMode {
		root.Doc.topLN--
		root.Doc.topLX = 0
		return
	}

	// WrapMode.
	// Same line.
	if root.Doc.topLX > 0 {
		listX, err := root.leftMostX(root.Doc.topLN + root.Doc.Header)
		if err != nil {
			log.Println(err)
			return
		}
		for n, x := range listX {
			if x >= root.Doc.topLX {
				root.Doc.topLX = listX[n-1]
				return
			}
		}
	}

	// Previous line.
	root.Doc.topLN--
	if root.Doc.topLN < 0 {
		root.Doc.topLN = 0
		root.Doc.topLX = 0
		return
	}
	listX, err := root.leftMostX(root.Doc.topLN + root.Doc.Header)
	if err != nil {
		log.Println(err)
		return
	}
	if len(listX) > 0 {
		root.Doc.topLX = listX[len(listX)-1]
		return
	}
	root.Doc.topLX = 0
}

// Move down one line.
func (root *Root) moveDown() {
	root.resetSelect()

	if !root.Doc.WrapMode {
		root.Doc.topLX = 0
		root.Doc.topLN++
		return
	}

	// WrapMode
	listX, err := root.leftMostX(root.Doc.topLN + root.Doc.Header)
	if err != nil {
		log.Println(err)
		return
	}
	for _, x := range listX {
		if x > root.Doc.topLX {
			root.Doc.topLX = x
			return
		}
	}

	// Next line.
	root.Doc.topLX = 0
	root.Doc.topLN++
}

// Move to the left.
func (root *Root) moveLeft() {
	root.resetSelect()
	if root.Doc.ColumnMode {
		if root.Doc.columnNum > 0 {
			root.Doc.columnNum--
			root.Doc.x = root.columnModeX()
		}
		return
	}
	if root.Doc.WrapMode {
		return
	}
	root.Doc.x--
	if root.Doc.x < root.minStartX {
		root.Doc.x = root.minStartX
	}
}

// Move to the right.
func (root *Root) moveRight() {
	root.resetSelect()
	if root.Doc.ColumnMode {
		root.Doc.columnNum++
		root.Doc.x = root.columnModeX()
		return
	}
	if root.Doc.WrapMode {
		return
	}
	root.Doc.x++
}

// columnModeX returns the actual x from root.Doc.columnNum.
func (root *Root) columnModeX() int {
	// root.Doc.Header+10 = Maximum columnMode target.
	for i := 0; i < root.Doc.Header+10; i++ {
		lc, err := root.Doc.lineToContents(root.Doc.topLN+root.Doc.Header+i, root.Doc.TabWidth)
		if err != nil {
			continue
		}
		lineStr, byteMap := contentsToStr(lc)
		// Skip lines that do not contain a delimiter.
		if !strings.Contains(lineStr, root.Doc.ColumnDelimiter) {
			continue
		}

		start, end := rangePosition(lineStr, root.Doc.ColumnDelimiter, root.Doc.columnNum)
		if start < 0 || end < 0 {
			root.Doc.columnNum = 0
			start, _ = rangePosition(lineStr, root.Doc.ColumnDelimiter, root.Doc.columnNum)
		}
		return byteMap[start]
	}
	return 0
}

// Move to the left by half a screen.
func (root *Root) moveHfLeft() {
	if root.Doc.WrapMode {
		return
	}
	root.resetSelect()
	moveSize := (root.vWidth / 2)
	if root.Doc.x > 0 && (root.Doc.x-moveSize) < 0 {
		root.Doc.x = 0
	} else {
		root.Doc.x -= moveSize
		if root.Doc.x < root.minStartX {
			root.Doc.x = root.minStartX
		}
	}
}

// Move to the right by half a screen.
func (root *Root) moveHfRight() {
	if root.Doc.WrapMode {
		return
	}
	root.resetSelect()
	if root.Doc.x < 0 {
		root.Doc.x = 0
	} else {
		root.Doc.x += (root.vWidth / 2)
	}
}
