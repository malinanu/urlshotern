#!/bin/bash
# Bash script to create admin users

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}\n🔧 Admin User Creation Script${NC}"
echo -e "${CYAN}================================${NC}\n"

# Parse command line arguments
NAME=""
EMAIL=""
PASSWORD=""
TYPE="super_admin"

while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--name)
            NAME="$2"
            shift 2
            ;;
        -e|--email)
            EMAIL="$2"
            shift 2
            ;;
        -p|--password)
            PASSWORD="$2"
            shift 2
            ;;
        -t|--type)
            TYPE="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  -n, --name      Admin name"
            echo "  -e, --email     Admin email"
            echo "  -p, --password  Admin password"
            echo "  -t, --type      Account type (default: super_admin)"
            echo "  -h, --help      Show this help message"
            echo ""
            echo "Account types:"
            echo "  - super_admin   Full system access (default)"
            echo "  - system_admin  System management"
            echo "  - support_admin User support"
            echo "  - billing_admin Billing management"
            echo "  - content_admin Content management"
            echo ""
            echo "Examples:"
            echo "  $0 -n \"Admin User\" -e \"admin@example.com\" -p \"securepass123\""
            echo "  $0 --name \"John Doe\" --email \"john@example.com\" --password \"pass123\" --type system_admin"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

# Interactive mode if parameters not provided
if [ -z "$NAME" ]; then
    read -p "Enter admin name: " NAME
fi

if [ -z "$EMAIL" ]; then
    read -p "Enter admin email: " EMAIL
fi

if [ -z "$PASSWORD" ]; then
    read -s -p "Enter admin password: " PASSWORD
    echo ""
fi

# Validate inputs
if [ -z "$NAME" ] || [ -z "$EMAIL" ] || [ -z "$PASSWORD" ]; then
    echo -e "${RED}❌ Error: Name, email, and password are required!${NC}"
    exit 1
fi

echo ""
echo "Creating admin user with:"
echo "  Name: $NAME"
echo "  Email: $EMAIL"
echo "  Type: $TYPE"
echo ""

# Get script directory and project root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"

# Run the Go script
cd "$PROJECT_ROOT"
go run scripts/create-admin.go -name="$NAME" -email="$EMAIL" -password="$PASSWORD" -type="$TYPE"

if [ $? -eq 0 ]; then
    echo -e "\n${GREEN}✅ Success! Admin user created.${NC}"
    echo -e "\n${YELLOW}Login at: http://localhost:3000/admin/login${NC}"
    echo -e "${YELLOW}Email: $EMAIL${NC}"
else
    echo -e "\n${RED}❌ Failed to create admin user. Check the error above.${NC}"
    exit 1
fi
