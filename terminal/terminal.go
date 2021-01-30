package terminal

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

// Screen buffer
// Its not recommended write to buffer dirrectly, use package Print,Printf,Println functions instead.
var Screen *bytes.Buffer = new(bytes.Buffer)

// Ouput buffer
var Output *bufio.Writer = bufio.NewWriter(os.Stdout)

// MoveCursor move cursor to a given x,y position
func MoveCursor(x int, y int) {
	fmt.Fprintf(Screen, "\033[%d;%dH", y, x)
}

// Print using Fprint
func Print(a ...interface{}) (n int, err error) {
	return fmt.Fprint(Screen, a...)
}

// Println using Fprintln
func Println(a ...interface{}) (n int, err error) {
	return fmt.Fprintln(Screen, a...)
}

// Printf uisng Fprintf
func Printf(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(Screen, format, a...)
}

// Height gets console height
func Height() int {
	ws, err := getWinsize()
	if err != nil {
		return -1
	}
	return int(ws.Row)
}

// Width gets console width
func Width() int {
	ws, err := getWinsize()

	if err != nil {
		return -1
	}

	return int(ws.Col)
}

// Clear screen
func Clear() {
	Output.WriteString("\033[2J")
}

// Flush buffer and ensure that it will not overflow screen
func Flush() {
	for idx, str := range strings.SplitAfter(Screen.String(), "\n") {
		if idx > Height() {
			return
		}

		Output.WriteString(str)
	}

	Output.Flush()
	Screen.Reset()
}
