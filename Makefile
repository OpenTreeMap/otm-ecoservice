.PHONY: test clean release

test:
	godep go test eco/*

clean:
	rm -rf ecoservice/ \
		   Godeps/_workspace/src/github.com/azavea/ecobenefits/ \
		   ecoservice.tar.gz

release: clean
	mkdir ecoservice

	mkdir -p Godeps/_workspace/src/github.com/azavea/ecobenefits/
	cp -r eco/ Godeps/_workspace/src/github.com/azavea/ecobenefits/

	godep go build -o ecoservice/ecobenefits

	cp -r data/ ecoservice/data/
	cp config.gcfg.template ecoservice/

	tar czf ecoservice.tar.gz ecoservice/
