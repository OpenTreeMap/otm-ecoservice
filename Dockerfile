FROM golang:1.11 

RUN apt-get update && apt-get install -y \
    libgeos-dev \
    mercurial

WORKDIR /usr/local/ecoservice
COPY . .

ENV GOPATH=/usr/local/ecoservice

RUN make build
WORKDIR /usr/local/ecoservice/ecoservice

CMD ecobenefits
