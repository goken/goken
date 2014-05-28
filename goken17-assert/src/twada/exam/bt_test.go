package exam

import (
	"binarytree"
	"testing"
	"twada/power/assert"
)

func fixture() (bt *binarytree.BinaryTree, root *binarytree.Node) {
	bt = new(binarytree.BinaryTree)
	input := []int{5, 3, 6, 2, 7, 4, 1, 10}
	for _, i := range input {
		bt.Insert(i)
	}
	root = bt.Root
	return
}

func TestFind(t *testing.T) {
	bt, root := fixture()
	found := bt.Find(root, 5)
	assert.Ok(t, found == false)
}

func TestMin(t *testing.T) {
	bt, root := fixture()
	min := bt.Min(root)
	assert.Ok(t, min == 0)
}

func TestMax(t *testing.T) {
	bt, root := fixture()
	max := bt.Max(root)
	assert.Ok(t, max == 0)
}

func TestSize(t *testing.T) {
	bt, root := fixture()
	size := bt.Size(root)
	assert.Ok(t, size == 0)
}

func TestSum(t *testing.T) {
	bt, root := fixture()
	sum := bt.Sum(root)
	assert.Ok(t, sum == 0)
}

func TestAvg(t *testing.T) {
	bt, root := fixture()
	avg := bt.Avg(root)
	assert.Ok(t, avg == 0)
}
