FROM alpine:3.21.3

WORKDIR /usr/src/migrator

# Dependencies
RUN apk add --no-cache curl

# Install wait-for-it
RUN curl -fsSL -o /usr/local/bin/wait-for https://github.com/eficode/wait-for/releases/download/v2.2.4/wait-for
RUN chmod +x /usr/local/bin/wait-for
RUN ls -l /usr/local/bin/wait-for

# Install dbmate
RUN curl -fsSL -o /usr/local/bin/dbmate https://github.com/amacneil/dbmate/releases/latest/download/dbmate-linux-amd64
RUN chmod +x /usr/local/bin/dbmate
RUN ls -l /usr/local/bin/dbmate
RUN rm -rf /var/lib/apt/lists/*

COPY ./db/migrations /usr/src/migrator/db/migrations