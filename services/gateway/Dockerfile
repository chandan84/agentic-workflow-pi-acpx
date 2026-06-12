FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . /src
RUN cd services/gateway && go build -o /out/gateway ./cmd/gateway

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/gateway /gateway
USER nonroot
ENTRYPOINT ["/gateway"]
