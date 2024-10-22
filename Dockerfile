FROM golang:1.22 AS builder

COPY . /pun

WORKDIR /pun
RUN make

FROM scratch
ARG TARGETARCH
COPY --from=builder /pun/dist/pun_${TARGETARCH} /bin/pun
ENTRYPOINT ["/bin/pun"]

