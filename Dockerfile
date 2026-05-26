FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY . .
RUN go build -o /out/krx-pension-bot ./cmd/bot

FROM alpine:3.20

RUN apk add --no-cache ca-certificates
COPY --from=build /out/krx-pension-bot /usr/local/bin/krx-pension-bot

ENV PORT=8080
EXPOSE 8080

CMD ["krx-pension-bot"]
