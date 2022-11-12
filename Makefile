check:
	mkdir -p _examples/foo_transformed
	go run . -debug _examples/foo/foo.go > _examples/foo_transformed/foo.go
	diff -u _examples/foo/foo.go  _examples/foo_transformed/foo.go | tee _examples/foo.diff
	