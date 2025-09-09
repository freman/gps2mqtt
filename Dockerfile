FROM --platform=$BUILDPLATFORM golang:1.25.1-alpine AS build
WORKDIR /src
COPY . .
RUN go mod download
ARG TARGETOS TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/gps2mqtt ./cmd/gps2mqtt

FROM alpine
COPY --from=build /out/gps2mqtt /bin/gps2mqtt
ENTRYPOINT ["/bin/gps2mqtt"]
