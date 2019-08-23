FROM golang:1.12 as build

WORKDIR /go/src/app
ADD . /go/src/app

ENV GOPROXY https://proxy.golang.org
RUN go get -d -v ./...
RUN go build -v -o /go/bin/app

# Now copy it into our base image.
FROM gcr.io/distroless/base
COPY --from=build /go/bin/app /
CMD ["/app"]

