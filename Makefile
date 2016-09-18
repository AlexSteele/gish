
all: main.go
	go build -o gish main.go

clean:
	rm -f gish
