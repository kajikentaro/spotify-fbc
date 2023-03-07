package repositories

import (
	"reflect"
	"testing"
)

func Test_splitProcess(t *testing.T) {
	massiveArray := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	cnt := 0
	expected := [][]int{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {9}}
	err := splitProcess(3, massiveArray, func(actual []int) error {
		exp := expected[cnt]
		isSame := reflect.DeepEqual(actual, exp)
		if !isSame {
			t.Error("actual:", actual, "expected:", exp)
		}
		cnt++
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
