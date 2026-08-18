package utils

import (
	"reflect"
	"testing"
)

func TestWithoutStringsChecksEveryValue(t *testing.T) {
	values := []string{"pre.sh", "brew.sh", "node.sh"}
	got := WithoutStrings(values, []string{"pre.sh"})
	want := []string{"brew.sh", "node.sh"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("WithoutStrings() = %#v, want %#v", got, want)
	}
}
