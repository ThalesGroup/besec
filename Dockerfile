FROM alpine:latest as prep

LABEL maintainer="Michael Macnair"

# This Dockerfile is purely for deployment.
# If you want to build the project in Docker, see build.Dockerfile

RUN apk --no-cache add ca-certificates
WORKDIR /prep
RUN echo "nobody:*:1000:1000:nobody:/:/eof" > passwd
RUN echo "nogroup::1000:1000" > group

FROM scratch
COPY --from=prep /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=prep /prep/passwd /etc/passwd
COPY --from=prep /prep/group /etc/group

COPY ./besec /besec

ENV PATH=/

USER nobody
EXPOSE 8080
ENTRYPOINT ["besec"]
CMD ["serve"]
