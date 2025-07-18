#!/bin/bash

# Exit on error
set -e

echo "🔧 Generating jwt key with openssl"

if [ ! -d "config/jwt"]; then
    echo "First create dir config/jwt"
    exit 1;
fi


# Generate private key
openssl genpkey -algorithm RSA -out config/jwt/private.pem -pkeyopt rsa_keygen_bits:4096

# Extract the public key
openssl rsa -pubout -in config/jwt/private.pem -out config/jwt/public.pem