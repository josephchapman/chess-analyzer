# FROM alpine:latest AS certs
# RUN apk --update add ca-certificates

# FROM alpine:latest AS stockfish
# RUN apk --update add \
#       git \
#       make \
#       g++ \
#       && rm -rf /var/cache/apk/*
# RUN TMP=$(mktemp -d /tmp/stockfish.XXXXXX); cd ${TMP} \
# &&  git clone https://github.com/official-stockfish/Stockfish.git \
# &&  cd Stockfish \
# &&  git checkout tags/sf_17 \
# &&  cd src \
# &&  make -j profile-build \
# &&  mv stockfish /usr/local/bin/

# FROM golang:1.23.5 AS build
# WORKDIR /src
# COPY src/* ./
# RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/chess-analyzer .

# FROM scratch
# COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# COPY --from=stockfish /usr/local/bin/stockfish /bin/stockfish
# COPY --from=build /bin/chess-analyzer /usr/local/bin/chess-analyzer
# CMD ["/usr/local/bin/chess-analyzer"]



# stockfish as subprocess failing under scratch. using ubuntu for now

FROM golang:1.23.5 AS build
WORKDIR /src
COPY src/* ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/chess-analyzer .

FROM ubuntu:jammy
RUN apt update \
&&  apt install -y \
      curl
RUN TMP=$(mktemp -d /tmp/stockfish.XXXXXX) \
&&  curl -fsSL -o ${TMP}/stockfish-ubuntu-x86-64.tar https://github.com/official-stockfish/Stockfish/releases/download/sf_17/stockfish-ubuntu-x86-64.tar \
&&  tar -xvf ${TMP}/stockfish-ubuntu-x86-64.tar -C ${TMP}/ \
&&  mv ${TMP}/stockfish/stockfish-ubuntu-x86-64 /usr/local/bin/stockfish
RUN mkdir /var/lib/data/
COPY --from=build /bin/chess-analyzer /usr/local/bin/chess-analyzer
CMD ["/usr/local/bin/chess-analyzer"]