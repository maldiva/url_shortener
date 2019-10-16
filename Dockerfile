from golang:latest AS compiler

COPY ./ go/src/urlshortener
WORKDIR go/src/urlshortener

RUN go build -o /bin/urlshortener ./cmd/urlshortener/main.go

from debian:jessie

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=compiler /bin/urlshortener /bin/urlshortener

CMD ["urlshortener"]