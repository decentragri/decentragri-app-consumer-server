# Docker Deployment Guide

This guide explains how to run the Decentragri App CX Server using Docker.

## Quick Start

### Option 1: Using Docker Compose (Recommended)

1. **Clone the repository and navigate to the project directory:**
   ```bash
   cd decentragri-app-cx-server
   ```

2. **Create a `.env` file with your configuration:**
   ```bash
   cp .env.example .env
   # Edit .env with your actual values
   ```

3. **Start the services:**
   ```bash
   docker-compose up -d
   ```

4. **Access the server:**
   - Server: http://localhost:9085
   - Redis: localhost:6379

### Option 2: Using Docker only

1. **Build the image:**
   ```bash
   docker build -t decentragri-server .
   ```

2. **Run the container:**
   ```bash
   docker run -d \
     --name decentragri-server \
     -p 9085:9085 \
     --env-file .env \
     decentragri-server
   ```

## Production Deployment

For production, use the optimized Dockerfile:

```bash
# Build production image
docker build -f Dockerfile.prod -t decentragri-server:prod .

# Run production container
docker run -d \
  --name decentragri-server-prod \
  -p 9085:9085 \
  --env-file .env.production \
  --restart unless-stopped \
  decentragri-server:prod
```

## Environment Variables

The following environment variables are required:

### Core Configuration
- `PORT`: Server port (default: :9085)
- `SECRET_KEY`: JWT secret key
- `JWT_SECRET_KEY`: JWT signing key
- `CLIENT_ID`: Application client ID

### Database Configuration
- `MEMGRAPH_URI`: Memgraph database connection string
- `MEMGRAPH_USERNAME`: Memgraph username
- `MEMGRAPH_PASSWORD`: Memgraph password

### Redis Configuration
- `REDIS_ADDR`: Redis server address (default: localhost:6379)
- `REDIS_PASSWORD`: Redis password (optional)
- `REDIS_DB`: Redis database number (default: 0)

### External Services
- `ENGINE_URI`: Engine service URL
- `ENGINE_ACCESS_TOKEN`: Engine access token
- `ENGINE_ENCRYPTION_PASSWORD`: Engine encryption password
- `DEEPSEEK_API_KEY`: DeepSeek API key
- `OPENAI_API_KEY`: OpenAI API key
- `WEATHER_API_KEY`: Weather API key
- `OKX_API_KEY`: OKX API key
- `GOOGLE_CLIENT_ID`: Google OAuth client ID

### Blockchain Configuration
- `ENGINE_ADMIN_WALLET_ADDRESS`: Admin wallet address
- `SERVER_WALLET_ADDRESS`: Server wallet address
- `SOIL_SCAN_NFT`: Soil scan NFT contract address
- `VAULT_ADMIN_KEY`: Vault admin key
- `VAULT_ACCESS_TOKEN`: Vault access token

### Development
- `DEV_BYPASS_TOKEN`: Development bypass token
- `SALT_ROUNDS`: Password hashing rounds (default: 10)

## Docker Commands

### Build Commands
```bash
# Development build
docker build -t decentragri-server:dev .

# Production build
docker build -f Dockerfile.prod -t decentragri-server:prod .

# Build with specific platform
docker build --platform linux/amd64 -t decentragri-server .
```

### Run Commands
```bash
# Run with docker-compose
docker-compose up -d

# Run standalone container
docker run -d --name decentragri-server -p 9085:9085 --env-file .env decentragri-server

# Run with specific environment
docker run -d --name decentragri-server -p 9085:9085 -e PORT=:9085 -e REDIS_ADDR=redis:6379 decentragri-server

# Run in foreground for debugging
docker run --rm -p 9085:9085 --env-file .env decentragri-server
```

### Management Commands
```bash
# View logs
docker logs decentragri-server
docker-compose logs -f decentragri-server

# Stop services
docker-compose down
docker stop decentragri-server

# Remove containers and volumes
docker-compose down -v
docker rm decentragri-server

# Update and restart
docker-compose pull
docker-compose up -d
```

### Health Check
```bash
# Check container health
docker ps
curl http://localhost:9085/health

# View health check logs
docker inspect --format='{{.State.Health}}' decentragri-server
```

## Troubleshooting

### Common Issues

1. **Port already in use:**
   ```bash
   # Find process using port 9085
   lsof -i :9085
   # Kill the process or use a different port
   docker run -p 9086:9085 decentragri-server
   ```

2. **Environment variables not loaded:**
   ```bash
   # Verify .env file exists and has correct format
   cat .env
   # Check container environment
   docker exec decentragri-server env
   ```

3. **Database connection issues:**
   ```bash
   # Check network connectivity
   docker network ls
   docker network inspect decentragri-network
   ```

4. **Redis connection failed:**
   ```bash
   # Check Redis container
   docker logs redis
   # Test Redis connection
   docker exec redis redis-cli ping
   ```

### Debug Mode

Run the container in debug mode:
```bash
docker run -it --rm \
  --env-file .env \
  -p 9085:9085 \
  decentragri-server /bin/sh
```

## Security Considerations

1. **Never commit sensitive environment files**
2. **Use Docker secrets for production deployments**
3. **Regularly update base images**
4. **Scan images for vulnerabilities:**
   ```bash
   docker scan decentragri-server
   ```

## Performance Optimization

1. **Use multi-stage builds** (Dockerfile.prod)
2. **Leverage Docker layer caching**
3. **Use specific image tags, not 'latest'**
4. **Configure appropriate resource limits:**
   ```yaml
   deploy:
     resources:
       limits:
         memory: 512M
         cpus: '0.5'
   ```

## Monitoring

Add monitoring to your docker-compose.yml:
```yaml
services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
```
