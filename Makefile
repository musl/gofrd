BIN := gofrd

.PHONY: all clean run

all: clean run

clean:
	go clean .

$(BIN):
	go build .

gctrace: $(BIN)
	GODEBUG=gctrace=2 ./$(BIN)

run: $(BIN)
	./$(BIN)

