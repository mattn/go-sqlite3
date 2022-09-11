TARGET = custom_driver_name
ifeq ($(OS),Windows_NT)
TARGET := $(TARGET).exe
endif

all : $(TARGET)

$(TARGET) : main.go
	go build -ldflags="-X 'github.com/mattn/go-sqlite3.driverName=my-sqlite3'"

clean :
	rm -f $(TARGET)
