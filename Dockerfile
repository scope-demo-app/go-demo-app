FROM golang:latest as Builder
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-demo-app .

FROM scratch
EXPOSE 80
COPY --from=Builder /app/go-demo-app /go-demo-app
CMD ["/go-demo-app"]
