#!/bin/bash

# Debug script to test Azure App Configuration fetching
# Usage: ./debug_appconfig.sh

echo "=== Environment Variables ==="
echo "APP_CONFIG: $APP_CONFIG"
echo "APP_CONFIG_NAME: $APP_CONFIG_NAME"
echo "IMAGE_NAME: $IMAGE_NAME"
echo "IMAGE_TAG: $IMAGE_TAG"
echo "RESOURCE_GROUP: $RESOURCE_GROUP"

echo ""
echo "=== Testing Azure App Configuration ==="

# Test with APP_CONFIG
if [ -n "$APP_CONFIG" ]; then
    echo "Testing with APP_CONFIG=$APP_CONFIG"
    az appconfig kv show --name "$APP_CONFIG" --key "global-configurations" --query "{key:key,value:value}" -o json 2>/dev/null || echo "Failed to fetch global-configurations"
    
    if [ -n "$IMAGE_NAME" ]; then
        echo "Testing with service key: $IMAGE_NAME"
        az appconfig kv show --name "$APP_CONFIG" --key "$IMAGE_NAME" --query "{key:key,value:value}" -o json 2>/dev/null || echo "Failed to fetch service-specific key"
    fi
fi

echo ""
echo "=== Testing with staging label ==="
if [ -n "$APP_CONFIG" ]; then
    echo "Testing global-configurations with staging label"
    az appconfig kv show --name "$APP_CONFIG" --key "global-configurations" --label "staging" --query "{key:key,value:value}" -o json 2>/dev/null || echo "Failed to fetch global-configurations with staging label"
    
    if [ -n "$IMAGE_NAME" ]; then
        echo "Testing service key with staging label"
        az appconfig kv show --name "$APP_CONFIG" --key "$IMAGE_NAME" --label "staging" --query "{key:key,value:value}" -o json 2>/dev/null || echo "Failed to fetch service-specific key with staging label"
    fi
fi

echo ""
echo "=== Expected Variables ==="
echo "ACR_REGISTRY should be in global-configurations"
echo "Service-specific variables should be in $IMAGE_NAME key"
