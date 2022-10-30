FROM tetafro/golang-gcc:latest

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY .env ./
COPY *.go ./
COPY logic ./logic

RUN go build -o /triangular-app

EXPOSE 8080


CMD [ "/triangular-app" ]
