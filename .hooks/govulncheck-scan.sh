#!/bin/bash
set -e

# Determine the path to govulncheck
GOPATH=$(go env GOPATH)
GOVULNCHECK="${GOPATH}/bin/govulncheck"

# Check if govulncheck is installed
if [ ! -f "$GOVULNCHECK" ]; then
	echo "govulncheck is not installed. Installing..."
	if ! go install golang.org/x/vuln/cmd/govulncheck@latest; then
		echo "Warning: Failed to install govulncheck, skipping vulnerability scan"
		exit 0
	fi
	echo "govulncheck installed successfully"
fi

# Verify govulncheck is now available
if [ ! -f "$GOVULNCHECK" ]; then
	echo "Warning: govulncheck not found after installation, skipping scan"
	exit 0
fi

# Run govulncheck vulnerability scan
echo "Running govulncheck vulnerability scan..."
if ! output=$("$GOVULNCHECK" ./... 2>&1); then
	echo ""
	echo "❌ govulncheck found vulnerabilities in dependencies!"
	echo "$output"
	echo ""
	echo "Please fix the vulnerabilities before committing."
	echo ""
	echo "To update vulnerable dependencies, run:"
	echo "  go get -u <package>@<fixed-version>"
	echo "  go mod tidy"
	echo ""
	echo "For more information, visit: https://go.dev/security/vuln"
	exit 1
fi

echo "✅ No vulnerabilities found by govulncheck"
