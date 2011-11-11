include $(GOROOT)/src/Make.inc

TARG     = github.com/mattn/go-sqlite3
CGOFILES = sqlite3.go

include $(GOROOT)/src/Make.pkg
