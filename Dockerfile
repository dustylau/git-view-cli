FROM golang:1.27.4-trixie@sha256:f18c3d73a184fb0582828510722e723b63a93f7188d11dd223d9be6f6880e65d

WORKDIR /src
ENV GOTOOLCHAIN=local \
    GOFLAGS=-buildvcs=false \
    CGO_ENABLED=1

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY cmd/ cmd/
COPY internal/ internal/
COPY tests/ tests/
RUN go build -mod=readonly -o /usr/local/bin/git-view ./cmd/git-view

CMD ["git-view", "--help"]
