package main

import (
	"log"
	"fmt"
	"os"
	"path/filepath"
)

// createDirectoriesTree fillings the tree structure with data from the system, returning the resulting values
func createDirectoriesTree(dirPath string, node *Node) (root *Node, countEdges int, countFiles int, countBytes int64) {
	fmt.Println("Building a file system tree...")
	root = fillRecursiveDirectoriesTree(dirPath, node, &countEdges, &countFiles, &countBytes)
	fmt.Print("\n")
	return
}

// fillRecursiveDirectoriesTree recursively traverses all files starting from the top directory, adding them to the structure
func fillRecursiveDirectoriesTree(dirPath string, node *Node, countEdges *int, countFiles *int, countBytes *int64) *Node {
	var newDir *Node
	if node == nil {
		newDir = NewNode(dirPath, nil)
	} else {
		*countEdges++
		newDir = node.Insert(dirPath)
	}

	entries, _ := os.ReadDir(dirPath)
	for _, entry := range entries {
		isDir := entry.Type() & os.ModeDir != 0
		isSymlink := entry.Type() & os.ModeSymlink != 0
		elemName := entry.Name()
		if isDir {
			dirName := dirPath + "/" + elemName
			fillRecursiveDirectoriesTree(dirName, newDir, countEdges, countFiles, countBytes)
		} else if isSymlink {
			continue // TODO
		} else {
			newDir.AddWeightElem(elemName)
			*countFiles++

			// Next we add the file size to the final size
			info, err := entry.Info()
			if err != nil {
				filePath := filepath.Join(dirPath, elemName)
				log.Println("Не удалось получить информацию о файле: " + filePath)
			} else {
				*countBytes += info.Size()
			}

		}
	}

	if node == nil {
		return newDir
	}
	return nil
}