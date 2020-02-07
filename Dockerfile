FROM golang:latest as Builder
WORKDIR /app
COPY . /app
RUN export SCOPE_COMMIT_SHA=$(git rev-parse HEAD) && \
 export SCOPE_SOURCE_ROOT=$(git rev-parse --show-toplevel) && \
 export && \
 CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-demo-app  .

FROM alpine
RUN apk update && apk add ca-certificates
EXPOSE 80
COPY --from=Builder /app/go-demo-app /go-demo-app
CMD ["/go-demo-app"]
