---
description: "Enforced Go module dependency management standards for AI suggestions"
applyTo: "**/go.mod"
version: "0.0.2"
---

# Go Module Dependency Management Standards (ENFORCED)

**IMPORTANT**: These are **enforced dependency management standards** for this project. When adding, updating, or managing dependencies in `go.mod`, you MUST follow these security and maintenance practices. Treat every dependency as untrusted input requiring verification.

## How to Apply These Standards

1. **Verify before suggesting** - Check package existence and reputation before recommending any dependency
2. **Security first** - Always check for vulnerabilities and license compatibility
3. **Prefer standard library** - Use Go's standard library when functionality exists
4. **Use semantic versions** - Always specify tagged versions, never use `@latest` in production
5. **Follow proper workflow** - Understand the relationship between `go get`, imports, and `go mod tidy`
6. **Document reasoning** - Explain why a dependency is needed and what alternatives were considered

---

## Dependency Selection and Vetting

### Rule 1: Verify Package Existence (Anti-Hallucination)

**Purpose**: Ensure suggested dependencies are real, reputable packages, not AI-generated "hallucinations".

**Why**: AI models can suggest non-existent packages that match plausible naming patterns. Attackers exploit this with "slopsquatting"—creating malicious packages to match commonly hallucinated names.

**Do**:
```go
// GOOD: Verified real package with specific version
require (
    github.com/stretchr/testify v1.9.0
    github.com/google/uuid v1.6.0
)
```

**Verification Steps**:
1. Check package exists on pkg.go.dev
2. Verify GitHub repository has substantial stars/activity
3. Confirm package matches intended functionality
4. Check official documentation for recommended import path

**Don't**:
```go
// BAD: Unverified package that may be hallucinated
require (
    github.com/common/stringutils v1.0.0  // Does not exist!
    github.com/helpers/jsonhelper v2.1.0  // Likely hallucinated!
)
```

---

### Rule 2: Prefer Latest Stable Tagged Versions

**Purpose**: Use semantic versioned releases rather than pseudo-versions or `@latest`.

**Why**: Pseudo-versions pin to arbitrary commits without guarantees of stability. `@latest` produces non-reproducible builds.

**Do**:
```go
// GOOD: Specific semantic version tags
require (
    github.com/gorilla/mux v1.8.1
    go.uber.org/zap v1.27.0
)
```

**Don't**:
```go
// BAD: Pseudo-version (only use if explicitly requested)
require (
    github.com/example/pkg v0.0.0-20241015-a1b2c3d4e5f6
)

// BAD: @latest breaks reproducibility
// go get github.com/example/pkg@latest
```

---

### Rule 3: Check for Known Vulnerabilities

**Purpose**: Prevent introducing dependencies with known security issues.

**Why**: Vulnerable dependencies expose the project to exploits. The Go vulnerability database tracks known CVEs.

**Command**:
```bash
# Always run after adding/updating dependencies
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

**Do**:
```go
// GOOD: Updated to patched version
require (
    github.com/gin-gonic/gin v1.10.0  // Fixed CVE-2023-XXXXX
)
```

**Don't**:
```go
// BAD: Version with known critical vulnerability
require (
    github.com/gin-gonic/gin v1.6.0  // Contains CVE-2020-XXXXX
)
```

**When suggesting dependencies**:
- State: "After adding this dependency, run `govulncheck` to verify no vulnerabilities"
- If aware of vulnerabilities in certain versions, explicitly warn against them

---

### Rule 4: Standard Library First

**Purpose**: Prefer Go's standard library over third-party dependencies when functionality exists.

**Why**: Standard library is maintained by the Go team, has zero external risk, and is always available.

**Do**:
```go
// GOOD: Use standard library
import (
    "encoding/json"
    "net/http"
    "crypto/rand"
)
```

**Don't**:
```go
// BAD: Unnecessary third-party dependency
require (
    github.com/some/jsonparser v1.0.0  // encoding/json exists
)
```

**When to use third-party**:
- Standard library lacks required functionality
- Performance requirements exceed standard library capabilities
- Specialized domain logic (e.g., JWT handling, advanced validation)

---

### Rule 5: Check License Compatibility

**Purpose**: Ensure dependency licenses are compatible with the project's license.

**Why**: Incompatible licenses can force unwanted license changes or create legal issues.

**Do**:
```go
// GOOD: MIT/Apache 2.0/BSD licenses (permissive)
require (
    github.com/stretchr/testify v1.9.0       // MIT
    go.uber.org/zap v1.27.0                  // MIT
    google.golang.org/grpc v1.62.0           // Apache 2.0
)
```

**Don't**:
```go
// BAD: AGPL/GPL license in permissive project
require (
    github.com/some/agpl-package v1.0.0  // AGPL-3.0 (copyleft)
)
```

**Verification**:
- Check LICENSE file in dependency repository
- For commercial projects, avoid AGPL/GPL licenses
- For AGPL/GPL projects, ensure all dependencies are GPL-compatible

---

### Rule 6: Avoid Deprecated or Unmaintained Packages

**Purpose**: Prevent using abandoned packages that won't receive security updates.

**Why**: Unmaintained packages accumulate vulnerabilities and become compatibility liabilities.

**Warning Signs**:
- No commits in 2+ years
- Many unresolved security issues
- "Deprecated" or "Archived" notice in repository
- No response to critical issues

**Do**:
```go
// GOOD: Actively maintained package
require (
    github.com/golang-jwt/jwt/v5 v5.2.0  // Maintained fork
)
```

**Don't**:
```go
// BAD: Deprecated/unmaintained package
require (
    github.com/dgrijalva/jwt-go v3.2.0  // Archived, use golang-jwt/jwt instead
)
```

---

### Rule 7: Limit to Two Require Blocks

**Purpose**: Maintain a clean, organized `go.mod` file with at most two `require` blocks.

**Why**: Go convention separates direct dependencies from indirect ones. Multiple require blocks create confusion and make dependency management harder.

**Structure**:
1. **First require block**: Direct dependencies (packages your code imports)
2. **Second require block**: Indirect dependencies (marked with `// indirect`)

**Do**:
```go
// GOOD: Two require blocks - direct and indirect
module github.com/example/project

go 1.24

require (
    github.com/gorilla/mux v1.8.1
    github.com/stretchr/testify v1.9.0
    go.uber.org/zap v1.27.0
)

require (
    github.com/davecgh/go-spew v1.1.1 // indirect
    github.com/pmezard/go-difflib v1.0.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)
```

**Don't**:
```go
// BAD: Multiple require blocks scattered throughout
module github.com/example/project

go 1.24

require github.com/gorilla/mux v1.8.1

require (
    github.com/stretchr/testify v1.9.0
)

require go.uber.org/zap v1.27.0

require (
    github.com/davecgh/go-spew v1.1.1 // indirect
)

require (
    github.com/pmezard/go-difflib v1.0.0 // indirect
)
```

**Maintenance**:
- Run `go mod tidy` to automatically organize require blocks
- Direct dependencies are determined by actual imports in `.go` files
- Indirect dependencies are automatically managed by Go tooling

---

## Managing Dependencies with Go Commands

### Rule 8: Use `go mod tidy` for Synchronization

**Purpose**: Keep `go.mod` and `go.sum` synchronized with actual imports in code.

**Why**: `go mod tidy` adds missing dependencies and removes unused ones, ensuring reproducible builds.

**Command**:
```bash
go mod tidy
```

**When to run**:
- After adding/removing import statements in `.go` files
- Before committing changes
- After merging branches

**Do**:
```bash
# Workflow when adding imports
# 1. Add import to .go file
echo 'import "github.com/google/uuid"' >> main.go

# 2. Run go mod tidy
go mod tidy

# 3. Verify in go.mod
cat go.mod
```

---

### Rule 9: Understand `go get` vs `go mod tidy` Workflow

**Purpose**: Understand when dependencies persist in `go.mod`.

**Why**: `go get` adds dependencies, but `go mod tidy` removes them if not imported in source code.

**Complete Workflow**:
```bash
# Step 1: Fetch specific version (optional, if you want to control version)
go get github.com/google/uuid@v1.6.0

# Step 2: Add import in .go file
# Add: import "github.com/google/uuid"

# Step 3: Synchronize go.mod
go mod tidy
```

**Important**: If you run `go get` but never import the package in source code, `go mod tidy` will remove it from `go.mod`.

**Do**:
```bash
# GOOD: Complete workflow
go get github.com/stretchr/testify@v1.9.0
# Edit code to add: import "github.com/stretchr/testify/assert"
go mod tidy
```

**Don't**:
```bash
# BAD: go get without importing
go get github.com/stretchr/testify@v1.9.0
go mod tidy  # Removes testify because no imports exist!
```

---

### Rule 10: Updating Dependencies Safely

**Purpose**: Update dependencies to receive bug fixes and security patches.

**Why**: Updates bring fixes, but major version changes introduce breaking changes requiring code modifications.

**Commands**:
```bash
# Update to latest minor/patch (safe, no breaking changes)
go get -u=patch ./...   # Patch updates only (1.2.3 -> 1.2.4)
go get -u ./...         # Minor/patch updates (1.2.3 -> 1.3.0)

# Update specific dependency
go get -u github.com/example/pkg

# Check available updates
go list -u -m all
```

**Do**:
```bash
# GOOD: Patch updates (safe)
go get -u=patch ./...
go mod tidy

# Then verify with tests
go test ./...
```

**Don't**:
```bash
# BAD: Blindly updating majors without code review
go get -u github.com/lib/pq  # May jump v1 -> v2 (breaking!)
```

**Major Version Upgrades**:
- Major versions (v2, v3) are **breaking changes** by semantic versioning
- Require manual code changes and thorough testing
- Update import paths: `github.com/lib/pq/v2`

---

## Tool Dependencies (Go 1.24+)

### Rule 11: Use the `tool` Directive for Developer Tools

**Purpose**: Pin versions of development tools (linters, generators, etc.) without bloating production dependencies.

**Why**: Before Go 1.24, tools were tracked in `go.mod` but excluded from builds via build tags. The `tool` directive is the official solution.

**Do**:
```go
// GOOD: go.mod with tool directive (Go 1.24+)
module github.com/example/project

go 1.24

require (
    github.com/stretchr/testify v1.9.0
)

tool (
    golang.org/x/tools/cmd/goimports
    github.com/golangci/golangci-lint/cmd/golangci-lint
)
```

**Adding Tools**:
```bash
# Add a versioned tool
go get -tool golang.org/x/tools/cmd/goimports@latest
go get -tool github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.2
```

**Running Tools**:
```bash
# Run version-pinned tool
go tool goimports -w .
go tool golangci-lint run
```

**Don't**:
```go
// BAD: Old pattern (pre-Go 1.24)
// +build tools

package tools

import (
    _ "golang.org/x/tools/cmd/goimports"
)
```

---

## AI-Specific Guidance for Dependency Suggestions

### When I Ask for a Dependency

**You MUST**:
1. ✅ Verify the package exists on pkg.go.dev
2. ✅ Provide the EXACT import path (not a guess)
3. ✅ Specify a concrete semantic version (e.g., `v1.9.0`), not `@latest`
4. ✅ Mention the package's license
5. ✅ State if the package is actively maintained (recent commits)
6. ✅ Suggest alternatives if standard library suffices
7. ✅ Remind me to run `govulncheck` after adding

**Response Template**:
```
I recommend adding `github.com/google/uuid` for UUID generation.

Package Details:
- Import: github.com/google/uuid
- Version: v1.6.0 (latest stable as of October 2024)
- License: BSD-3-Clause (permissive)
- Maintenance: Active (last commit within 6 months)
- Vulnerability: None known

Command:
```bash
go get github.com/google/uuid@v1.6.0
```

Then add to your code:
```go
import "github.com/google/uuid"
```

Finally, run:
```bash
go mod tidy
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

Alternative: If you only need basic random IDs, consider crypto/rand from the standard library.
```

### When I Ask "How Do I Install X?"

**Always provide the complete workflow**:
1. `go get <package>@<version>`
2. Add `import` statement to relevant `.go` file
3. Run `go mod tidy`
4. Run `go test ./...` to ensure compatibility
5. Run `govulncheck` for security check

---

## Summary: Checklist for Every Dependency Change

Before suggesting or adding a dependency, verify:

- [ ] Package exists on pkg.go.dev (not hallucinated)
- [ ] Using specific semantic version tag (not pseudo-version or `@latest`)
- [ ] License is compatible with project
- [ ] Package is actively maintained (recent commits, responsive maintainers)
- [ ] No known vulnerabilities in suggested version
- [ ] Standard library doesn't already provide this functionality
- [ ] Proper workflow: `go get` → import in code → `go mod tidy`
- [ ] Run `govulncheck` after changes
- [ ] For tools: use `tool` directive (Go 1.24+)

---

By following these standards, you will maintain a **secure, reproducible, and maintainable** build environment for this Go project.
