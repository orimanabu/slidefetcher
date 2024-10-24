build:
	go build

prepare0:
	go run slidefetcher.go prepare kubecon-eu-2018 > data/kubecon-eu-2018.json
	go run slidefetcher.go prepare kubecon-eu-2019 > data/kubecon-eu-2019.json
	go run slidefetcher.go prepare kubecon-eu-2020 > data/kubecon-eu-2020.json
	go run slidefetcher.go prepare kubecon-eu-2021 > data/kubecon-eu-2021.json
	go run slidefetcher.go prepare kubecon-eu-2022 > data/kubecon-eu-2022.json
	go run slidefetcher.go prepare kubecon-eu-2023 > data/kubecon-eu-2023.json
	go run slidefetcher.go prepare kubecon-eu-2024 > data/kubecon-eu-2024.json
	go run slidefetcher.go prepare kubecon-na-2018 > data/kubecon-na-2018.json
	go run slidefetcher.go prepare kubecon-na-2019 > data/kubecon-na-2019.json
	go run slidefetcher.go prepare kubecon-na-2020 > data/kubecon-na-2020.json
	go run slidefetcher.go prepare kubecon-na-2021 > data/kubecon-na-2021.json
	go run slidefetcher.go prepare kubecon-na-2022 > data/kubecon-na-2022.json
	go run slidefetcher.go prepare kubecon-na-2023 > data/kubecon-na-2023.json
	go run slidefetcher.go prepare kubecon-na-2024 > data/kubecon-na-2024.json

prepare-all:
	go build
	@echo $(shell ./slidefetcher list)
	@for name in $(shell ./slidefetcher list); do echo "=> $$name"; ./slidefetcher prepare $$name > data/$$name.json; done
	rm ./slidefetcher
