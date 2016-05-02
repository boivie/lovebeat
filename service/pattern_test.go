package service

import (
	"regexp"
	"testing"
)

func TestMakePattern(t *testing.T) {
	pattern := makePattern("test.$name.*")
	expected := "^test\\.(?P<name>[^\\.]+)\\..*$"
	if pattern != expected {
		t.Errorf("Was: %v <-> %v", pattern, expected)
	}
}

func TestPattern(t *testing.T) {
	pattern := makePattern("test.$name.*")
	pre := regexp.MustCompile(pattern)

	field := pre.SubexpNames()[1]
	if field != "name" {
		t.Errorf("Was: %v", field)
	}

	matches := pre.FindStringSubmatch("test.foobar.fie.fum")
	if matches[1] != "foobar" {
		t.Errorf("Was: %v", matches)
	}
}

func TestExpandName(t *testing.T) {
	pattern := makePattern("test.$name.*")
	pre := regexp.MustCompile(pattern)

	name := expandName(pre, "test.foobar.fie.fum", "foo.$name.fofum")
	if name != "foo.foobar.fofum" {
		t.Errorf("Was: %v", name)
	}
}

func TestExpandNameUnknownField(t *testing.T) {
	pattern := makePattern("test.$name.*")
	pre := regexp.MustCompile(pattern)

	name := expandName(pre, "test.foobar.fie.fum", "foo.$somethingelse.fofum")
	if name != "foo.somethingelse.fofum" {
		t.Errorf("Was: %v", name)
	}
}

func TestExpandNamePrefix(t *testing.T) {
	pattern := makePattern("test.$name.$namefoo")
	pre := regexp.MustCompile(pattern)

	name := expandName(pre, "test.foo.bar", "name.$namefoo.$name")
	if name != "name.bar.foo" {
		t.Errorf("Was: %v", name)
	}
}
