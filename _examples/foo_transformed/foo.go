package internal

import (
	"fmt"

	"github.com/pkg/errors"
)

func Foo() error {
	if err := Bar(); err != nil {
		return fmt.Errorf("foo -- %w", err)
	}
	return nil
}

func Bar() error {
	if err := Boo(); err != nil {
		return fmt.Errorf("on %s%d -- %w", "bar", 0, err)
	}
	return nil
}
func Boo() error {
	return fmt.Errorf("hmm")
}

func Multiples() (bool, error) {
	if err := Bar(); err != nil {
		return false, fmt.Errorf("multiples -- %w", err)
	}
	return true, nil
}

func NeverChanges() error {
	if err := Bar(); err != nil {
		return fmt.Errorf("foo -- %w", err)
	}
	return nil
}

func WithSprintf() error {
	if err := Bar(); err != nil {
		return fmt.Errorf(fmt.Sprintf("error message with sprintf %d", 0)+" -- %w", err)
	}
	return nil
}

func WithNewError() error {
	newXXXError := func(err error) error {
		fmt.Println("hmm")
		return err
	}
	return newXXXError(fmt.Errorf("hmm -- %w", Boo()))
}

func WithWithMessage() error {
	if err := Bar(); err != nil {
		return fmt.Errorf("multiples -- %w", err)
	}
	return nil
}

func WithWithStack() error {
	if err := Bar(); err != nil {
		return err
	}
	return nil
}

var errSuspend = fmt.Errorf("suspend")

func WithNew() error {
	if err := Bar(); err != nil {
		return errSuspend
	}
	return fmt.Errorf("hmm")
}

var errSuspend2 = errors.Errorf("suspend %d", 2)

func WithErrorf() error {
	if err := Bar(); err != nil {
		return errSuspend
	}
	return errors.Errorf("hmm %d", 2)
}
