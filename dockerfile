FROM golang:1.22.0-alpine

RUN apk add --no-cache gcc musl-dev
RUN addgroup -S mercari && adduser -S trainee -G mercari
#RUN chown -R trainee:mercari /app/db

WORKDIR /app

COPY go/go.mod .
COPY go/go.sum .
RUN go mod download

COPY ../db/ /app/db/  
COPY ../db/mercari.sqlite3 /app/db/

COPY go/app/main.go ./

RUN CGO_ENABLED=1 GOOS=linux go build -o /app/mywebserver


EXPOSE 9000

CMD ["/app/mywebserver"]
