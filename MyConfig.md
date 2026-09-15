# Hướng dẫn cài đặt và sử dụng Docker để chạy ứng dụng

## 1. Cài đặt Docker Desktop

- Link tải: https://www.docker.com/products/docker-desktop/
- Cài đặt bằng winget (Windows): `winget install Docker.sbx`

## 2. File config

- docker-compose.yml: 
```yaml
version: '3.8'
services:
  redis:
    image: redis:7
    container_name: redis
    ports:
      - "6379:6379"
  mongo:
    image: mongo:7
    container_name: mongo
    restart: always
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: password123
    volumes:
      - mongo_data:/data/db

volumes:
  mongo_data:
```

## 3. Chạy ứng dụng

```bash
docker-compose up -d
```