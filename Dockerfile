FROM golang:latest as Builder
WORKDIR /app
COPY . /app
RUN GitCommit=$(git rev-parse HEAD) && \
 GitSourceRoot=$(git rev-parse --show-toplevel) && \
 CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-demo-app -ldflags "-X main.GitCommit=$GitCommit -X main.GitSourceRoot=$GitSourceRoot" .

FROM alpine
RUN apk update && apk add ca-certificates
EXPOSE 80
COPY --from=Builder /app/go-demo-app /go-demo-app
CMD ["/go-demo-app"]
