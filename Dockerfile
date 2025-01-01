FROM golang:1.23.3-alpine3.20 as builder

WORKDIR /app


COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/sulfone-boron

FROM alpine:3.20
WORKDIR /app
LABEL cyanprint.name="sulfone-boron"
COPY --from=builder /app/sulfone-boron /app/sulfone-boron

ENTRYPOINT [ "/app/sulfone-boron" ]
CMD [ "start" ]