package advancedtechniques

import (
	"testing"

	"github.com/Rhymond/go-money"
)

func FuzzCurrency(f *testing.F) {
	f.Fuzz(func(t *testing.T, currencyAmount int64, splitAmount int) {
		amount := money.New(currencyAmount, money.GBP)

		split, err := amount.Split(splitAmount)
		if err != nil {
			return
		}

		final := money.New(0, money.GBP)
		for _, s := range split {
			final, err = final.Add(s)
			if err != nil {
				// This shouldn't happen
				t.Fatal("Error adding split currency back together", final, s)
			}
		}

		eq, err := amount.Equals(final)
		if err != nil {
			t.Fatal("Error when comparing currency values that should be valid", err)
		}

		if !eq {
			t.Fatal("Splitting currency into", splitAmount, "parts, and adding back together, produces mismatch")
		}
	})
}
