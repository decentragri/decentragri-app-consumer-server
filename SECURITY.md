# Security Guidelines for Decentragri App CX Server

## Overview
This document outlines the security measures implemented in the Decentragri App CX Server and best practices for maintaining security.

## Implemented Security Measures

### 1. Input Validation & Sanitization
- **All user inputs are validated** using custom validation functions
- **SQL injection prevention** through parameterized queries and input sanitization
- **XSS protection** via input sanitization and output encoding
- **Path traversal protection** in file handling operations

### 2. Authentication & Authorization
- **JWT-based authentication** with secure token handling
- **Token validation** on all protected endpoints
- **Secure token storage** practices
- **Role-based access control** where applicable

### 3. Rate Limiting
- **Global rate limiting**: 100 requests per 15 minutes per IP
- **API-specific rate limiting**: 50 requests per 10 minutes per IP+User-Agent
- **Sliding window implementation** for accurate rate limiting
- **Custom rate limit responses** with appropriate error codes

### 4. Security Headers
- **XSS Protection**: `X-XSS-Protection: 1; mode=block`
- **Content Type Sniffing**: `X-Content-Type-Options: nosniff`
- **Frame Options**: `X-Frame-Options: DENY`
- **Referrer Policy**: `Referrer-Policy: no-referrer`
- **Cross-Origin Policies**: Properly configured CORS

### 5. Error Handling
- **Sanitized error responses** prevent information disclosure
- **Internal error logging** for debugging without exposing details
- **Custom error handlers** for different error types
- **Stack trace protection** in production

### 6. Environment Security
- **Environment-driven configuration** for all sensitive settings
- **Secret management** through environment variables
- **CORS configuration** based on environment
- **Production vs development** configurations

## Security Best Practices

### 1. Environment Variables
```bash
# Never commit .env files
# Use different values for development/staging/production
# Rotate secrets regularly
JWT_SECRET=your-super-secret-jwt-signing-key
SECRET_KEY=your-thirdweb-secret-key
```

### 2. CORS Configuration
```bash
# Development
CORS_ORIGINS=http://localhost:3000,http://localhost:3001

# Production
CORS_ORIGINS=https://yourdomain.com,https://app.yourdomain.com
```

### 3. Rate Limiting
- Monitor rate limit violations
- Adjust limits based on usage patterns
- Consider user-specific rate limits for authenticated users

### 4. Logging Security
- Never log sensitive data (passwords, tokens, API keys)
- Log security events (failed logins, rate limit violations)
- Use structured logging in production
- Monitor logs for suspicious activity

### 5. Input Validation
```go
// Example of proper input validation
func validateInput(input string) error {
    if len(input) > 100 {
        return errors.New("input too long")
    }
    if !utils.ValidationRules.SafeString.MatchString(input) {
        return errors.New("invalid characters")
    }
    return nil
}
```

## Security Checklist

### Before Deployment
- [ ] All environment variables configured
- [ ] .env files not committed to repository
- [ ] CORS origins restricted to production domains
- [ ] Rate limiting configured appropriately
- [ ] Error messages sanitized
- [ ] Debug logging disabled in production
- [ ] Security headers enabled
- [ ] JWT secrets rotated
- [ ] Database credentials secured

### Regular Maintenance
- [ ] Monitor security logs
- [ ] Update dependencies regularly
- [ ] Review and rotate secrets
- [ ] Test rate limiting effectiveness
- [ ] Validate input sanitization
- [ ] Review CORS configuration
- [ ] Check for new security vulnerabilities

## Incident Response

### If Security Breach Detected
1. **Immediate**: Rotate all secrets and API keys
2. **Assess**: Determine scope and impact
3. **Contain**: Block malicious IPs if identified
4. **Investigate**: Review logs for attack vectors
5. **Patch**: Fix vulnerabilities identified
6. **Monitor**: Enhanced monitoring post-incident

### Emergency Contacts
- Security Team: [security@decentragri.com]
- DevOps Team: [devops@decentragri.com]
- Infrastructure: [infrastructure@decentragri.com]

## Security Tools & Dependencies

### Middleware
- `fiber/v2/middleware/helmet` - Security headers
- `fiber/v2/middleware/limiter` - Rate limiting
- `fiber/v2/middleware/cors` - CORS protection
- `fiber/v2/middleware/recover` - Panic recovery

### Validation
- Custom validation utilities in `/utils/validation.go`
- Input sanitization functions
- Error handling utilities

### Monitoring
- Request logging (development only)
- Error logging with context
- Rate limit violation logging
- Security event logging

## Compliance

This server implements security measures that align with:
- **OWASP Top 10** security recommendations
- **Web Application Security** best practices
- **API Security** standards
- **Data Protection** requirements

## Contact

For security questions or to report vulnerabilities:
- Email: security@decentragri.com
- Create a private issue in the repository
- Contact the development team directly
