#!/bin/bash
# Generate self-signed SSL certificates for PostgreSQL
# This script creates certificates for development and staging environments
# For production, use proper CA-signed certificates or cert-manager

set -e

echo "Generating PostgreSQL SSL Certificates..."
echo "=========================================="

# Create certificates directory
CERT_DIR="certs/postgres"
mkdir -p "$CERT_DIR"

echo "1. Generating CA certificate..."
openssl req -new -x509 -days 3650 -nodes -text \
  -out "$CERT_DIR/root.crt" \
  -keyout "$CERT_DIR/root.key" \
  -subj "/CN=PostgreSQL-CA"

echo "2. Generating server certificate request..."
openssl req -new -nodes -text \
  -out "$CERT_DIR/server.csr" \
  -keyout "$CERT_DIR/server.key" \
  -subj "/CN=postgres"

echo "3. Signing server certificate with CA..."
openssl x509 -req -in "$CERT_DIR/server.csr" \
  -text -days 3650 \
  -CA "$CERT_DIR/root.crt" \
  -CAkey "$CERT_DIR/root.key" \
  -CAcreateserial \
  -out "$CERT_DIR/server.crt"

echo "4. Setting proper permissions..."
chmod 600 "$CERT_DIR/server.key" "$CERT_DIR/root.key"
chmod 644 "$CERT_DIR/server.crt" "$CERT_DIR/root.crt"

echo ""
echo "✅ Certificates generated successfully!"
echo ""
echo "Generated files:"
echo "  - $CERT_DIR/root.crt (CA certificate)"
echo "  - $CERT_DIR/root.key (CA private key)"
echo "  - $CERT_DIR/server.crt (Server certificate)"
echo "  - $CERT_DIR/server.key (Server private key)"
echo ""
echo "Next steps:"
echo "  1. For Kubernetes, create secret:"
echo "     kubectl create secret generic postgres-ssl-certs \\"
echo "       --from-file=server.crt=$CERT_DIR/server.crt \\"
echo "       --from-file=server.key=$CERT_DIR/server.key \\"
echo "       --from-file=root.crt=$CERT_DIR/root.crt \\"
echo "       -n production"
echo ""
echo "  2. For Docker Compose, mount the certs directory as volume"
echo ""
echo "⚠️  IMPORTANT: Add certs/ to .gitignore!"
