package foo

import (
	"fmt"
	"strings"
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

type MultiError []error

func (es MultiError) Error() string {
	errs := ([]error)(es)
	xs := make([]string, len(errs))
	for i, err := range errs {
			xs[i] = err.Error()
	}
	return fmt.Sprintf("multi error %s", strings.Join(xs, "\n"))
}

func TestAs(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		oops := func() error {
			return MultiError([]error{
				errors.New("oyoyo"),
				errors.Errorf("hmm %d", 0),
			})
		}
		var want MultiError
		if got := oops(); !errors.As(got, &want) {
			t.Errorf("unexpected: %v != %v", want, got)
		}
	})
	t.Run("ng", func(t *testing.T) {
		oops := func() error {
			return errors.New("oyoyo")
		}
		var want MultiError
		if got := oops(); errors.As(got, &want) {
			t.Errorf("unexpected: %v == %v", want, got)
		}
	})
}
