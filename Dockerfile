# Build
FROM golang:1.19-alpine as Build

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o /binary cmd/sportos/main.go

# Deploy
FROM golang:1.19-alpine

COPY --from=Build /binary /binary

EXPOSE 8080 8081 8082

ENTRYPOINT [ "/binary" ,"-db.name","sportos","-db.host","sportos","-db.port","5432","-db.user","sportos","-db.pass","secret","-api.pub.port",":8080","-api.bo.port",":8081","-api.lo.port",":8082","-llev","warn","-cors.enable=true","-audit.enable=true"]