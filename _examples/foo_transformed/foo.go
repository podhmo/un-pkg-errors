package internal

import (
	"fmt"
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
