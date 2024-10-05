package src

import (
	"fmt"
	"testing"
)

func TestBTree(t *testing.T) {
	data := []struct {
		order             int
		keys              []int
		expectAfterInsert []string
		expectAfterRemove []string
	}{
		{
			order: 3,
			keys:  []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			expectAfterInsert: []string{
				"[1]",
				"[1 2]",
				"[2] [1] [3]",
				"[2] [1] [3 4]",
				"[2 4] [1] [3] [5]",
				"[2 4] [1] [3] [5 6]",
				"[4] [2] [6] [1] [3] [5] [7]",
				"[4] [2] [6] [1] [3] [5] [7 8]",
				"[4] [2] [6 8] [1] [3] [5] [7] [9]",
				"[4] [2] [6 8] [1] [3] [5] [7] [9 10]",
			},
			expectAfterRemove: []string{
				"[6] [4] [8] [2 3] [5] [7] [9 10]",
				"[6] [4] [8] [3] [5] [7] [9 10]",
				"[6 8] [4 5] [7] [9 10]",
				"[6 8] [5] [7] [9 10]",
				"[8] [6 7] [9 10]",
				"[8] [7] [9 10]",
				"[9] [8] [10]",
				"[9 10]",
				"[10]",
				"[]",
			},
		},
		{
			order: 3,
			keys:  []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			expectAfterInsert: []string{
				"[10]",
				"[9 10]",
				"[9] [8] [10]",
				"[9] [7 8] [10]",
				"[7 9] [6] [8] [10]",
				"[7 9] [5 6] [8] [10]",
				"[7] [5] [9] [4] [6] [8] [10]",
				"[7] [5] [9] [3 4] [6] [8] [10]",
				"[7] [3 5] [9] [2] [4] [6] [8] [10]",
				"[7] [3 5] [9] [1 2] [4] [6] [8] [10]",
			},
			expectAfterRemove: []string{
				"[5] [3] [7] [1 2] [4] [6] [8 9]",
				"[5] [3] [7] [1 2] [4] [6] [8]",
				"[3 5] [1 2] [4] [6 7]",
				"[3 5] [1 2] [4] [6]",
				"[3] [1 2] [4 5]",
				"[3] [1 2] [4]",
				"[2] [1] [3]",
				"[1 2]",
				"[1]",
				"[]",
			},
		},
	}

	for i, v := range data {
		fmt.Printf("test %dth....\n", i)

		bt := NewBTree[int](v.order)
		for idx, n := range v.keys {
			bt.Insert(n)
			// fmt.Println("after insert:", n)
			// bt.Print()
			bs := bt.String()
			if v.expectAfterInsert[idx] != bs {
				t.Errorf("expectAfterInsert[%d] = %s, but got %s",
					n, v.expectAfterInsert[idx], bs)
			}
		}

		for idx, n := range v.keys {
			bt.Remove(n)
			// fmt.Println("after remove:", n)
			// bt.Print()
			bs := bt.String()
			if v.expectAfterRemove[idx] != bs {
				t.Errorf("expectAfterRemove[%d] = %s, but got %s",
					n, v.expectAfterRemove[idx], bs)
			}
		}

		fmt.Println("\n--------------------------------")
	}
}
