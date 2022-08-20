package atomiccounter_test

import (
	"testing"

	"github.com/chen3feng/atomiccounter"
)

func TestInt64_Load(t *testing.T) {
	count := atomiccounter.NewInt64()
	count.Load()
}

func TestInt64_Add(t *testing.T) {
	count := atomiccounter.NewInt64()
	count.Add(1)
}

func TestInt64_Inc(t *testing.T) {
	count := atomiccounter.NewInt64()
	count.Inc()
}

func TestInt64_Set(t *testing.T) {
	count := atomiccounter.NewInt64()
	count.Set(10)
}
