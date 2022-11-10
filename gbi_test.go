package bedtools

import (
	"testing"
	"reflect"
)

func TestIntersect(t *testing.T) {
	ain := Bed{
		BedEntry{"chr1", 5, 25, []string{"first"}, nil},
		BedEntry{"chr1", 15, 25, []string{"second"}, nil},
		BedEntry{"chr2", 15, 25, []string{"third"}, nil},
	}
	bin := Bed{
		BedEntry{"chr1", 5, 8, []string{"first"}, nil},
		BedEntry{"chr2", 100, 200, []string{"second"}, nil},
	}
	expect := Bed{
		BedEntry{"chr1", 5, 8, []string{"first"}, nil},
	}

	out, err := IntersectBeds(ain, []string{}, bin)
	if err != nil {
		panic(err)
	}

	var outbed Bed
	for entry := range out {
		outbed = append(outbed, entry)
	}

	if !reflect.DeepEqual(outbed, expect) {
		t.Errorf("outbed %v != expect %v", outbed, expect)
	}
}
