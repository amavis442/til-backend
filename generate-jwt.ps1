# Exit on error
$ErrorActionPreference = "Stop"

Write-Host "🔧 Generating jwt key with OpenSSL" -ForegroundColor Cyan

# Check if config/jwt directory exists
$jwtPath = "config/jwt"
if (-Not (Test-Path $jwtPath)) {
    Write-Host "❌ Directory 'config/jwt' does not exist. Please create it first." -ForegroundColor Red
    exit 1
}

# Generate private key
& openssl genpkey -algorithm RSA -out "$jwtPath/private.pem" -pkeyopt rsa_keygen_bits:4096

# Extract public key
& openssl rsa -pubout -in "$jwtPath/private.pem" -out "$jwtPath/public.pem"

Write-Host "✅ Keys generated in $jwtPath" -ForegroundColor Green
