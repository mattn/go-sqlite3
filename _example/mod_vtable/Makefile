ifeq ($(OS),Windows_NT)
EXE=extension.exe
LIB_EXT=dll
RM=cmd /c del
LIBCURL=-lcurldll
LDFLAG=
else
EXE=extension
ifeq ($(shell uname -s),Darwin)
LIB_EXT=dylib
else
LIB_EXT=so
endif
RM=rm -f
LDFLAG=-fPIC
LIBCURL=-lcurl
endif
LIB=sqlite3_mod_vtable.$(LIB_EXT)

all : $(EXE) $(LIB)

$(EXE) : extension.go
	go build $<

$(LIB) : sqlite3_mod_vtable.cc
	g++ $(LDFLAG) -shared -o $@ $< -lsqlite3 $(LIBCURL)

clean :
	@-$(RM) $(EXE) $(LIB)
