package rbg2p

import (
	"regexp"
	"testing"
)

var fsExpGot = "Expected: /%v/ got: /%v/"

//t.Errorf(fsExpGot, expect, result)

// container to compare variables in testing
type tVar struct {
	Name  string
	Value string
}

func TestLoadFile(t *testing.T) {
}

func TestNewtVar(t *testing.T) {
	validLines := map[string]tVar{
		"VAR VOICED_PLOSIVE [dgb]": tVar{Name: "VOICED_PLOSIVE", Value: "[dgb]"},
		"VAR VOWEL [aoiuye]":       tVar{Name: "VOWEL", Value: "[aoiuye]"},
	}
	invalidLines := map[string]tVar{
		"VAR VOICED_PLOSIVE [dgb]": tVar{Name: "VOICED_PLOSIVE", Value: "dgb"},
		"VAR VOWEL [aoiuye]":       tVar{Name: "VOWEL", Value: "[aoiuye"},
	}
	failLines := []string{
		"VAR VOICED_PLOSIVE",
		"VAR VOICED_PLOSIVE [dgb] anka",
	}

	for l, expect := range validLines {
		name, val, err := newVar(l)
		result := tVar{Name: name, Value: val}
		if err != nil {
			t.Errorf("didn't expect error for input var line %s : %s", l, err)
		} else if expect != result {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for l, expect := range invalidLines {
		name, val, err := newVar(l)
		result := tVar{Name: name, Value: val}
		if err != nil {
			t.Errorf("didn't expect error for input var line %s : %s", l, err)
		} else if expect == result {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for _, l := range failLines {
		_, _, err := newVar(l)
		if err == nil {
			t.Errorf("expected error for input var line %s", l)
		}
	}

}

func TestNewTest(t *testing.T) {
	validLines := map[string]Test{
		"TEST anka -> AnkA":            Test{Input: "anka", Output: []string{"AnkA"}},
		"TEST banka -> {bAnkA, bANkA}": Test{Input: "banka", Output: []string{"bAnkA", "bANkA"}},
	}
	invalidLines := map[string]Test{
		"TEST anka -> AnkA":            Test{Input: "anka", Output: []string{"anka"}},
		"TEST banka -> {bAnkA, bANkA}": Test{Input: "banka", Output: []string{"bAnkA", "bANKkA"}},
	}
	failLines := []string{
		"TEST anka",
		"TEST anka AnkA",
		"TEST anka -> AnkA -> ANkA",
	}

	for l, expect := range validLines {
		result, err := newTest(l)
		if err != nil {
			t.Errorf("didn't expect error for input test line %s : %s", l, err)
		} else if !expect.Equals(result) {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for l, expect := range invalidLines {
		result, err := newTest(l)
		if err != nil {
			t.Errorf("didn't expect error for input test line %s : %s", l, err)
		} else if expect.Equals(result) {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for _, l := range failLines {
		_, err := newTest(l)
		if err == nil {
			t.Errorf("expected error for input test line %s", l)
		}
	}

}

func TestNewRule(t *testing.T) {
	vars := map[string]string{
		"VOICED": "[dgjlvbnm]",
	}
	validLines := map[string]Rule{
		"sch -> {x, S} / _ #": Rule{Input: "sch",
			Output:       []string{"x", "S"},
			LeftContext:  Context{},
			RightContext: Context{regexp.MustCompile("$")}},
		"sch -> {x, S}": Rule{Input: "sch",
			Output:       []string{"x", "S"},
			LeftContext:  Context{},
			RightContext: Context{}},
		"a -> A": Rule{Input: "a",
			Output:       []string{"A"},
			LeftContext:  Context{},
			RightContext: Context{}},
		"a -> A / _ VOICED": Rule{Input: "a",
			Output:       []string{"A"},
			LeftContext:  Context{},
			RightContext: Context{regexp.MustCompile("[dgjlvbnm]")}},
	}
	invalidLines := map[string]Rule{}
	failLines := []string{
		"",
	}

	for l, expect := range validLines {
		result, err := newRule(l, vars)
		if err != nil {
			t.Errorf("didn't expect error for input rule line %s : %s", l, err)
		} else if !expect.Equals(result) {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for l, expect := range invalidLines {
		result, err := newRule(l, vars)
		if err != nil {
			t.Errorf("didn't expect error for input rule line %s : %s", l, err)
		} else if expect.Equals(result) {
			t.Errorf(fsExpGot, expect, result)
		}
	}

	for _, l := range failLines {
		_, err := newRule(l, vars)
		if err == nil {
			t.Errorf("expected error for input rule line %s", l)
		}
	}

}
