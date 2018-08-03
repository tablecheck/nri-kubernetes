FROM golang:1.10.3-alpine3.8 as builder

RUN apk add --no-cache make git
WORKDIR /go/src/github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e
COPY . .
RUN make compile

# -----------------------------------------------------------------------------

FROM alpine:3.8

WORKDIR /var/db/newrelic-infra/fsi

ENV HELM_LATEST_VERSION="v2.9.1"

RUN apk add --update ca-certificates \
 && apk add --update -t deps wget \
 && wget https://storage.googleapis.com/kubernetes-helm/helm-${HELM_LATEST_VERSION}-linux-amd64.tar.gz \
 && tar -xvf helm-${HELM_LATEST_VERSION}-linux-amd64.tar.gz \
 && mv linux-amd64/helm /usr/local/bin \
 && apk del --purge deps \
 && rm /var/cache/apk/* \
 && rm -f /helm-${HELM_LATEST_VERSION}-linux-amd64.tar.gz

COPY --from=builder /go/src/github.com/newrelic/infra-integrations-beta/integrations/kubernetes/e2e/bin/nr-kubernetes-e2e nr-kubernetes-e2e
COPY bin/nr-kubernetes-e2e nr-kubernetes-e2e
COPY schema schema
COPY charts charts

ENTRYPOINT ["/var/db/newrelic-infra/fsi/nr-kubernetes-e2e"]
