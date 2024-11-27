FROM golang:1.23-alpine
RUN apk add --no-cache git make build-base
WORKDIR /go/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . ./
RUN CGO_ENABLED=0 go build -a -o app .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=0 /go/app ./
CMD ["./app"]