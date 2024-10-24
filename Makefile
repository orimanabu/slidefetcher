build:
	go build

prepare0:
	# go run slidefetcher.go prepare kccnceu2019 > data/kccnceu2019.json
	# go run slidefetcher.go prepare kccnceu2020 > data/kccnceu2020.json
	go run slidefetcher.go prepare kccnceu2021 > data/kccnceu2021.json
	go run slidefetcher.go prepare kccnceu2022 > data/kccnceu2022.json
	go run slidefetcher.go prepare kccnceu2023 > data/kccnceu2023.json
	go run slidefetcher.go prepare kccnceu2024 > data/kccnceu2024.json
	# go run slidefetcher.go prepare kccncna2019 > data/kccncna2019.json
	# go run slidefetcher.go prepare kccncna2020 > data/kccncna2020.json
	go run slidefetcher.go prepare kccncna2021 > data/kccncna2021.json
	go run slidefetcher.go prepare kccncna2022 > data/kccncna2022.json
	go run slidefetcher.go prepare kccncna2023 > data/kccncna2023.json
	go run slidefetcher.go prepare kccncna2024 > data/kccncna2024.json

prepare-all:
	go build
	@echo $(shell ./slidefetcher list)
	@for name in $(shell ./slidefetcher list); do echo "=> $$name"; ./slidefetcher prepare $$name > data/$$name.json; done
	rm ./slidefetcher
