FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . /src
RUN cd services/flow && go build -o /out/flow ./cmd/flow

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/flow /flow
USER nonroot
ENTRYPOINT ["/flow"]
