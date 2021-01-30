package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	tm "github.com/brnuts/matrix/terminal"
	sh "github.com/codeskyblue/go-sh"
	tb "github.com/nsf/termbox-go"
)

const message = "Windows size is %dx%d"
const wait = time.Second / 100000

const trueColors = true

var matrixChars = [...]int{
	// Kanji Table
	// Hiragana Chars
	0x3050, 0x3051, 0x3052, 0x3053, 0x3054, 0x3055, 0x3056, 0x3057, 0x3058, 0x3059, 0x3060,
	0x3061, 0x3062, 0x3063, 0x3064, 0x3065, 0x3066, 0x3067, 0x3068, 0x3069, 0x3070, 0x3071,
	// Katakana Chars
	0x30a1, 0x30a2, 0x30a3, 0x30a4, 0x30a5, 0x30a6, 0x30a7, 0x30a8, 0x30a9, 0x30aa, 0x30ab,
	0x30ac, 0x30ad, 0x30ae, 0x30af, 0x30b0, 0x30b1, 0x30b2, 0x30b3, 0x30b4, 0x30b5, 0x30b6,
	0x30b7, 0x30b8, 0x30b9, 0x30ba, 0x30bb, 0x30bc, 0x30bd, 0x30be, 0x30bf, 0x30c0, 0x30c1,
	0x30c2, 0x30c3, 0x30c4, 0x30c5, 0x30c6, 0x30c7, 0x30c8, 0x30c9, 0x30d0, 0x30d1, 0x30d2,
	0x30d3, 0x30d4, 0x30d5, 0x30d6, 0x30d7, 0x30d8, 0x30d9, 0x30e0, 0x30e1, 0x30e2, 0x30e3,
	0x30e4, 0x30e5, 0x30e6, 0x30e7, 0x30e8, 0x30e9, 0x30f0, 0x30f1, 0x30f2, 0x30f3, 0x30f4,
	// Numbers
	0xff10, 0xff11, 0xff12, 0xff13, 0xff14, 0xff15, 0xff16, 0xff17, 0xff18, 0xff19,
	// CJK Unifed Chars
	0x4e01, 0x4e02, 0x4e03, 0x4e04, 0x4e05, 0x4e06, 0x4e07, 0x4e08, 0x4e09, 0x4ea0, 0x4ea1,
}

var green256Scale = [...]int{
	// In case the terminal does not support True colours, let's use the codes below
	255, 251, 189, 194, 190, 156, 154, 120, 118, 82, 112, 76, 40, 70, 34, 29, 28, 22,
	59, 238, 235, 234, 16, 0,
}

type rbg struct {
	r int
	g int
	b int
}

var greenTrueScale = [...]rbg{
	// Not all terminals support RGB True Colors, for MAC iTerm2 works, but iTerm does not
	{255, 255, 255},
	{80, 210, 80}, {80, 200, 80}, {70, 195, 70}, {60, 190, 60}, {50, 185, 50},
	{40, 180, 40}, {30, 175, 30}, {30, 170, 30}, {20, 160, 20}, {10, 150, 10}, {10, 140, 10},
	{10, 130, 10}, {0, 120, 0}, {0, 100, 0}, {0, 90, 0}, {0, 80, 0}, {0, 70, 0}, {0, 60, 0},
	{0, 50, 0}, {0, 40, 0}, {0, 60, 0}, {0, 20, 0}, {0, 10, 0}, {0, 0, 0}, {0, 0, 0},
}

type printType struct {
	x     int
	y     int
	chars []string
	size  int
}

type jobMap map[int]bool

type jobType struct {
	mu   sync.Mutex
	id   jobMap
	stop bool
}

func (j *jobType) drawDown(chanPrt chan printType, col int) {
	// Allocate the job by col value
	j.markRunning(col)

	// Define the speed
	interval := rand.Intn(200)
	// Define the size
	size := rand.Intn(50) + 5

	var chars []string

	// Kanji chars use 2 spaces
	xPos := 1 + col*2

	for yPos := 1; yPos < tm.Height()+size; yPos++ {
		if j.stop {
			break
		}
		randomIndex := rand.Intn(len(matrixChars))
		char := fmt.Sprintf("%c", matrixChars[randomIndex])
		// Let's prepend the value, so new values are in first position
		chars = append(chars, "any")
		copy(chars[1:], chars)
		chars[0] = char
		if len(chars) > size {
			// Remove last one which is the last one to print
			chars = chars[:len(chars)-1]
		}
		chanPrt <- printType{x: xPos, y: yPos, chars: chars, size: size}
		time.Sleep((10 + time.Duration(interval)) * time.Millisecond)
	}

	// Free the job, so another can be launched
	j.markFinished(col)
}

func getColorFactor(size int) float64 {
	if trueColors {
		return float64(len(greenTrueScale)-1) / float64(size-1)
	}
	return float64(len(green256Scale)-1) / float64(size-1)
}

func getPrintChar(colorFactor float64, position int, char string) string {
	if trueColors {
		color := greenTrueScale[int((float64(position) * colorFactor))]
		return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", color.r, color.g, color.b, char)
	}
	color := green256Scale[int((float64(position) * colorFactor))]
	return fmt.Sprintf("\033[38;5;%dm%s", color, char)
}

func printLine(c chan printType) {
	// using range allow us to close the channel to end the function
	for i := range c {
		colorFactor := getColorFactor(i.size)
		for position, char := range i.chars {
			y := i.y - position
			pchar := getPrintChar(colorFactor, position, char)
			if y > 0 && y < tm.Height() {
				tm.MoveCursor(i.x, y)
				tm.Printf(pchar)
			}
		}
		tm.Flush()
	}
}

func (j *jobType) waitToFinish() {
	for {
		oneRunning := false
		for _, running := range j.id {
			if running == true {
				oneRunning = true
			}
		}
		if oneRunning == false {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func waitKeyboardPress(keyPressChan chan bool) {
	err := tb.Init()
	if err != nil {
		panic(err)
	}
	defer tb.Close()

	tm.MoveCursor(1, tm.Height())
	tm.Printf("Press any key to stop")

	tb.PollEvent()
	keyPressChan <- true

}

func (j *jobType) markFinished(id int) {
	j.mu.Lock()
	j.id[id] = false
	j.mu.Unlock()
}

func (j *jobType) markRunning(id int) {
	j.mu.Lock()
	j.id[id] = true
	j.mu.Unlock()
}

func jobManager(maxCollumns int) {

	// Initiating Jobs
	jobs := jobType{}
	jobs.id = make(jobMap)
	jobs.stop = false

	// Start Goroutine which monitors the keyboard
	keyPressChan := make(chan bool)
	go waitKeyboardPress(keyPressChan)

	// Start Goroutine which prints on screen
	printChan := make(chan printType)
	go printLine(printChan)

	// Let's randomly select an x value with up to "maxCollums"
	// Kanji consumes 2 normal chars, so it's limited by half the terminal width
	xValues := rand.Perm(int(tm.Width() / 2))[:maxCollumns]

	// Initializing all jobs as finished
	for _, x := range xValues {
		jobs.markFinished(x)
	}

Loop:
	for {
		for x, running := range jobs.id {
			select {
			case <-keyPressChan:
				jobs.stop = true
				break Loop
			default:
				if running == false {
					jobs.markRunning(x)
					go jobs.drawDown(printChan, x)
				}
			}
			time.Sleep(100 * time.Millisecond)
		}

	}

	jobs.waitToFinish()
	// Closing the print channel stops printing goroutine
	close(printChan)
}

func main() {

	tm.Clear()
	rand.Seed(time.Now().UTC().UnixNano())

	// Hiding the cursor with tput civis
	// Could not use exec.Command("tput", "civis"), but using 'sh' package works
	// TODO: find out why 'exec.Command' does not work
	sh.Command("tput", "civis").Run()

	// Kanji does use the same of two chars
	// TODO: Add command line parameter to define the max number of collumns
	Maxcollumns := tm.Width() / 2
	jobManager(Maxcollumns)

	// Reseting all colors
	tm.Println("\033[0m")

	// returning visual cursor
	sh.Command("tput", "cnorm").Run()

	// Moving down and clear all
	tm.MoveCursor(1, tm.Height())
	tm.Clear()
	tm.Flush()

}
