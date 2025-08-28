# Decentragri App CX Server

A high-performance Go-based REST API server for the Decentragri Consumer platform, providing authentication, wallet management, marketplace functionality, and portfolio services for blockchain-based agricultural NFTs.

**Core Technologies:**
- **Backend Development**: Go, Node.js, distributed systems architecture
- **Blockchain Integration**: Smart contracts, ThirdWeb Engine, Web3 APIs  
- **Database Systems**: Graph databases (Memgraph), Redis caching
- **Infrastructure**: Microservices, API design, performance optimizationPerformance**: Utilizes all available CPU cores for optimal performance
- **JWT Authentication**: Secure token-based authentication system
- **Wallet Management**: Create wallets, fetch balances for native and ERC20 tokens
- **NFT Portfolio**: Manage and display farm plot NFTs with image fetching
- **Marketplace**: Browse and purchase farm plot listings
- **Caching**: Redis-based caching for improved performance
- **Database**: Memgraph integration for graph-based data storage
- **Image Processing**: Concurrent image fetching with IPFS gateway support

##  Architecture

### Tech Stack
- **Framework**: Fiber (Express-like web framework for Go)
- **Database**: Memgraph (Graph database)
- **Cache**: Redis
- **Authentication**: JWT tokens
- **Blockchain**: ThirdWeb Engine integration
- **Language**: Go 1.19+

### Project Structure
```
decentragri-app-cx-server/
├── auth.services/          # Authentication service and utilities
├── cache/                  # Redis cache management
├── config/                 # Configuration constants and settings
├── db/                     # Database connection and utilities
├── marketplace.services/   # Marketplace functionality
├── middleware/             # HTTP middleware (auth, logging)
├── portfolio.services/     # Portfolio management
├── routes/                # HTTP route definitions
├── token.services/        # JWT token management
├── utils/                 # Utility functions
├── wallet.services/       # Wallet operations
├── main.go                # Application entry point
└── README.md              # This file
```

##  Installation & Setup

### Prerequisites
- Go 1.19 or higher
- Redis server
- Memgraph database
- ThirdWeb Engine API access

### Environment Variables
Create a `.env` file in the root directory:

```env
# Database
MEMGRAPH_URI=bolt://localhost:7687
MEMGRAPH_USERNAME=your_username
MEMGRAPH_PASSWORD=your_password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password

# JWT
JWT_SECRET_KEY=your_jwt_secret_key

# ThirdWeb
SECRET_KEY=your_thirdweb_secret_key
ENGINE_CLOUD_BASE_URL=https://your-engine-url

# Blockchain
CHAIN=8453  # Base chain ID
FARM_PLOT_CONTRACT_ADDRESS=0x...
DAGRI_CONTRACT_ADDRESS=0x...
```

### Installation Steps

1. **Clone the repository**
   ```bash
   git clone https://github.com/YirenNing24/decentragri-app-cx-server.git
   cd decentragri-app-cx-server
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start Redis and Memgraph services**
   ```bash
   # Redis (example with Docker)
   docker run -d --name redis -p 6379:6379 redis:alpine
   
   # Memgraph (example with Docker)
   docker run -d --name memgraph -p 7687:7687 memgraph/memgraph
   ```

5. **Run the server**
   ```bash
   go run main.go
   ```

The server will start on port `9085` by default.

##  API Endpoints

### Authentication
- `POST /api/auth/login` - User login
- `POST /api/auth/refresh` - Refresh JWT token

### Wallet Operations
- `POST /api/wallet/create` - Create a new smart wallet
- `GET /api/wallet/balances` - Get user's token balances (native + DAGRI)
- `GET /api/wallet/nfts/:contract` - Get owned NFTs from a contract

### Portfolio Management
- `GET /api/portfolio/summary` - Get portfolio summary (NFT count)
- `GET /api/portfolio/entire` - Get complete portfolio with images

### Marketplace
- `GET /api/marketplace/valid-farmplots` - Get all valid farm plot listings
- `GET /api/marketplace/featured-property` - Get featured property
- `POST /api/marketplace/buy-from-listing` - Purchase from marketplace

##  Configuration

### Server Configuration
The server is configured in `main.go` with the following settings:
- **Port**: 9085
- **Body Limit**: 50MB
- **Idle Timeout**: 60 seconds
- **Multi-core**: Utilizes all available CPU cores

### Caching Strategy
- **Images**: Cached for 1 hour
- **Portfolio Data**: Cached for 3 minutes
- **Token Balances**: No caching (real-time data)

### Concurrency Limits
- **Image Fetching**: Maximum 20 concurrent requests per operation
- **API Requests**: No artificial limits (handled by Fiber)

## ecurity Features

- **JWT Authentication**: All protected routes require valid JWT tokens
- **Token Validation**: Automatic token expiry and refresh mechanism
- **Input Validation**: Request validation and sanitization
- **Rate Limiting**: Built-in protection against abuse

##  Performance Optimizations

- **Multi-core Processing**: Utilizes all CPU cores
- **Concurrent Operations**: Goroutines for image fetching and I/O operations
- **Redis Caching**: Reduces database load and improves response times
- **Connection Pooling**: Efficient database connection management
- **Image Optimization**: IPFS gateway integration with caching

## Error Handling

The server implements comprehensive error handling:
- **Structured Logging**: Detailed request/response logging
- **Graceful Failures**: Non-critical failures don't crash the server
- **Error Recovery**: Automatic recovery from transient errors
- **Status Codes**: Proper HTTP status code usage

## Development

### Running in Development Mode
```bash
go run main.go
```

### Building for Production
```bash
go build -o decentragri-server main.go
./decentragri-server
```

### Testing
```bash
go test ./...
```

## 📝 API Response Formats

### Successful Response
```json
{
  "data": {...},
  "status": "success"
}
```

### Error Response
```json
{
  "error": "Error message",
  "status": "error"
}
```

### Wallet Balance Response
```json
{
  "walletAddress": "0x...",
  "native": {
    "balance": "1.23",
    "rawBalance": "1230000000000000000",
    "priceUSD": 2500.50,
    "valueUSD": 3075.615
  },
  "dagri": {
    "balance": "100.0",
    "rawBalance": "100000000000000000000",
    "priceUSD": 0,
    "valueUSD": 0
  },
  "lastUpdated": 1635724800
}
```

## Deployment

### Docker Deployment
```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

### Environment Setup
Ensure all environment variables are properly set in your deployment environment.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Team

- **Development Team**: Decentragri Core Team
- **Blockchain Integration**: ThirdWeb Engine
- **Database**: Memgraph
- **Caching**: Redis

## 📞 Support

For support and questions:
- Create an issue in this repository
- Contact the development team
- Check the documentation for common solutions

---

Built with ❤️ for the future of decentralized agriculture.
