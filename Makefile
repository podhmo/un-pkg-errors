check:
	mkdir -p _examples/foo_transformed
	go run . -debug _examples/foo/foo.go > _examples/foo_transformed/foo.go
	diff -u _examples/foo/foo.go  _examples/foo_transformed/foo.go | tee _examples/foo.diff
	go run . -debug _examples/foo/is_test.go > _examples/foo_transformed/is_test.go
	diff -u _examples/foo/is_test.go  _examples/foo_transformed/is_test.go | tee _examples/is_test.diff

run:
	go run . -debug _examples/foo/foo.go