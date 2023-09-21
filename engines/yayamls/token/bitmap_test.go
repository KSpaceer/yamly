package token

import (
	"testing"

	"github.com/KSpaceer/yamly/engines/yayamls/yamlchar"
)

func TestConformationBitmap(t *testing.T) {
	t.Parallel()

	sets := []yamlchar.CharSetType{
		yamlchar.DecimalCharSetType,
		yamlchar.WordCharSetType,
		yamlchar.URICharSetType,
		yamlchar.TagCharSetType,
		yamlchar.AnchorCharSetType,
		yamlchar.PlainSafeCharSetType,
		yamlchar.SingleQuotedCharSetType,
		yamlchar.DoubleQuotedCharSetType,
	}

	t.Run("set true", func(t *testing.T) {
		t.Parallel()
		m := conformationBitmap(0)

		for i := range sets {
			result, ok := m.Get(sets[i])
			if ok {
				t.Errorf("expected result to be absent, but map show presence with value %t", result)
			}
			m = m.Set(sets[i], true)
			result, ok = m.Get(sets[i])
			if !ok {
				t.Errorf("expected result to be present")
			} else if !result {
				t.Errorf("expected result to be true, but got false")
			}
		}
	})

	t.Run("set false", func(t *testing.T) {
		t.Parallel()
		m := conformationBitmap(0)

		for i := range sets {
			result, ok := m.Get(sets[i])
			if ok {
				t.Errorf("expected result to be absent, but map show presence with value %t", result)
			}
			m = m.Set(sets[i], false)
			result, ok = m.Get(sets[i])
			if !ok {
				t.Errorf("expected result to be present")
			} else if result {
				t.Errorf("expected result to be false, but got true")
			}
		}
	})

	t.Run("rewrite value", func(t *testing.T) {
		t.Parallel()
		m := conformationBitmap(0)

		set := sets[2]
		result, ok := m.Get(set)
		if ok {
			t.Errorf("expected result to be absent, but map show presence with value %t", result)
		}
		m = m.Set(set, true)
		result, ok = m.Get(set)
		if !ok {
			t.Errorf("expected result to be present")
		} else if !result {
			t.Errorf("expected result to be true, but got false")
		}
		m = m.Set(set, false)
		result, ok = m.Get(set)
		if !ok {
			t.Errorf("expected result to be present")
		} else if result {
			t.Errorf("expected result to be true, but got false")
		}
	})

	t.Run("multiple assignments", func(t *testing.T) {
		t.Parallel()
		m := conformationBitmap(0)

		sets := sets[1:4]

		for i := range sets {
			result, ok := m.Get(sets[i])
			if ok {
				t.Errorf("expected result to be absent, but map show presence with value %t", result)
			}
		}

		for i := range sets {
			value := i%2 == 0
			m = m.Set(sets[i], value)
		}

		for i := range sets {
			result, ok := m.Get(sets[i])
			if !ok {
				t.Errorf("expected result to be present")
			} else if value := i%2 == 0; result != value {
				t.Errorf("expected result to be %t, but got %t", value, result)
			}
		}
	})
}
