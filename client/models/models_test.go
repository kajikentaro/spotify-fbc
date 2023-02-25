package models

import "testing"

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
