package main

import (
	"testing"
)

func TestRandStringRunes(t *testing.T) {
	s := RandStringRunes(5)
	if len(s) != 5 {
		t.Errorf("Generated string supposed to be of the length 5.")
	}
}
