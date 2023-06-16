TARGET = sipserver

all: clean build

build:
	GOOS=linux go build -v -o $(TARGET)

clean:
	rm $(TARGET)