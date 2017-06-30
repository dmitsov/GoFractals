package main

import (
	"fmt"
	"image"
	"image/color"
//	"image/color/palette"
	"image/png"
	"math/cmplx"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	maxIter                            = 700
	rMin                               = -2.
	rMax                               = 2.
	iMin                               = -2.
	iMax                               = 2.
	imgWidth                           = 640
	imgHeight                          = 640
	grtnsCount                         = 1
	outputFile                         = "zad18.png"
	funcMap    map[string]func(string) = map[string]func(string){
		"-t":      setTaskCount,
		"-task":   setTaskCount,
		"-o":      setOutput,
		"-output": setOutput,
		"-r":      setRect,
		"-rect":   setRect,
		"-s":      setSize,
		"-size":   setSize,
	}
	
	isQuiet bool = false
)

func mandelbrot(a complex128) color.Color {
	i := 0
	for z := a; cmplx.Abs(z) < 50 && i < maxIter; i++ {
		z = cmplx.Exp(cmplx.Cos(z * a))
	}
	
	k := i / maxIter;
	
	return color.RGBA{ uint8(255 * k), uint8(80 * (1 - k) + 255 * k), uint8(140 * (1 - k) + 255 * k), 255}
}

func mandelbrotSet(img *image.RGBA, startX, startY, width, height int) {
	for x := startX; x < startX+width; x++ {
		for y := startY; y < startY+height; y++ {
			pixelColor := mandelbrot(complex(
				(rMax-rMin)*float64(x)/float64(imgWidth-1)+rMin,
				(iMax-iMin)*float64(y)/float64(imgHeight-1)+iMin)).(color.RGBA)
			img.Set(x, y, pixelColor)
		}
	}

}

func setRect(rectStr string) {
	dims := strings.Split(rectStr, ":")

	var realMin, realMax, imagMin, imagMax float64
	var err error
	if realMin, err = strconv.ParseFloat(dims[0], 64); err != nil {
		fmt.Println("Couldn't set rect. Error: ", err.Error())
		return
	}

	if realMax, err = strconv.ParseFloat(dims[1], 64); err != nil {
		fmt.Println("Couldn't set rect. Error: ", err.Error())
		return
	}

	if imagMin, err = strconv.ParseFloat(dims[2], 64); err != nil {
		fmt.Println("Couldn't set rect. Error: ", err.Error())
		return
	}

	if imagMax, err = strconv.ParseFloat(dims[3], 64); err != nil {
		fmt.Println("Couldn't set rect. Error: ", err.Error())
		return
	}

	rMin, rMax, iMin, iMax = realMin, realMax, imagMin, imagMax
}

func setSize(size string) {
	dims := strings.Split(size, "x")
	var width, height int64
	var err error

	if width, err = strconv.ParseInt(dims[0], 10, 64); err != nil {
		fmt.Println("Couldn't set size. Error: ", err.Error())
		return
	}

	if height, err = strconv.ParseInt(dims[1], 10, 64); err != nil {
		fmt.Println("Couldn't set size. Error: ", err.Error())
		return
	}

	imgWidth, imgHeight = int(width), int(height)
}

func setTaskCount(taskNum string) {
	taskCount, err := strconv.ParseInt(taskNum, 10, 64)
	if err != nil {
		fmt.Println("Couldn't set task count. Error: ", err.Error())
		return
	}

	grtnsCount = int(taskCount)
}

func setOutput(output string) {
	outputFile = output
}

func main() {

	args := os.Args[1:]
	
	fmt.Println("Max threads: ",runtime.GOMAXPROCS(32))
	
	for i := 0; i < len(args); i += 2 {
		if args[i] == "-q" || args[i] == "-quiet" {
			isQuiet = true
		} else {
			funcMap[args[i]](args[i+1])
		}
	}

	bounds := image.Rect(0, 0, imgWidth, imgHeight)
	img := image.NewRGBA(bounds)
	currentTime := time.Now()
	
	if !isQuiet {
		fmt.Printf("Threads used in current run: %d\n", grtnsCount)
	}
	granularity := grtnsCount
	
	if grtnsCount == 1 {
		mandelbrotSet(img, 0, 0, imgWidth, imgHeight)
	} else {
		var wg sync.WaitGroup

		blockWidth := imgWidth / granularity
		blockHeight := imgHeight / granularity

		var columnNum int = granularity
		var rowNum int = granularity

		if imgWidth%granularity != 0 {
			columnNum = granularity + 1
		}

		if imgHeight%granularity != 0 {
			rowNum = granularity + 1
		}

		var blockChan [][]int = make([][]int, columnNum*rowNum)

		for i := 0; i < granularity; i++ {
			for j := 0; j < granularity; j++ {
				blockChan[i*granularity+j] = []int{i * blockWidth, j * blockHeight, blockWidth, blockHeight}
			}
		}

		if rowNum > granularity {
			for i := 0; i < granularity; i++ {
				blockChan[granularity*granularity+i] = []int{i * blockWidth, granularity * blockHeight, blockWidth, imgHeight%granularity}
			}
			
			blockChan[granularity*granularity + granularity] = []int{granularity * blockWidth, granularity * blockHeight, imgWidth % granularity, imgHeight % granularity}
		}

		if columnNum > granularity {
			for i := 0; i < granularity; i++ {
				blockChan[granularity*granularity+columnNum+i] = []int{granularity * blockWidth, i * blockHeight, imgWidth % granularity, blockHeight}
			}
		}

		var blockCount int

		// fmt.Println("Block count ", rowNum*columnNum)

		for i := 0; i < grtnsCount - 1; i++ {
			wg.Add(1)
			go func(id int, blocks [][]int) {
				defer wg.Done()
				//runtime.Gosched()
				grtnTime := time.Now()
				if !isQuiet {
					fmt.Printf("Thread %d started\n", id)
				}
				
				for blockCount < columnNum*rowNum {
					block := blocks[blockCount]
					blockCount++
					mandelbrotSet(img, block[0], block[1], block[2], block[3])
				}
				
				if !isQuiet {
					fmt.Printf("Thread %d stopped\n", id)
					fmt.Printf("Thread %d execution time %fms\n", id, time.Since(grtnTime).Seconds()* 1000)
				}
				
			}(i, blockChan)
		}

		// fmt.Println("Goroutines started")
		for blockCount < columnNum*rowNum {
			block := blockChan[blockCount]
			blockCount++
			mandelbrotSet(img, block[0], block[1], block[2], block[3])
		}
		
		wg.Wait()

	}

	renderingTime := time.Since(currentTime)

	fmt.Printf("Total execution time for current run  %fms\n", float64(renderingTime.Seconds()) * 1000)

	f, err := os.Create(outputFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err = png.Encode(f, img); err != nil {
		fmt.Println(err)
	}
	if err = f.Close(); err != nil {
		fmt.Println(err)
	}
}
