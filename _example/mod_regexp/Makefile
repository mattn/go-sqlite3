ifeq ($(OS),Windows_NT)
EXE=extension.exe
LIB_EXT=dll
RM=cmd /c del
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
endif
LIB=sqlite3_mod_regexp.$(LIB_EXT)

all : $(EXE) $(LIB)

$(EXE) : extension.go
	go build $<

$(LIB) : sqlite3_mod_regexp.c
	gcc $(LDFLAG) -shared -o $@ $< -lsqlite3 -lpcre

clean :
	@-$(RM) $(EXE) $(LIB)
