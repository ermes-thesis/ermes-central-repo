FROM golang:1.24
LABEL authors="lucat"

RUN apt-get update

ENV GOPRIVATE=github.com/ermes-thesis
RUN git config --global url."https://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@github.com/ermes-thesis/ermes-wasm-pool".insteadOf "https://github.com/ermes-thesis/ermes-wasm-pool"

RUN curl -sSf https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install.sh | bash -s -- -p /usr/local