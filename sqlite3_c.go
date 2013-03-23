package sqlite

/*
#cgo CFLAGS: -I.
#cgo windows CFLAGS:  -fno-stack-check -fno-stack-protector -mno-stack-arg-probe
#cgo windows LDFLAGS: -lmingwex -lmingw32
#cgo linux LDFLAGS: -dl
#cgo freebsd LDFLAGS: -dl
#cgo netbsd LDFLAGS: -dl
#cgo openbsd LDFLAGS: -dl
*/
import "C"
