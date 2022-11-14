package foo

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

func TestIs(t *testing.T) {
	xxx := fmt.Errorf("xxx")
	if want, got := xxx, xxx; !errors.Is(got, want) {
		t.Errorf("unexpected: %v != %v", want, got)
	}
	if want, got := xxx, xxx; !errors.Is(errors.Wrap(got, "hmm"), want) {
		t.Errorf("unexpected: %v != %v", want, got)
	}

	if want, got := xxx, fmt.Errorf("xxx"); errors.Is(got, want) {
		t.Errorf("unexpected: %v == %v", want, got)
	}
}
