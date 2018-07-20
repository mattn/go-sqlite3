# Get Current Directory
CURDIR = $(shell pwd)

ifeq ($(OS),Windows_NT)
EXE=extension.exe
EXT=sqlite3_mod_vtable.dll
RM=cmd /c del
LIBCURL=-lcurldll
LDFLAG=
else
EXE=extension
EXT=sqlite3_mod_vtable.so
RM=rm
LDFLAG=-fPIC
LIBCURL=-lcurl
endif

# Include source from repository
CFLAGS=-I$(CURDIR)/../../driver

all : $(EXT) $(EXE)

$(EXE) : extension.go
	go build -v -tags=sqlite_vtable $<

$(EXT) : sqlite3_mod_vtable.cc
	g++ $(LDFLAG) $(CFLAGS) -shared -o $@ $< $(LIBCURL)

clean :
	@-$(RM) $(EXE) $(EXT)
