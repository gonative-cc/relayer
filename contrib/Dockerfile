FROM ubuntu:latest


# Install bitcoin depenencies
RUN apt-get update
RUN apt-get install build-essential cmake pkg-config python3 -y
RUN apt-get install libevent-dev libboost-dev -y
RUN apt-get install libsqlite3-dev -y
RUN apt-get install git -y

RUN git clone https://github.com/bitcoin/bitcoin.git

WORKDIR /bitcoin

COPY . /bitcoin


RUN chmod +x ./script.sh

CMD ["./script.sh"]
