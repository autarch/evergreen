FROM golang:1.19.3
ENV GOPATH=
ADD . .
RUN make clean && make build
ENV MONGO_URL="mongodb://mongo:27017"
ENTRYPOINT DBHOST=mongo make local-evergreen
