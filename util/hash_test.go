package util

import (
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {
	res := ComputeCapacity(17)
	fmt.Println(res)
}

func TestFnv(t *testing.T) {
	res := Fnv32("1")
	fmt.Println(res)
}
