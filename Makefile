.PHONY: test clean build release

test:
	godep go test eco/*

clean:
	rm -rf ecoservice/ \
		   src/github.com/OpenTreeMap/otm-ecoservice/ \
		   ecoservice.tar.gz

build: clean
	mkdir -p src/github.com/OpenTreeMap/otm-ecoservice/
	cp -r eco/ src/github.com/OpenTreeMap/otm-ecoservice/
	cp -r ecorest/ src/github.com/OpenTreeMap/otm-ecoservice/
	mkdir ecoservice
	go build -o ecoservice/ecobenefits

release: build
	cp -r data/ ecoservice/data/
	tar czf ecoservice.tar.gz ecoservice/
