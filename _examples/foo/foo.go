package foo

import (
	"fmt"

	"github.com/pkg/errors"
)

func Foo() error {
	if err := Bar(); err != nil {
		return errors.Wrap(err, "foo --")
	}
	return nil
}

func Bar() error {
	if err := Boo(); err != nil {
		return errors.Wrapf(err, "on %s%d", "bar", 0)
	}
	return nil
}
func Boo() error {
	return fmt.Errorf("hmm")
}

func Multiples() (bool, error) {
	if err := Bar(); err != nil {
		return false, errors.Wrap(err, "multiples")
	}
	return true, nil
}

func NeverChanges() error {
	if err := Bar(); err != nil {
		return fmt.Errorf("foo -- %w", err)
	}
	return nil
}

func NewError() error {
	newXXXError := func(err error) error {
		fmt.Println("hmm")
		return err
	}
	return newXXXError(errors.Wrap(Boo(), "hmm"))
}
func Paren() error {
	if err := Boo(); err != nil {
		return (errors.Wrapf(err, "on %s%d", "bar", 0))
	}
	return nil
}
func Assign() error {
	if err := Boo(); err != nil {
		err = errors.Wrapf(err, "on %s%d", "bar", 0)
		return err
	}
	return nil
}
func Assign2() error {
	if err := Boo(); err != nil {
		err2 := errors.Wrapf(err, "on %s%d", "bar", 0)
		return err2
	}
	return nil
}
func Call() error {
	if err := Boo(); err != nil {
		fmt.Printf("!!%sn", errors.Wrapf(err, "on %s%d", "bar", 0).Error())
		return nil
	}
	return nil
}

func WithSprintf() error {
	if err := Bar(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("error message with sprintf %d", 0))
	}
	return nil
}

func WithWithMessage() error {
	if err := Bar(); err != nil {
		return errors.WithMessage(err, "multiples")
	}
	return nil
}

func WithWithStack() error {
	if err := Bar(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

var errSuspend = errors.New("suspend")

func WithNew() error {
	if err := Bar(); err != nil {
		return errors.WithStack(errSuspend)
	}
	return errors.New("hmm")
}

var errSuspend2 = errors.Errorf("suspend %d", 2)

func WithErrorf() error {
	if err := Bar(); err != nil {
		return errors.WithStack(errSuspend)
	}
	return errors.Errorf("hmm %d", 2)
}
