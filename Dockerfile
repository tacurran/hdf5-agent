FROM node:20 AS frontend-builder

WORKDIR /frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

FROM golang:latest AS backend-builder

RUN apt-get update && apt-get install -y \
    libhdf5-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o hdf5-agent .

FROM golang:latest

RUN apt-get update && apt-get install -y \
    libhdf5-dev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=backend-builder /app/hdf5-agent /app/hdf5-agent
COPY --from=frontend-builder /frontend/dist /app/public

RUN chmod +x /app/hdf5-agent

EXPOSE 8080

ENV PORT=8080

CMD ["/app/hdf5-agent"]