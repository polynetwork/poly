FROM debian:bullseye-slim
RUN apt-get update && apt-get install -y unzip wget curl build-essential
RUN curl -L https://golang.org/dl/go1.20.linux-`dpkg --print-architecture`.tar.gz | tar -C /usr/local -xzf -

WORKDIR /workspace

ARG commit=master

RUN PATH="$PATH:/usr/local/go/bin" git clone https://github.com/polynetwork/poly.git  && \
    cd poly && git checkout ${commit} && make poly
