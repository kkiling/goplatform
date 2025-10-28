.PHONY: bin-deps
bin-deps:
	go install github.com/pav5000/smartimports/cmd/smartimports@v0.2.0

.PHONY: format
format:
	@echo "\n --- 🚀 Start format imports --- \n"
	smartimports -local "github.com/kkiling/goplatform/" -path . -exclude "*_mock.go"