FROM golang:1.20-bullseye
WORKDIR /workspace

RUN git clone https://github.com/polynetwork/poly.git  && \
    cd poly && \
    make poly
