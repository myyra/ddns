FROM golang:1.19 as builder

RUN apk add --no-cache ca-certificates git

COPY go.sum go.mod /ddns/

WORKDIR /ddns

RUN go mod download

COPY . /ddns/

RUN CGO_ENABLED=0 go build -mod readonly -o /usr/bin/ddns

FROM scratch

COPY --from=builder /usr/bin/ddns /usr/bin/ddns
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

USER 1001

ENTRYPOINT [ "/usr/bin/ddns" ]
