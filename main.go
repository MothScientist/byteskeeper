package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	progressbar "github.com/MothScientist/goProgressBar"
	"github.com/google/uuid"
)

var SessionUUID uuid.UUID
var WorkingPath string
var CurrentPath string
var SelectedHash HashAlgorithm

func init() {
	var err error
	SessionUUID, err = uuid.NewV7()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Session UUID: " + SessionUUID.String())

	CurrentPath, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// graceful shutdown
	rootCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	logFile := setupLogs(SessionUUID.String())
	defer logPanic()
	go func() {
		// If you don't do it via ctx.Done(), then when you cancel (Ctrl+C) you'll get a locked directory with logs
		<-rootCtx.Done()
		log.Print("Context cancelled, log file closed")
		logFile.Close()
	}()

	log.Println("Session UUID: " + SessionUUID.String())
	log.Printf("Datetime from session UUID: %s\n", GetDateFromUUIDv7(SessionUUID))
	log.Println("Current directory: " + CurrentPath)

	SelectedHash = getHashAlgorithm()
	WorkingPath = searchWorkingPath()
	log.Println("Working directory: " + WorkingPath)

	startTime := time.Now()

	root, countEdges, countFiles, countBytes := createDirectoriesTree(WorkingPath, nil)

	strData := fmt.Sprintf("%d directories, %d files found. %.2f MB will be processed.\n", countEdges, countFiles, float64(countBytes)/(1024*1024))
	fmt.Print(strData)
	log.Print(strData)

	files := root.GetAllWeightElementsWithPath()

	var readyFiles atomic.Int32
	var wg sync.WaitGroup
	wg.Add(1)
	go printProgressBar(rootCtx, &wg, countFiles, &readyFiles)
	hash := FilesHashSum(rootCtx, files, &readyFiles)
	wg.Wait()

	log.Println("Hash:", hash)
	fmt.Println("Hash:", hash)

	err := writeHashFile(hash)
	if err != nil {
		log.Fatal("Error writing output file")
	}

	stopTime := time.Now()
	timeProgramNs := stopTime.Sub(startTime)
	timeProgramSec := timeProgramNs.Seconds()

	fmt.Printf("The check took %.2f sec.\n", timeProgramSec)
	log.Printf("The check took %.2f sec.\n", timeProgramSec)

	fmt.Println("To close the program, press Enter...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// searchWorkingPath processes the string received by the user with the path to the directory
func searchWorkingPath() string {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter the path to the required directory: ")

	if scanner.Scan() {
		path := strings.TrimSpace(scanner.Text()) // Receive the entered text and remove the newline character
		if !validatePath(path) {
			fmt.Println("Incorrect path specified")
			log.Fatal("Incorrect path specified")
		}
		return path
	} else {
		fmt.Println("scanner.Scan() error")
		log.Fatal("scanner.Scan() error")
	}
	return ""
}

// getHashAlgorithm processes the input string with the user-selected hashing algorithm
func getHashAlgorithm() HashAlgorithm {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Select a hashing algorithm:\n0 - SHA256\n1 - Blake2b\n")

	if scanner.Scan() {
		selectedAlgorithm := strings.TrimSpace(scanner.Text())
		switch selectedAlgorithm {
		case "0":
			log.Println("Selected hash algorithm: " + SHA256.name)
			return SHA256
		case "1":
			log.Println("Selected hash algorithm: " + BLAKE2b.name)
			return BLAKE2b
		default:
			fmt.Println("Incorrect hash algorithm")
			log.Fatal("Incorrect hash algorithm")
		}
	} else {
		fmt.Println("scanner.Scan() error")
		log.Fatal("scanner.Scan() error")
	}
	return HashAlgorithm{}
}

// printProgressBar displays and updates the process progress bar in the console
func printProgressBar(ctx context.Context, wg *sync.WaitGroup, countFiles int, readyFiles *atomic.Int32) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	defer wg.Done()
	bar := progressbar.GetNewProgressBar()
	bar.SetColors([2]string{progressbar.ColorBlue, progressbar.ColorBrightCyan})
	for {
		current := int(readyFiles.Load())
		ready := int((float64(current) / float64(countFiles)) * 100)
		select {
		case <-ticker.C:
			pg, _ := bar.Update(ready)
			fmt.Print("\r")
			fmt.Print(pg)
			fmt.Printf(" (%d / %d)", current, countFiles)
		case <-ctx.Done():
			log.Println("Context cancelled, progress bar was stopped")
			return
		}
		if current == countFiles {
			break
		}
	}
	pg, _ := bar.Update(100)
	fmt.Print("\r")
	fmt.Println(pg)
}