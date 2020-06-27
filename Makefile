
# make
all: build package
build: 
	GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o words

package: 
	tar -czf jared_warren_word_count.tgz main.go

