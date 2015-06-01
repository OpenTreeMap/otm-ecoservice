.PHONY: test clean build release

test:
	godep go test eco/*

clean:
	rm -rf ecoservice/ \
		   Godeps/_workspace/src/github.com/OpenTreeMap/ecoservice/ \
		   ecoservice.tar.gz

build: clean
	mkdir -p Godeps/_workspace/src/github.com/OpenTreeMap/ecoservice/
	cp -r eco/ Godeps/_workspace/src/github.com/OpenTreeMap/ecoservice/
	cp -r ecorest/ Godeps/_workspace/src/github.com/OpenTreeMap/ecoservice/
	mkdir ecoservice
	godep go build -o ecoservice/ecobenefits

release: build
	cp -r data/ ecoservice/data/
	cp config.gcfg.template ecoservice/
	tar czf ecoservice.tar.gz ecoservice/
