package tskmgr

# Create a tools.go file in your project
cat > tools.go << 'EOF'
//go:build tools
// +build tools

package main

import (
	_ "github.com/golang-migrate/migrate/v4/cmd/migrate"
)
EOF