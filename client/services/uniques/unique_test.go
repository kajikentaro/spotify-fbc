package uniques

import "testing"

func Test_unique(t *testing.T) {
	unique := NewUnique()

	actual := unique.Take("hoge")
	expected := "hoge"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = unique.Take("hoge")
	expected = "hoge 2"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = unique.Take("HOGE")
	expected = "HOGE 3"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = unique.Take("hoge 2")
	expected = "hoge 2 2"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}
}
