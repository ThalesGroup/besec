FROM alpine:latest

LABEL maintainer="Michael Macnair"

# This Dockerfile is purely for (debug) deployment.
# If you want to build the project in Docker, see build.Dockerfile

# This debug image is like the main image, but it includes the base Alpine distro too,
# so it contains a shell. This is useful for running it in a pipeline.

RUN apk --no-cache add ca-certificates
ENV PATH="/:${PATH}"

COPY ./besec /besec
COPY ./config.yaml ./config.local.yam[l] /

EXPOSE 8080
ENTRYPOINT ["besec"]
CMD ["serve"]
