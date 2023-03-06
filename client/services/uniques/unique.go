package uniques

import (
	"fmt"
	"strconv"
	"strings"
)

type Unique struct {
	used map[string]struct{}
}

func NewUnique() *Unique {
	used := map[string]struct{}{}
	return &Unique{used: used}
}

// nameがすでにusedSetに存在する場合は末尾に連番の数字を足したものを返す
// 大文字,小文字の区別はしない
/* 例:
 * unique := NewUnique()
 * res := unique.Take("hoge")
 * // res is "hoge"
 * res := unique.Take("hoge")
 * // res is "hoge 2"
 * res := unique.Take("HOGE")
 * // res is "HOGE 3"
 */
func (u *Unique) Take(name string) string {
	uniqueName := name
	for i := 2; i < 1e7; i++ {
		if u.IsUsed(uniqueName) {
			uniqueName = name + " " + strconv.Itoa(i)
		} else {
			break
		}
	}
	u.Add(uniqueName)
	return uniqueName
}

func (u *Unique) IsUsed(name string) bool {
	low := strings.ToLower(name)
	if _, isUsed := u.used[low]; isUsed {
		return true
	}
	return false
}

func (u *Unique) Add(name string) error {
	if u.IsUsed(name) {
		return fmt.Errorf("add error: %s is already used", name)
	}
	low := strings.ToLower(name)
	u.used[low] = struct{}{}
	return nil
}

func (u *Unique) Delete(name string) error {
	if !u.IsUsed(name) {
		return fmt.Errorf("delete error: %s is not used", name)
	}
	delete(u.used, name)
	return nil
}
