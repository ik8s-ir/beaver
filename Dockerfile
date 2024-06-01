FROM registry.ik8s.ir/golang:latest as builder
ENV HOME /
ENV CGO_ENABLED 0
ENV GOOS linux
WORKDIR /
COPY . .
RUN go get -d && go mod download && go build -a -ldflags "-s -w" -installsuffix cgo -o beaver .

FROM registry.ik8s.ir/alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /
COPY --from=builder /beaver .
CMD ["./beaver"]
