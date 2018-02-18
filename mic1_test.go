package main

import (
	"testing"
        "fmt"
)

func TestUnpack(t *testing.T) {
        var tmp uint32
        fmt.Sscanf("00000000110000000000000000000000", "%b", &tmp)
	ins := Unpack(tmp)
	if ins.RD != 1 {
		t.Errorf("Unpacking instruction failed!")
	}
	if ins.MAR != 1 {
		t.Errorf("Unpacking instruction failed!")
	}
}
