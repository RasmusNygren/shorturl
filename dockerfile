FROM alpine:latest

RUN apk add -v build-base
RUN apk add -v go 
RUN apk add -v ca-certificates
RUN apk add --no-cache \
    unzip \
    openssh

COPY ./shorturl /pb
WORKDIR /pb

RUN go build
WORKDIR /

EXPOSE 8090


CMD ["/pb/shorturl", "serve", "--http=0.0.0.0:8090"]
