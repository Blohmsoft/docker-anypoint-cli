FROM node:8.16.1-alpine

RUN apk update \
	&& apk upgrade \
	&& apk add bash \
	&& apk add git

RUN npm install -g anypoint-cli@3.2.6

CMD /bin/bash
