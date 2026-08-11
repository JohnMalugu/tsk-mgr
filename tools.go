package tskmgr

cat > tools.go << 'EOF'
//go:build tools
// +build tools

package main

import (
	_ "github.com/golang-migrate/migrate/v4/cmd/migrate"
)
EOF