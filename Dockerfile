FROM golang:1.22 AS builder

WORKDIR /go/src/github.com/metal-stack/gardener-extension-dns-powerdns

COPY . .

RUN make install

FROM alpine:3.19

WORKDIR /

COPY --from=builder /go/bin/gardener-extension-dns-powerdns /gardener-extension-dns-powerdns

CMD ["/gardener-extension-dns-powerdns"]
