FROM golang:1.17

LABEL maintainer="Michael Macnair"
LABEL description="Contains all of the utilities required to build and test the project"

ENV NODEV="16.14.0" \
    GOSWAGGERV="0.29.0" \
    GOLANGCILINTV="1.44.2" \
    GCPSDKV="373.0.0" \
    DOTNETV="2.1"

# apt repo for dotnet
RUN wget -q https://packages.microsoft.com/config/ubuntu/21.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb && \
    dpkg -i packages-microsoft-prod.deb && \
    rm packages-microsoft-prod.deb

# jre for gcloud emulators
# dotnet for nswag
# jq for deploy scripts
# parallel for job performance
# python for yq
RUN apt-get update && apt-get install --no-install-recommends -y \
    default-jre-headless \
    dotnet-sdk-${DOTNETV} \
    httpie \
    jq \
    parallel \
    python3-pip \
    python3-setuptools \
    python3-wheel \
    && rm -rf /var/lib/apt/lists/*

RUN http --ignore-stdin https://nodejs.org/dist/v${NODEV}/node-v${NODEV}-linux-x64.tar.gz | \
    tar xz -C "/opt"
ENV PATH="${PATH}:/opt/node-v${NODEV}-linux-x64/bin"

RUN http --ignore-stdin --check-status --download -o /usr/local/bin/goswagger "https://github.com/go-swagger/go-swagger/releases/download/v${GOSWAGGERV}/swagger_linux_amd64" && \
    chmod +x /usr/local/bin/goswagger

RUN http --ignore-stdin https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
    sh -s -- -b "$(go env GOPATH)/bin" v${GOLANGCILINTV}

RUN go get -u golang.org/x/lint/golint

RUN http --ignore-stdin https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCPSDKV}-linux-x86_64.tar.gz | \
    tar xz -C "/opt"
ENV PATH="${PATH}:/opt/google-cloud-sdk/bin"
RUN gcloud components install beta cloud-firestore-emulator

RUN pip3 install --no-cache-dir yq

RUN printf "\
    $(go version)\\n\
    node: $(node -v), npm: $(npm -v)\\n\
    goswagger: $(goswagger version | head -n1)\\n\
    $(golangci-lint --version)\\n\
    $(gcloud -v)\\n\
    $(yq --version)\\n"
