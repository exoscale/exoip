package exoip

import (
	"reflect"
	"testing"
)

func TestStrToUUID(t *testing.T) {
	s := "1128bd56-b4d9-4ac6-a7b9-c715b187ce11"
	u, err := StrToUUID(s)
	if err != nil {
		t.Fatal(err)
	}
	expected := []byte{0x11, 0x28, 0xbd, 0x56, 0xb4, 0xd9, 0x4a, 0xc6, 0xa7, 0xb9, 0xc7, 0x15, 0xb1, 0x87, 0xce, 0x11}
	if !reflect.DeepEqual(expected, u) {
		t.Errorf("%v != %v", expected, u)
	}
}

func TestUUIDToStrTo(t *testing.T) {
	u := []byte{0x11, 0x28, 0xbd, 0x56, 0xb4, 0xd9, 0x4a, 0xc6, 0xa7, 0xb9, 0xc7, 0x15, 0xb1, 0x87, 0xce, 0x11}
	s, err := UUIDToStr(u)
	if err != nil {
		t.Fatal(err)
	}

	expected := "1128bd56-b4d9-4ac6-a7b9-c715b187ce11"
	if expected != s {
		t.Errorf("%s != %s", expected, s)
	}
}
