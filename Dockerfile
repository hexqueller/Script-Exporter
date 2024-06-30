FROM alpine

WORKDIR /app

COPY configs /app/configs
COPY scripts /app/scripts
COPY exporter .

RUN chmod +x /app/exporter && \
    apk add --no-cache bash

EXPOSE 9105
CMD ["/app/exporter", "-d"]
