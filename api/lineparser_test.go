package api

import (
	"testing"
)

func testEq(a, b []LineCommand) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Action != b[i].Action ||
		a[i].Name != b[i].Name ||
		a[i].Value != b[i].Value {
			return false
		}
	}

	return true
}

func TestParseEmpty(t *testing.T) {
	cmds := Parse([]byte{})
	if len(cmds) != 0 {
		t.Errorf("Failed")
	}
}

func TestParseBeat(t *testing.T) {
	cmds := Parse([]byte("foo.bar.beat:1|c"))
	ref := []LineCommand{LineCommand{"beat", "foo.bar", 1000}}
	if !testEq(cmds, ref) {
		t.Errorf("Failed %v %v", cmds, ref)
	}
}

func TestParseTimeout(t *testing.T) {
	cmds := Parse([]byte("foo.bar.timeout:1|c"))
	ref := []LineCommand{LineCommand{"timeout", "foo.bar", 1000}}
	if !testEq(cmds, ref) {
		t.Errorf("Failed %v %v", cmds, ref)
	}
}

func TestParseAutoBeat(t *testing.T) {
	cmds := Parse([]byte("foo.bar.autobeat:1|c"))
	ref := []LineCommand{LineCommand{"autobeat", "foo.bar", 1000}}
	if !testEq(cmds, ref) {
		t.Errorf("Failed %v %v", cmds, ref)
	}
}

func TestParseNegative(t *testing.T) {
	cmds := Parse([]byte("foo.bar.timeout:-2|g"))
	ref := []LineCommand{LineCommand{"timeout", "foo.bar", -2}}
	if !testEq(cmds, ref) {
		t.Errorf("Failed %v %v", cmds, ref)
	}
}

func TestParseInvalid(t *testing.T) {
	cmds := Parse([]byte("foo.bar.invalid:1|c"))
	if len(cmds) != 0 {
		t.Errorf("Failed %v", cmds)
	}
}

func TestParseMultiple(t *testing.T) {
	cmds := Parse([]byte("foo.bar.timeout:1|c\nfoo.fum.timeout:4|g\nfoo.fie.beat:8|c"))
	ref := []LineCommand{LineCommand{"timeout", "foo.bar", 1000},
		LineCommand{"timeout", "foo.fum", 4000},
		LineCommand{"beat", "foo.fie", 8000}}
	if !testEq(cmds, ref) {
		t.Errorf("Failed %v %v", cmds, ref)
	}
}

func TestServiceWithUnderscore(t *testing.T) {
	cmds := Parse([]byte("foo_bar.beat:1|c"))
	ref := []LineCommand{LineCommand{"beat", "foo_bar", 1000}}
	if !testEq(cmds, ref) {
		t.Errorf("Failed %v %v", cmds, ref)
	}
}

func TestIgnoreBadServiceName(t *testing.T) {
	cmds := Parse([]byte("foo/bar.beat:1|c"))
	if len(cmds) != 0 {
		t.Errorf("Failed")
	}
}
