install:
	go install ./cmd/car-scrape/

run:
	car-scrape

runenv:
	CAR_URL="https://example.com/cars?sell_price_from=%d&sell_price_to=%d&page_nr=%d&older_not=7&order_by=3&order_direction=DESC" go run ./cmd/car-scrape

prof:
	curl --output prof.prof "localhost:8080/debug/pprof/profile?seconds=300"

prof2:
	go tool pprof -http localhost:8080 prof.prof
	go tool pprof prof.prof

iw: install
