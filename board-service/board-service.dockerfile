FROM golang:1.19-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLE=0 go build -o boardApp .

RUN chmod +x /app/boardApp

FROM alpine:latest
RUN mkdir /app
COPY --from=builder /app/boardApp /app
COPY --from=builder /app/img /app/img
CMD [ "/app/boardApp" ]