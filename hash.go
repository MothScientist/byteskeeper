package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"hash"

	"golang.org/x/crypto/blake2b"
)

// FilesHashSum calculates a hash for multiple files (the result does not depend on the order of the data in the passed slice)
func FilesHashSum(ctx context.Context, files []File, readyFiles *atomic.Int32) string {
	var wg sync.WaitGroup
	countFiles := len(files)
	workers := make(chan struct{}, runtime.NumCPU()*4) // Maximum number of simultaneously running goroutines
	fileHashsumCh := make(chan FileHash, countFiles)

	wg.Add(countFiles) //  We immediately set the counter to the full number of files
	for _, file := range files {
		if ctx.Err() != nil {
			log.Print("Context cancelled, stopping file processing")
			break
		}
		workers <- struct{}{} // We take up space for a new goroutine
		go func(file File) {
			defer wg.Done()
			defer func() { readyFiles.Add(1) }()
			defer func() { <-workers }() // Freeing up space in the active goroutine pool
			select {
			case <-ctx.Done():
				log.Print("Context cancelled, reading of the file was interrupted")
				return
			default:
				hashsum, err := fileHashSum(ctx, file)
				fileHashsumCh <- FileHash{
					File:  file,
					Hash:  hashsum,
					Error: err,
				}
			}
		}(file)
	}

	wg.Wait()
	close(fileHashsumCh)

	// We collect hashes and sort them by file name.
	hashesData := sortedHashesByFilename(fileHashsumCh, countFiles)

	// It's important to always sort in the same order for a repeatable result
	// so we hash in alphabetical order by path and file names
	hash := sha256.New()
	for _, hashData := range hashesData {
		if hashData.Error == nil {
			hash.Write([]byte(hashData.Hash))
		} else {
			hashErr := fmt.Sprintf("Для файла %s не удалось вычислить хеш сумму", hashData.File.Name)
			errorCore := fmt.Sprintf(" Error: %v", hashData.Error)
			log.Println("[ERROR] " + hashErr + "; Error:" + errorCore)
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// fileHashSum calculates the hash of a file by its name.
func fileHashSum(ctx context.Context, file File) (string, error) {
	filePath := filepath.Join(file.Path, file.Name)
	openFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func(osFile *os.File) {
		err := osFile.Close()
		if err != nil {
			log.Println("[WARNING] не получилось закрыть файл")
		}
	}(openFile)

	return getFileHash(ctx, openFile)
}

// getFileHash returns the hash of a file given its name and path.
func getFileHash(ctx context.Context, f *os.File) (string, error) {
	var hash hash.Hash
	switch SelectedHash.name {
	case "sha256":
		hash = sha256.New()
	case "blake2b":
		var err error
		hash, err = blake2b.New256(nil)
		if err != nil {
			log.Print("Error calling function blake2b")
			return "", err
		}
	}

	buffer := make([]byte, bufferSize)
	for {
		if ctx.Err() != nil {
			log.Print("Context cancelled, stopping file hashing")
			break
		}
		n, err := f.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		hash.Write(buffer[:n]) // At each iteration, the data in buf is overwritten.
	}
	hash.Write([]byte(f.Name())) // To avoid collisions, add the file name with path to the hash
	fileHash := hex.EncodeToString(hash.Sum(nil))
	log.Println("File: " + f.Name() + ", Hash: " + fileHash)
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// sortedHashesByFilename sort hashes from a channel by full path and file name
func sortedHashesByFilename(fileHashCh <-chan FileHash, buffSize int) []FileHash {
	hashesData := make([]FileHash, 0, buffSize)
	for fh := range fileHashCh {
		hashesData = append(hashesData, fh)
	}
	sort.Slice(hashesData, func(i, j int) bool {
		if hashesData[i].File.Path != hashesData[j].File.Path {
			return hashesData[i].File.Path < hashesData[j].File.Path
		}
		return hashesData[i].File.Name < hashesData[j].File.Name
	})
	return hashesData
}

// writeHashFile writes the resulting file with the resulting hash
func writeHashFile(hash string) error {
	filePath := filepath.Join(CurrentPath, SessionUUID.String(), SelectedHash.name)
	file, err := os.Create(filePath)
	if err != nil {
		log.Println("[WARNING] ошибка при создании файла: os.Create(" + filePath + ")")
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("[WARNING] ошибка при закрытии файла: file.Close()")
		}
	}(file)

	_, err = file.WriteString(hash)
	if err != nil {
		log.Println("[WARNING] ошибка при записи в файл: file.WriteString(" + hash + ")")
		return err
	}
	return nil
}