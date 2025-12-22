# Generate self-signed SSL certificates for PostgreSQL
# This script creates certificates for development and staging environments
# For production, use proper CA-signed certificates

Write-Host "Generating PostgreSQL SSL Certificates..." -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green

# Create certificates directory
$CertDir = "certs\postgres"
New-Item -ItemType Directory -Force -Path $CertDir | Out-Null

Write-Host "1. Generating CA certificate..." -ForegroundColor Yellow
openssl req -new -x509 -days 3650 -nodes -text `
  -out "$CertDir\root.crt" `
  -keyout "$CertDir\root.key" `
  -subj "/CN=PostgreSQL-CA"

Write-Host "2. Generating server certificate request..." -ForegroundColor Yellow
openssl req -new -nodes -text `
  -out "$CertDir\server.csr" `
  -keyout "$CertDir\server.key" `
  -subj "/CN=postgres"

Write-Host "3. Signing server certificate with CA..." -ForegroundColor Yellow
openssl x509 -req -in "$CertDir\server.csr" `
  -text -days 3650 `
  -CA "$CertDir\root.crt" `
  -CAkey "$CertDir\root.key" `
  -CAcreateserial `
  -out "$CertDir\server.crt"

Write-Host ""
Write-Host "✅ Certificates generated successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Generated files:" -ForegroundColor Cyan
Write-Host "  - $CertDir\root.crt (CA certificate)"
Write-Host "  - $CertDir\root.key (CA private key)"
Write-Host "  - $CertDir\server.crt (Server certificate)"
Write-Host "  - $CertDir\server.key (Server private key)"
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. For Kubernetes, create secret:"
Write-Host "     kubectl create secret generic postgres-ssl-certs \"
Write-Host "       --from-file=server.crt=$CertDir\server.crt \"
Write-Host "       --from-file=server.key=$CertDir\server.key \"
Write-Host "       --from-file=root.crt=$CertDir\root.crt \"
Write-Host "       -n production"
Write-Host ""
Write-Host "  2. For Docker Compose, mount the certs directory as volume"
Write-Host ""
Write-Host "⚠️  IMPORTANT: Add certs/ to .gitignore!" -ForegroundColor Red
