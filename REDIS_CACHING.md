# Redis Caching Implementation

This server now includes comprehensive Redis caching to improve performance by reducing redundant API calls and expensive operations.

## Configuration

### Environment Variables
Add these to your `.env` file:
```
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### Starting Redis
Make sure Redis is running locally:
```bash
# Install Redis (if not already installed)
sudo dnf install redis  # Fedora/RHEL/CentOS
# or
sudo apt install redis-server  # Ubuntu/Debian

# Start Redis service
sudo systemctl start redis
sudo systemctl enable redis  # Auto-start on boot
```

## Caching Implementation

### Services with Caching

1. **Marketplace Service** (`marketplace.services/utils.go`)
   - `GetAllValidFarmPlotListings()`: 5-minute cache
   - `FetchImageBytes()`: 1-hour cache (uses MD5 hash of URL as key)

2. **Wallet Service** (`wallet.services/wallet.service.go`)
   - `GetTokenPriceUSD()`: 2-minute cache for price data

3. **Portfolio Service** (`portfolio.services/portfolio.service.go`)
   - `GetPortfolioSummary()`: 3-minute cache for user portfolio data

### Cache Key Patterns

- Farm listings: `"farm_listings"`
- Image data: `"image:{md5_hash_of_url}"`
- Token prices: `"price:{chainID}:{tokenAddress}"`
- Portfolio data: `"portfolio:{userID}"`

### Cache Behavior

- **Cache Hit**: Returns data directly from Redis
- **Cache Miss**: Fetches from API, stores in cache, then returns data
- **Error Handling**: If Redis is unavailable, falls back to direct API calls

## Performance Benefits

- Reduced API calls to external services
- Faster response times for frequently requested data
- Lower bandwidth usage
- Improved server responsiveness under load

## Monitoring

Check Redis usage:
```bash
redis-cli info memory
redis-cli dbsize
redis-cli monitor  # Watch real-time commands
```

View cache keys:
```bash
redis-cli keys "*"
```

Clear all cache:
```bash
redis-cli flushdb
```
