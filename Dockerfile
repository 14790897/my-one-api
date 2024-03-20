FROM node:16-slim as builder

WORKDIR /build
COPY web/package.json .
COPY web/yarn.lock .
RUN yarn install --network-timeout 1000000
COPY ./web .
COPY ./VERSION .
RUN DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$(cat VERSION) yarn build

FROM golang:1.19-alpine AS builder2
RUN apk add --no-cache build-base
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux

WORKDIR /build
#ADD go.mod go.sum ./
COPY . .
COPY --from=builder /build/build ./web/build

RUN go mod tidy \
    && go build -ldflags "-s -w -X 'one-api/common.Version=$(cat VERSION)' -extldflags '-static'" -o one-api

FROM alpine

COPY --from=builder2 /build/one-api /
EXPOSE 3000
WORKDIR /data
ENTRYPOINT ["/one-api"]
