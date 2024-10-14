FROM golang:1.22 AS builder

COPY go.mod /pun/
COPY go.sum /pun/
COPY Makefile /pun/
COPY main.go /pun/
WORKDIR /pun
RUN make

FROM scratch
COPY --from=builder /pun/pun /bin/pun
ENTRYPOINT ["/bin/pun"]
