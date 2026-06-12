FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . /src
RUN cd services/orchestrator && go build -o /out/orchestrator ./cmd/orchestrator

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/orchestrator /orchestrator
USER nonroot
ENTRYPOINT ["/orchestrator"]
