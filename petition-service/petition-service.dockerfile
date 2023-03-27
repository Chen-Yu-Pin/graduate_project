FROM golang:1.19-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLE=0 go build -o petitionApp .

RUN chmod +x /app/petitionApp

FROM alpine:latest
RUN mkdir /app
COPY --from=builder /app/petitionApp /app

CMD [ "/app/petitionApp" ]