package main

import (
	"fmt"
	"log"
)

type Tree interface {
	NewNode(string, *Node) *Node
	Insert(string) *Node
	AddWeightElem(string, bool) bool
	GetAllWeightElementsWithPath() []string
	PrintTree(int)
	PrintTreeLog(int)
}

// Node describes all the properties of the node
type Node struct {
	head         *Node    // Top-level directory
	childs       []*Node  // Nested directories
	weight       int      // Number of final files in the directory
	name         string   // Node (directory) name
	nodeElements []string //Files inside the directory (file names)
}

// NewNode creates a new node
func NewNode(name string, head *Node) *Node {
	return &Node{
		head:         head,
		childs:       []*Node{},
		name:         name,
	}
}

// Insert inserts a new node into the tree
func (n *Node) Insert(name string) *Node {
	NewNode := NewNode(name, n)
	n.childs = append(n.childs, NewNode)
	return NewNode
}

// updateWeight recursively updates the weight of a node and all parent nodes
func (n *Node) updateWeight(delta int) {
	currentNode := n
	for currentNode != nil {
		if delta > 0 {
			currentNode.weight += delta
		} else {
			currentNode.weight -= -delta
		}
		currentNode = currentNode.head
	}
}

// AddWeightElem adds a new element to the node (its weight element), taking into account the uniqueness
func (n *Node) AddWeightElem(newElement string) {
	n.nodeElements = append(n.nodeElements, newElement)
	n.updateWeight(1)
}

// GetAllWeightElementsWithPath get all files from a tree (breadth-first search)
func (n *Node) GetAllWeightElementsWithPath() (result []File) {
	queue := []*Node{n}
	for len(queue) > 0 {
		currentNode := queue[0]
		queue = queue[1:]                            // Извлекаем из очереди узел, помещенный в currentNode
		queue = append(queue, currentNode.childs...) // Добавляем детей узла в очередь
		for _, elem := range currentNode.nodeElements {
			result = append(result, File{Path: currentNode.name, Name: elem})
		}
	}
	return
}

// PrintTree prints the tree structure
func (n *Node) PrintTree(level int) {
	indent := ""
	for range level {
		indent += "  "
	}

	fmt.Printf("%s -> (Name: %s, Files: %d, nodeElements: %s)\n", indent, n.name, n.weight, n.nodeElements)

	for _, child := range n.childs {
		child.PrintTree(level + 1)
	}
}

// PrintTreeLog logging the tree structure
func (n *Node) PrintTreeLog(level int) {
	indent := ""
	for range level {
		indent += "  "
	}

	log.Printf("%s[Name: %s, Files: %d, nodeElements: %s]\n", indent, n.name, n.weight, n.nodeElements)

	for _, child := range n.childs {
		child.PrintTreeLog(level + 1)
	}
}