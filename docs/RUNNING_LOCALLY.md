# Running Bitcoin SPV Relayer Locally

Complete guide for setting up and running the Native Bitcoin SPV Relayer on your local development environment.

## Table of Contents

- [Overview](#overview)
- [What is the Bitcoin SPV Relayer?](#what-is-the-bitcoin-spv-relayer)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Relayer](#running-the-relayer)
- [Monitoring & Debugging](#monitoring--debugging)
- [Troubleshooting](#troubleshooting)
- [Development Workflow](#development-workflow)

---

## Overview

The Bitcoin SPV Relayer is a critical component of the Native ecosystem that:
- Monitors Bitcoin blockchain for new blocks and transactions
- Syncs Bitcoin block headers to Native's light client
- Creates and submits SPV (Simplified Payment Verification) proofs
- Enables trust-minimized verification of Bitcoin transactions

This guide walks you through setting up and running the relayer locally for development and testing.

---

## What is the Bitcoin SPV Relayer?

The relayer acts as a **messenger** between three key components:

1. **Bitcoin Full Node**: Records all Bitcoin activity (blocks, transactions)
2. **Native Light Client**: Lightweight system that validates transactions using block headers
3. **SPV Relayer**: Connects the two, syncing headers and creating cryptographic proofs

### How It Works

```
┌─────────────────┐         ┌──────────────┐         ┌─────────────────┐
│  Bitcoin Full   │ ◄─────► │ SPV Relayer  │ ◄─────► │  Native Light   │
│      Node       │         │   (You!)     │         │     Client      │
└─────────────────┘         └──────────────┘         └─────────────────┘
       │                           │                          │
       │ Provides blocks          │ Syncs headers            │ Validates
       │ and transactions         │ Creates SPV proofs       │ transactions
       └──────────────────────────┴──────────────────────────┘
```

**The Relayer's Job:**
1. **Initial Sync**: Check if light client is missing block headers → fetch and submit them in batches
2. **Real-Time Monitoring**: Listen for new Bitcoin blocks → submit headers immediately
3. **Transaction Verification**: When transactions occur → create SPV proofs → submit to light client

---

## Prerequisites

### System Requirements

- **OS**: Linux (Ubuntu 20.04+), macOS (11+), or Windows with WSL2
- **RAM**: 4GB minimum, 8GB+ recommended
- **Disk Space**: 10GB for relayer + Bitcoin testnet data
- **CPU**: 2+ cores recommended

### Required Software

**1. Go (1.21+)**
```bash
# Check if Go is installed
go version

# If not installed (Ubuntu/Debian):
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# macOS (using Homebrew):
brew install go
```

**2. Make**
```bash
# Ubuntu/Debian:
sudo apt-get update
sudo apt-get install build-essential

# macOS:
# Already included with Xcode Command Line Tools
xcode-select --install
```

**3. Git**
```bash
# Ubuntu/Debian:
sudo apt-get install git

# macOS:
brew install git
```

**4. Bitcoin Node (Optional but Recommended)**

For production-like testing, you can run Bitcoin Core in testnet mode:

```bash
# Install Bitcoin Core
# Visit: https://bitcoin.org/en/download

# Or use Docker:
docker pull ruimarinho/bitcoin-core:latest
docker run -d \
  --name bitcoin-testnet \
  -p 18332:18332 \
  ruimarinho/bitcoin-core:latest \
  -testnet \
  -rpcuser=native \
  -rpcpassword=your_secure_password \
  -rpcallowip=0.0.0.0/0
```

**For development**, you can use public testnet RPC endpoints:
- `https://blockstream.info/testnet/api/`
- Or connect to Native's testnet nodes (see Discord for endpoints)

---

## Installation

### 1. Clone the Repository

```bash
cd ~/projects  # or your preferred directory
git clone https://github.com/gonative-cc/relayer.git
cd relayer
```

### 2. Install Dependencies

```bash
# Download Go dependencies
go mod download

# Verify everything is working
go mod verify
```

### 3. Setup Git Hooks (Optional but Recommended)

```bash
make setup-hooks
```

This sets up:
- Pre-commit hooks for code quality checks
- DCO (Developer Certificate of Origin) sign-off requirements

---

## Configuration

### 1. Create Environment File

```bash
cp .env.example .env
```

### 2. Configure Environment Variables

Edit `.env` with your preferred editor:

```bash
nano .env  # or vim, code, etc.
```

**Key Configuration Parameters:**

```bash
# === Bitcoin RPC Configuration ===
# For local Bitcoin Core testnet:
BITCOIN_RPC_HOST=localhost:18332
BITCOIN_RPC_USER=native
BITCOIN_RPC_PASSWORD=your_secure_password

# For public testnet endpoint:
# BITCOIN_RPC_HOST=https://blockstream.info/testnet/api
# BITCOIN_RPC_USER=  # Not needed for public endpoints
# BITCOIN_RPC_PASSWORD=

# === Native Network Configuration ===
NATIVE_RPC_URL=https://testnet-rpc.gonative.cc
NATIVE_BTCINDEXER_BEARER_TOKEN=your_bearer_token_here

# === Database Configuration ===
DB_HOST=localhost
DB_PORT=5432
DB_NAME=native_relayer
DB_USER=postgres
DB_PASSWORD=postgres

# === Logging ===
LOG_LEVEL=info  # Options: debug, info, warn, error
LOG_FORMAT=json  # Options: json, text

# === Relayer Behavior ===
SYNC_INTERVAL=10s            # How often to check for new blocks
BATCH_SIZE=100               # Number of headers to submit at once
START_BLOCK_HEIGHT=2000000   # Bitcoin block height to start from (testnet genesis)
```

### 3. Database Setup (If Using Local DB)

```bash
# Start PostgreSQL (Docker):
docker run -d \
  --name native-postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=native_relayer \
  -p 5432:5432 \
  postgres:15

# Or install locally (Ubuntu):
sudo apt-get install postgresql postgresql-contrib

# Create database:
createdb native_relayer
```

### 4. Get Bearer Token

To get your `NATIVE_BTCINDEXER_BEARER_TOKEN`:
1. Join [Native Discord](https://discord.gg/gonative)
2. Request testnet access in #developer-support
3. Receive your bearer token

---

## Running the Relayer

### Build the Project

```bash
make build
```

This compiles the relayer binary to `./bin/relayer`.

### Start the Relayer

**Option 1: Using Make (Recommended)**
```bash
make start
```

**Option 2: Direct Binary Execution**
```bash
./bin/relayer
```

**Option 3: Run with Custom Config**
```bash
./bin/relayer --config=custom.config.yml
```

### Expected Output

When the relayer starts successfully, you should see:

```
{"level":"info","time":"2026-05-18T16:00:00Z","message":"Starting Native Bitcoin SPV Relayer"}
{"level":"info","block":2000000,"message":"Starting from block height"}
{"level":"info","message":"Connecting to Bitcoin RPC"}
{"level":"info","message":"Connected to Bitcoin node","network":"testnet"}
{"level":"info","message":"Checking Native light client status"}
{"level":"info","current_height":2000500,"message":"Native light client synced to"}
{"level":"info","headers":500,"message":"Syncing missing headers"}
{"level":"info","message":"Initial sync complete"}
{"level":"info","message":"Listening for new Bitcoin blocks"}
```

---

## Monitoring & Debugging

### View Real-Time Logs

```bash
# If running with make start:
tail -f logs/relayer.log

# If running with Docker:
docker logs -f native-relayer
```

### Check Sync Status

```bash
# Query the relayer's status endpoint:
curl http://localhost:8080/status

# Expected response:
{
  "status": "synced",
  "bitcoin_height": 2500000,
  "native_height": 2500000,
  "pending_headers": 0,
  "last_sync": "2026-05-18T16:30:00Z"
}
```

### Enable Debug Logging

For verbose output during development:

```bash
# In .env:
LOG_LEVEL=debug

# Or set environment variable:
export LOG_LEVEL=debug
./bin/relayer
```

### Database Inspection

```bash
# Connect to database:
psql -U postgres -d native_relayer

# Check synced blocks:
SELECT * FROM btc_blocks ORDER BY height DESC LIMIT 10;

# Check pending SPV proofs:
SELECT * FROM spv_proofs WHERE status = 'pending';

# Exit:
\q
```

---

## Troubleshooting

### Common Issues

#### **1. "Cannot connect to Bitcoin RPC"**

**Problem:** Relayer can't reach Bitcoin node.

**Solution:**
```bash
# Test connection manually:
curl --user native:your_password --data-binary '{"jsonrpc":"1.0","id":"test","method":"getblockchaininfo","params":[]}' http://localhost:18332/

# Check Bitcoin Core is running:
docker ps | grep bitcoin

# Verify RPC credentials in .env match Bitcoin Core config
```

#### **2. "Bearer token invalid"**

**Problem:** Native API rejects your token.

**Solution:**
- Verify token in .env has no extra spaces
- Request new token from Discord #developer-support
- Check you're using testnet token (not mainnet)

#### **3. "Database connection failed"**

**Problem:** PostgreSQL not accessible.

**Solution:**
```bash
# Check if PostgreSQL is running:
docker ps | grep postgres
# or
sudo systemctl status postgresql

# Test connection:
psql -U postgres -h localhost -d native_relayer -c "SELECT 1;"
```

#### **4. "Out of sync - too many missing headers"**

**Problem:** Large gap between Bitcoin and Native light client.

**Solution:**
```bash
# Increase batch size in .env:
BATCH_SIZE=500

# Or start from recent block:
START_BLOCK_HEIGHT=2450000  # Close to current testnet height

# Restart relayer:
make restart
```

#### **5. High CPU/Memory Usage**

**Problem:** Relayer consuming excessive resources.

**Solution:**
```bash
# Reduce sync frequency:
SYNC_INTERVAL=30s  # Check every 30s instead of 10s

# Lower batch size:
BATCH_SIZE=50

# Disable debug logging:
LOG_LEVEL=info
```

---

## Development Workflow

### Running Tests

```bash
# Run all tests:
make test

# Run with coverage:
make test-coverage

# Run specific package:
go test ./bitcoinspv/...

# Run with verbose output:
go test -v ./...
```

### Code Quality Checks

```bash
# Run linter:
make lint

# Fix auto-fixable issues:
make lint-fix

# Run markdown linter:
markdownlint-cli2 "**/*.md"
```

### Making Changes

**1. Create Feature Branch**
```bash
git checkout -b feat/my-feature
# or
git checkout -b fix/bug-description
```

**2. Make Changes & Commit**
```bash
# Stage changes:
git add .

# Commit with DCO sign-off:
git commit -s -m "feat: add support for batched header submission

Improves sync performance by submitting headers in configurable batches.

Signed-off-by: Your Name <your.email@example.com>"
```

**3. Push & Create PR**
```bash
git push origin feat/my-feature
```

Then create PR on GitHub following [CONTRIBUTING.md](https://github.com/gonative-cc/contributig/blob/master/CONTRIBUTING.md).

### Local E2E Testing

```bash
# Run end-to-end tests:
make e2e-test

# Or manually:
docker-compose -f e2e-bitcoin-spv.yml up
```

---

## Additional Resources

- **Documentation**: https://docs.gonative.cc
- **Discord**: https://discord.gg/gonative (#developer-support)
- **GitHub Issues**: https://github.com/gonative-cc/relayer/issues
- **Contributing Guidelines**: https://github.com/gonative-cc/contributig

---

## Next Steps

Once your local relayer is running:

1. ✅ **Monitor Bitcoin testnet** - Watch it sync headers in real-time
2. ✅ **Test SPV proof generation** - Send test transactions and verify proofs
3. ✅ **Experiment with configuration** - Try different sync strategies
4. ✅ **Contribute improvements** - Found a bug or optimization? Submit a PR!

For production deployment, see [Deployment Guide](./DEPLOYMENT.md) (coming soon).

---

**Last Updated:** May 2026  
**Maintained by:** Native Community  
**Questions?** Ask in [Discord #developer-support](https://discord.gg/gonative)
