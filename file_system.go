package main

import (
	"log"
	"fmt"
	"os"
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
		newDir = NewNode(dirPath, nil, nil)
	} else {
		*countEdges++
		newDir = node.Insert(dirPath, nil)
	}

	entries, _ := os.ReadDir(dirPath)
	for _, entry := range entries {
		isDir := entry.IsDir()
		elemName := entry.Name()
		if isDir {
			dirName := dirPath + "/" + elemName
			fillRecursiveDirectoriesTree(dirName, newDir, countEdges, countFiles, countBytes)
		} else {
			newDir.AddWeightElem(elemName, false)
			*countFiles++

			// Next we add the file size to the final size
			info, err := entry.Info()
			if err != nil {
				filePath := dirPath + "/" + elemName
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