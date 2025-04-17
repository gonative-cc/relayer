#!/bin/bash
set -e

echo "Running E2E tests..."
./e2e/sync_and_validate_status.sh
echo "E2E Tests completed."
