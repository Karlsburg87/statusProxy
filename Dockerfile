# syntax=docker/dockerfile:1
#buildkit :  https://docs.docker.com/develop/develop-images/build_enhancements/#overriding-default-frontends

FROM docker.io/library/golang:1-alpine  AS build-env
WORKDIR /go/src/statusProxy

#Let us cache modules retrieval as they do not change often.
#Better use of cache than go get -d -u
COPY go.mod .
COPY go.sum .
RUN go mod download

#Update certificates
RUN apk --update add ca-certificates

#Get source and build binary
COPY . .

#Need git for Go Get to work. Apline does not have this installed by default
RUN apk --no-cache add git

#Path to main function
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /statusProxy/bin

#Production image - scratch is the smallest possible but Alpine is a good second for bash-like access
FROM scratch
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /statusProxy/bin /bin/statusProxy

#Default root user container envars
ARG PORT="8080"
ARG PROXY_TO
#gcp
ARG PROJECT_ID

ENV PROXY_TO=${PROXY_TO}
#gcp
ENV PROJECT_ID=${PROJECT_ID}

#Expose port for webhook server
EXPOSE ${PORT}

CMD ["/bin/statusProxy"]