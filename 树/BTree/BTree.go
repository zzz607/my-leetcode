package src

import (
	"fmt"
	"strings"
)

type Ordered interface {
	Integer | Float | ~string
}

type Integer interface {
	Signed | Unsigned
}

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Float interface {
	~float32 | ~float64
}

type BTNodePosi[T Ordered] *BTNode[T]

type BTNode[T Ordered] struct {
	parent BTNodePosi[T]
	key    []T
	child  []BTNodePosi[T]
}

func NewBTNode[T Ordered](e T, lc *BTNode[T], rc *BTNode[T]) *BTNode[T] {
	return &BTNode[T]{
		parent: nil,
		key:    []T{e},
		child:  []BTNodePosi[T]{lc, rc},
	}
}

func DefaultBTNode[T Ordered]() *BTNode[T] {
	return &BTNode[T]{
		parent: nil,
		key:    []T{},
		child:  []BTNodePosi[T]{},
	}
}

func (n *BTNode[T]) KeySize() int {
	return len(n.key)
}

func (n *BTNode[T]) ChildSize() int {
	return len(n.child)
}

// Split 将当前节点中的第sIdx开始的元素移动到other节点中
func (n *BTNode[T]) Split(sIdx int) (T, BTNodePosi[T]) {
	e := n.key[sIdx]

	other := DefaultBTNode[T]()
	other.key = append(other.key, n.key[sIdx+1:]...)
	other.child = append(other.child, n.child[sIdx+1:]...)

	n.key = n.key[:sIdx]
	n.child = n.child[:sIdx+1]

	return e, other
}

// Search 在当前节点中，找到不大于e的最大关键码
func (n *BTNode[T]) Search(e T) (int, bool) {
	for i := 0; i < len(n.key); i++ {
		if e == n.key[i] {
			return i - 1, true
		}
		if e < n.key[i] {
			return i - 1, false
		}
	}
	return len(n.key) - 1, false
}

// Insert 在当前节点中，插入一个关键码
func (n *BTNode[T]) Insert(idx int, e T, rc BTNodePosi[T]) {
	// 插入只能发生在叶节点，所以child不需要插入，
	// 但是上溢时需要将当前节点移动到其他节点中，所以需要将child移动
	// 这么说的话，这个函数还需要左右二个孩子节点指针？？
	n.key = append(n.key[:idx], append([]T{e}, n.key[idx:]...)...)
	n.child = append(n.child[:idx+1], append([]BTNodePosi[T]{rc}, n.child[idx+1:]...)...)
}

func (n *BTNode[T]) InsertL(idx int, e T, lc BTNodePosi[T]) {
	n.key = append(n.key[:idx], append([]T{e}, n.key[idx:]...)...)
	n.child = append(n.child[:idx], append([]BTNodePosi[T]{lc}, n.child[idx:]...)...)
}

// Remove 在当前节点中，删除一个关键码
// direction 删除节点时，一并删除节点左边还是右边的孩子链接
//
//	0 删除左边的孩子链接
//	1 删除右边的孩子链接
func (n *BTNode[T]) Remove(e T, direction int) bool {
	idx, ok := n.Search(e)
	if !ok {
		return false
	}

	idx++
	n.key = append(n.key[:idx], n.key[idx+1:]...)
	if direction == 0 {
		n.child = append(n.child[:idx], n.child[idx+1:]...)
	} else {
		n.child = append(n.child[:idx+1], n.child[idx+2:]...)
	}

	return true
}

type BTree[T Ordered] struct {
	// 存放的关键码总数
	size int
	// B-树的阶次，至少为3——创建时指定，一般不能修改
	order int
	root  BTNodePosi[T]
	hot   BTNodePosi[T]
}

func NewBTree[T Ordered](order int) *BTree[T] {
	if order < 3 {
		order = 3
	}

	root := DefaultBTNode[T]()
	root.child = append(root.child, nil)
	return &BTree[T]{
		order: order,
		size:  0,
		root:  root,
		hot:   nil,
	}
}

func (bt *BTree[T]) Order() int {
	return bt.order
}

func (bt *BTree[T]) Size() int {
	return bt.size
}

func (bt *BTree[T]) Root() BTNodePosi[T] {
	return bt.root
}

func (bt *BTree[T]) Empty() bool {
	return bt.root == nil
}

func (bt *BTree[T]) Search(e T) BTNodePosi[T] {
	v := bt.root
	bt.hot = nil

	for v != nil {
		r, ok := (*v).Search(e)
		if ok {
			return v
		}

		bt.hot = v
		v = v.child[r+1]
	}
	return nil
}

func (bt *BTree[T]) Insert(e T) bool {
	v := bt.Search(e)
	if v != nil {
		return false
	}

	r, _ := (*bt.hot).Search(e)
	(*bt.hot).Insert(r+1, e, nil)
	bt.size++
	bt.solveOverflow(bt.hot)
	return true
}

// Remove 在B树中删除一个关键码
// todo 暂只支持删除叶节点
func (bt *BTree[T]) Remove(e T) bool {
	v := bt.Search(e)
	if v == nil {
		return false
	}

	// 如果删除的不是叶节点，则将其与直接后继交换
	if v.child[0] != nil {
		r, _ := (*v).Search(e)
		rc := v.child[r+1]
		for rc.child[0] != nil {
			rc = rc.child[0]
		}

		successor := rc.key[0]
		v.key[r] = successor
		e = successor
		v = rc
	}

	ok := (*v).Remove(e, 1)
	if !ok {
		return false
	}
	bt.size--
	bt.solveUnderflow(v)

	// 如果节点被全部删除，为了防止再插入的时候出错，此时将B树初始化
	if bt.root == nil {
		bt.root = DefaultBTNode[T]()
		bt.root.child = append(bt.root.child, nil)
	}
	return true
}

// Print 层序遍历打印B树
func (bt *BTree[T]) Print() {
	queue := []BTNodePosi[T]{bt.root}
	rightNode := bt.root
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]

		fmt.Print(v.key)
		fmt.Print("  ")
		if v == rightNode {
			fmt.Println("")
			rightNode = v.child[(*v).ChildSize()-1]
		}

		for _, child := range v.child {
			if child != nil {
				queue = append(queue, child)
			}
		}
	}
}

// String 层序遍历打印B树
func (bt *BTree[T]) String() string {
	var str string

	queue := []BTNodePosi[T]{bt.root}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]

		str += fmt.Sprint(v.key) + " "

		for _, child := range v.child {
			if child != nil {
				queue = append(queue, child)
			}
		}
	}

	return strings.TrimSpace(str)
}

func (bt *BTree[T]) solveOverflow(v BTNodePosi[T]) {
	// 如果当前节点未上溢，则返回
	if (*v).ChildSize() <= bt.order {
		return
	}

	mid := bt.order / 2
	e, other := (*v).Split(mid)

	// 插入到父节点中
	if v.parent == nil {
		bt.root = NewBTNode(e, v, other)
		v.parent = bt.root
		other.parent = bt.root
	} else {
		r, _ := (*v.parent).Search(e)
		(*v.parent).Insert(r+1, e, other)
		other.parent = v.parent
	}

	// 更新孩子指针的父指针
	for _, child := range other.child {
		if child != nil {
			child.parent = other
		}
	}

	bt.solveOverflow(v.parent)
}

func (bt *BTree[T]) solveUnderflow(v BTNodePosi[T]) {
	borrowFromRightSibling := func(v BTNodePosi[T], r int) {
		// 右兄弟有足够的关键码，则从右兄弟借一个关键码
		rightSibling := v.parent.child[r+1]
		re := (*rightSibling).key[0]
		rp := (*rightSibling).child[0]
		pe := (*v.parent).key[r]

		v.parent.key[r] = re
		(*v).Insert((*v).KeySize(), pe, rp)
		if rp != nil {
			rp.parent = v
		}
		(*rightSibling).Remove(re, 0)
	}

	borrowFromLeftSibling := func(v BTNodePosi[T], r int) {
		// 左兄弟有足够的关键码，则从左兄弟借一个关键码
		leftSibling := v.parent.child[r-1]
		le := (*leftSibling).key[(*leftSibling).KeySize()-1]
		lp := (*leftSibling).child[(*leftSibling).ChildSize()-1]
		pe := (*v.parent).key[r-1]

		(*v.parent).key[r-1] = le
		(*v).InsertL(0, pe, lp)
		if lp != nil {
			lp.parent = v
		}
		(*leftSibling).Remove(le, 1)
	}

	mergeWithRightSibling := func(v BTNodePosi[T], r int) {
		// 与右兄弟合并。把右兄弟合并到当前节点
		rightSibling := v.parent.child[r+1]
		pe := (*v.parent).key[r]
		(*v).Insert((*v).KeySize(), pe, (*rightSibling).child[0])
		for idx, key := range (*rightSibling).key {
			(*v).Insert((*v).KeySize(), key, (*rightSibling).child[idx+1])
		}
		for _, child := range (*rightSibling).child {
			if child != nil {
				child.parent = v
			}
		}
		// 合并之后，父节点中的关键码和右兄弟的指针需要删除
		// 右节点本身不需要删除，golang会自动回收
		(*v.parent).Remove(pe, 1)
	}

	mergeWithLeftSibling := func(v BTNodePosi[T], r int) {
		// 与左兄弟合并。把当前节点合并到左兄弟
		leftSibling := v.parent.child[r-1]
		pe := (*v.parent).key[r-1]
		(*leftSibling).Insert((*leftSibling).KeySize(), pe, (*v).child[0])
		for idx, key := range (*v).key {
			(*leftSibling).Insert((*leftSibling).KeySize(), key, (*v).child[idx+1])
		}
		for _, child := range (*v).child {
			if child != nil {
				child.parent = leftSibling
			}
		}
		// 合并之后，父节点中的关键码和当前节点的指针需要删除
		// 当前节点本身不需要删除，golang会自动回收
		(*v.parent).Remove(pe, 1)
	}

	// 如果当前节点未下溢，则返回
	minChildSize := (bt.order + 1) / 2
	if (*v).ChildSize() >= minChildSize {
		return
	}

	// 如果当前节点为根节点，则返回
	if v.parent == nil {
		if (*v).KeySize() == 0 {
			bt.root = v.child[0]
			if bt.root != nil {
				bt.root.parent = nil
			}
		}
		return
	}

	// 在父节点中的位置
	var r int
	for r = 0; r < (*v.parent).ChildSize(); r++ {
		if (*v.parent).child[r] == v {
			break
		}
	}

	// 是否有右兄弟
	if r < (*v.parent).ChildSize()-1 {
		rightSibling := v.parent.child[r+1]
		if rightSibling != nil && (*rightSibling).ChildSize() > minChildSize {
			borrowFromRightSibling(v, r)
			return
		}
	}

	// 是否有左兄弟
	if r > 0 {
		leftSibling := v.parent.child[r-1]
		if leftSibling != nil && (*leftSibling).ChildSize() > minChildSize {
			borrowFromLeftSibling(v, r)
			return
		}
	}

	// 无法从兄弟节点借，则需要合并
	if r < (*v.parent).ChildSize()-1 {
		mergeWithRightSibling(v, r)
	} else {
		mergeWithLeftSibling(v, r)
	}

	// 继续向上递归
	bt.solveUnderflow(v.parent)
}
