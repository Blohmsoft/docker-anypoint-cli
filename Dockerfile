FROM golang:1.13.3-alpine3.10 AS builder

RUN mkdir /mvn-download
WORKDIR /mvn-download

COPY mvn-download/* .

RUN go build

FROM node:8.16.1-alpine

RUN apk update \
	&& apk upgrade \
	&& apk add bash \
	&& apk add git \
	&& apk add zip \
	&& apk add perl-xml-xpath

RUN npm install -g anypoint-cli@3.2.6

COPY --from=builder /mvn-download/mvn-download /bin/mvn-download

CMD /bin/bash
