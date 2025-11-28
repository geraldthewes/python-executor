# Security Considerations

python-executor is designed for secure execution of untrusted code. However, no sandbox is perfect. This document outlines the security measures and their limitations.

## Security Layers

### 1. Docker Container Isolation

Each execution runs in a separate Docker container that is destroyed after completion.

**Configuration:**
- Read-only root filesystem
- Temporary `tmpfs` mounts for `/work` and `/tmp`
- User namespace: non-root user (UID 1000)

### 2. Network Isolation

By default, containers have no network access (`--network none`).

**To enable network:**
```json
{
  "config": {
    "network_disabled": false
  }
}
```

**Warning:** Enabling network access allows the code to:
- Make external HTTP requests
- Download malicious code
- Exfiltrate data

Only enable network when necessary and from trusted sources.

### 3. Resource Limits

All executions are subject to strict resource limits:

| Resource | Limit | Enforcement |
|----------|-------|-------------|
| Memory | Configurable (default 1GB) | Docker memory limit |
| Disk | Configurable (default 2GB) | tmpfs size limit |
| CPU | CPU shares (default 1024) | Docker CPU quota |
| Time | Configurable (default 5min) | Context timeout |

**OOM behavior:** If a process exceeds memory limits, Docker kills it with exit code 137.

### 4. Filesystem Isolation

- Work directory is on tmpfs (not persistent)
- Root filesystem is read-only
- No access to host filesystem (except via Docker socket, see below)

### 5. Path Traversal Protection

Tar archives are validated to prevent path traversal attacks:

```go
// Rejected paths:
"../etc/passwd"
"/etc/passwd"
"foo/../../../secret"
```

## Known Limitations

### Docker-in-Docker Requirement

The server requires access to the Docker socket and must run in privileged mode for Docker-in-Docker. This means:

1. **The server itself must be trusted** - it has root-level access to the host via Docker socket
2. **Container escape is theoretically possible** - sophisticated attackers could potentially escape the Docker sandbox

### Mitigation Strategies

1. **Network Isolation**
   - Run python-executor on an isolated network segment
   - Use firewall rules to restrict outbound access

2. **Authentication & Authorization**
   - python-executor v1.0 does not include built-in auth
   - Deploy behind a reverse proxy (nginx, Traefik) with authentication
   - Use API gateway for rate limiting and access control

3. **Monitoring**
   - Monitor resource usage
   - Set up alerts for suspicious activity
   - Log all execution requests

4. **Resource Quotas**
   - Set conservative default limits
   - Enforce quotas at the API layer
   - Limit concurrent executions

## Deployment Recommendations

### ✅ Recommended Use Cases

- Internal tools and automation
- Trusted user environments (e.g., internal dev tools)
- Educational platforms with proper rate limiting
- AI agent sandboxes with monitored environments

### ⚠️ Use with Caution

- Public-facing services (add authentication)
- Multi-tenant environments (implement isolation)
- Processing code from unknown sources (strict limits)

### ❌ Not Recommended

- Processing code from completely untrusted sources without additional safeguards
- Environments where container escape would be catastrophic
- Without authentication and rate limiting

## Security Best Practices

### 1. Authentication

Deploy behind an authenticating reverse proxy:

```nginx
location /api/ {
    auth_basic "python-executor";
    auth_basic_user_file /etc/nginx/.htpasswd;
    proxy_pass http://python-executor:8080;
}
```

### 2. Rate Limiting

Prevent abuse with rate limiting:

```nginx
limit_req_zone $binary_remote_addr zone=executions:10m rate=10r/m;

location /api/v1/exec/ {
    limit_req zone=executions burst=5;
    proxy_pass http://python-executor:8080;
}
```

### 3. Input Validation

Validate inputs before submission:
- Maximum tar archive size
- Maximum execution time
- Whitelist of allowed Docker images
- Validate metadata structure

### 4. Monitoring

Monitor for suspicious patterns:
- High resource usage
- Repeated failures
- Unusual execution times
- Network activity (if network is enabled)

### 5. Principle of Least Privilege

- Run server with minimal permissions
- Use network isolation
- Disable network by default
- Use minimal Docker base images

## Incident Response

If you suspect a security issue:

1. **Isolate** - Stop the service immediately
2. **Investigate** - Review logs and execution history
3. **Report** - File a security issue on GitHub
4. **Update** - Apply security patches promptly

## Security Updates

Subscribe to security advisories:
- GitHub repository: Watch for security releases
- Follow @python-executor on Twitter for announcements

## Reporting Security Issues

**DO NOT** open public issues for security vulnerabilities.

Email security@example.com with:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fixes (if any)

We will respond within 48 hours.
