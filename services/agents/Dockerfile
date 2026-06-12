FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . /src
RUN cd services/agents && go build -o /out/agents ./cmd/agents

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/agents /agents
USER nonroot
ENTRYPOINT ["/agents"]
