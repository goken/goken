package main

import (
	"fmt"
	"runtime"
)

type Node struct {
	value  int
	left   *Node
	right  *Node
	parent *Node
}

type BinaryTree struct {
	root *Node
}

func addNode(parent *Node, value int) *Node {
	return &Node{
		value:  value,
		left:   nil,
		right:  nil,
		parent: parent,
	}
}

func (b *BinaryTree) insertNode(root *Node, parent *Node, value int) *Node {
	switch {
	case root == nil:
		return addNode(parent, value)
	case value <= root.value:
		root.left = b.insertNode(root.left, root, value)
	case value > root.value:
		root.right = b.insertNode(root.right, root, value)
	}
	return root
}

func (b *BinaryTree) Insert(value int) (n *Node) {
	if b.root == nil {
		n = addNode(nil, value)
		b.root = n
	} else {
		n = b.insertNode(b.root, nil, value)
	}
	return
}

func (b *BinaryTree) Find(root *Node, value int) bool {
	if root == nil {
		return false
	} else if root.value == value {
		return true
	} else if root.value > value {
		return b.Find(root.left, value)
	} else {
		return b.Find(root.right, value)
	}
}

func (b *BinaryTree) Size(root *Node) (ret int) {
	if root == nil {
		return 0
	}

	return b.Size(root.left) + b.Size(root.right) + 1
}

func (b *BinaryTree) Max(root *Node) int {
	if root.right == nil {
		return root.value
	} else {
		return b.Max(root.right)
	}
}

func (b *BinaryTree) Min(root *Node) int {
	if root.left == nil {
		return root.value
	} else {
		return b.Min(root.left)
	}
}

func (b *BinaryTree) Avg(root *Node) float64 {
	return float64(b.Sum(root)) / float64(b.Size(root))
}

func (b *BinaryTree) Sum(root *Node) (sum int) {
	tree_size := b.Size(root)
	ch := make(chan int, tree_size)
	out := make(chan bool, tree_size)

	go b.Walk(root, tree_size, ch, out)

	for i := range ch {
		sum += i
	}
	return sum
}

func (b *BinaryTree) Walk(root *Node, numNode int, ch chan int, out chan bool) {
	go b.WalkWorker(root, ch, out)
	for i := 0; i < numNode; i++ {
		<-out
	}
	close(ch)
}

func (b *BinaryTree) WalkWorker(root *Node, ch chan int, out chan bool) {
	if root == nil {
		return
	}
	ch <- root.value
	go b.WalkWorker(root.left, ch, out)
	go b.WalkWorker(root.right, ch, out)
	out <- true
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	bt := new(BinaryTree)
	input := []int{5, 3, 6, 2, 7, 4, 1, 10}

	for _, i := range input {
		bt.Insert(i)
	}

	root := bt.root
	fmt.Printf("This tree has 5: %v \n", bt.Find(root, 5))
	fmt.Printf("Min node is %v\n", bt.Min(root))
	fmt.Printf("Max node is %v\n", bt.Max(root))
	fmt.Printf("This tree has %v nodes\n", bt.Size(root))
	fmt.Printf("Sum: %v\n", bt.Sum(root))
	fmt.Printf("Avg: %v\n", bt.Avg(root))

}
