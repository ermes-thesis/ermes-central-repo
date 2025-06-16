FROM golang:1.24
LABEL authors="lucat"

RUN apt-get update

ENV GOPRIVATE=github.com/ermes-thesis
RUN git config --global url."https://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@github.com/ermes-thesis/ermes-wasm-pool".insteadOf "https://github.com/ermes-thesis/ermes-wasm-pool"

RUN curl -sSf https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install.sh | bash -s -- -p /usr/local