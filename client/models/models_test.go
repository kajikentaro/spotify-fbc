package models

import "testing"

func Test_unique(t *testing.T) {
	usedMap := map[string]struct{}{}

	actual := unique(&usedMap, "hoge")
	expected := "hoge"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = unique(&usedMap, "hoge")
	expected = "hoge 2"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = unique(&usedMap, "hoge 2")
	expected = "hoge 2 2"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}
}

func Test_replaceBannedCharacter(t *testing.T) {
	actual := replaceBannedCharacter("foo\\bar")
	expected := "foo bar"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = replaceBannedCharacter("foo/bar")
	expected = "foo bar"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = replaceBannedCharacter("foo:bar")
	expected = "foo bar"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = replaceBannedCharacter("foo*bar")
	expected = "foo bar"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = replaceBannedCharacter("foo?bar")
	expected = "foo bar"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = replaceBannedCharacter("foo\"bar")
	expected = "foo bar"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = replaceBannedCharacter("foo<>bar")
	expected = "foo  bar"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}

	actual = replaceBannedCharacter("foo|bar")
	expected = "foo bar"
	if actual != expected {
		t.Errorf("actual: %s, expected: %s", actual, expected)
	}
}
