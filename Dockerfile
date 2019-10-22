from golang:latest AS compiler

COPY ./ go/src/urlshortener
WORKDIR go/src/urlshortener

RUN go build -o /bin/server ./cmd/server/main.go

from debian:jessie

RUN apt-get update && apt-get install -y ca-certificates

COPY --from=compiler /bin/server /bin/server

CMD ["server"]