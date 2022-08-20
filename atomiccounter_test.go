package atomiccounter_test

import (
	"testing"

	"github.com/chen3feng/atomiccounter"
)

func TestInt64_Load(t *testing.T) {
	count := atomiccounter.MakeInt64()
	if count.Read() != 0 {
		t.Fail()
	}
}

func TestInt64_Add(t *testing.T) {
	count := atomiccounter.MakeInt64()
	count.Add(2)
	if count.Read() != 2 {
		t.Fail()
	}
}

func TestInt64_Inc(t *testing.T) {
	count := atomiccounter.MakeInt64()
	count.Inc()
	if count.Read() != 1 {
		t.Fail()
	}
}

func TestInt64_Set(t *testing.T) {
	count := atomiccounter.MakeInt64()
	count.Set(10)
	if count.Read() != 10 {
		t.Fail()
	}
}

func TestInt64_Swap(t *testing.T) {
	count := atomiccounter.MakeInt64()
	count.Set(1)
	n := count.Swap(10)
	if n != 1 {
		t.Fail()
	}
	if count.Read() != 10 {
		t.Fail()
	}
}

func TestMakeInt64(t *testing.T) {
	for i := 0; i < 10000; i++ {
		count := atomiccounter.MakeInt64()
		count.Set(1)
		n := count.Swap(10)
		if n != 1 {
			t.Fail()
		}
		if count.Read() != 10 {
			t.Fail()
		}
	}
}
