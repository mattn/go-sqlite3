package sqlite

/*
#cgo CFLAGS: -I.
#cgo windows CFLAGS:  -fno-stack-check -fno-stack-protector -mno-stack-arg-probe
#cgo windows LDFLAGS: -lmingwex -lmingw32
#cgo linux LDFLAGS: -ldl
#cgo freebsd LDFLAGS: -ldl
#cgo netbsd LDFLAGS: -ldl
#cgo openbsd LDFLAGS: -ldl
*/
import "C"
