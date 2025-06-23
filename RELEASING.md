# Releasing myshift-go

This document explains how to create releases for myshift-go using the automated GitHub Actions workflow.

## Release Process

The release workflow is triggered automatically when you push a semantic version tag to the repository.

### Creating a Release

1. **Ensure all changes are committed and pushed to the main branch**
   ```bash
   git push origin main
   ```

2. **Create and push a semantic version tag**
   ```bash
   # Create a new tag
   git tag v1.0.0
   
   # Or create a tag with a message
   git tag -a v1.0.0 -m "Release version 1.0.0"
   
   # Push the tag to trigger the release workflow
   git push origin v1.0.0
   ```

### Tag Format

The workflow supports the following tag formats:
- **Release versions**: `v1.0.0`, `v2.1.3`, etc.
- **Pre-release versions**: `v1.0.0-beta`, `v1.0.0-rc1`, `v2.0.0-alpha.1`, etc.

Pre-release tags (containing a hyphen) will be marked as "pre-release" in GitHub.

### What the Workflow Does

When a tag is pushed, the workflow automatically:

1. **Runs Tests**: Ensures all tests pass before building
2. **Runs Linting**: Checks code quality with golangci-lint
3. **Builds Binaries**: Creates binaries for multiple platforms:
   - Linux AMD64 and ARM64
   - macOS AMD64 and ARM64  
   - Windows AMD64
4. **Creates Archives**: Packages binaries with documentation
5. **Generates Checksums**: Creates SHA256 checksums for security
6. **Creates GitHub Release**: Automatically creates a GitHub release with:
   - Release notes generated from git commits
   - Download instructions
   - Security verification information
   - All platform binaries and checksums

### Release Assets

Each release includes the following files:

```
myshift-v1.0.0-linux-amd64.tar.gz
myshift-v1.0.0-linux-amd64.tar.gz.sha256
myshift-v1.0.0-linux-arm64.tar.gz
myshift-v1.0.0-linux-arm64.tar.gz.sha256
myshift-v1.0.0-darwin-amd64.tar.gz
myshift-v1.0.0-darwin-amd64.tar.gz.sha256
myshift-v1.0.0-darwin-arm64.tar.gz
myshift-v1.0.0-darwin-arm64.tar.gz.sha256
myshift-v1.0.0-windows-amd64.zip
myshift-v1.0.0-windows-amd64.zip.sha256
```

### Version Information

The workflow automatically injects the version information into the binary during build time. Users can check the version with:

```bash
myshift --version
```

### Manual Release Process (if needed)

If you need to create a release manually:

1. **Build for all platforms**:
   ```bash
   # Linux AMD64
   GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=v1.0.0" -o myshift-linux-amd64 ./cmd/myshift
   
   # macOS AMD64
   GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=v1.0.0" -o myshift-darwin-amd64 ./cmd/myshift
   
   # Windows AMD64
   GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=v1.0.0" -o myshift-windows-amd64.exe ./cmd/myshift
   ```

2. **Create archives and checksums**:
   ```bash
   # Create tar.gz for Unix-like systems
   tar -czf myshift-v1.0.0-linux-amd64.tar.gz myshift-linux-amd64 README.md LICENSE
   
   # Create zip for Windows
   zip myshift-v1.0.0-windows-amd64.zip myshift-windows-amd64.exe README.md LICENSE
   
   # Generate checksums
   sha256sum myshift-v1.0.0-*.* > checksums.txt
   ```

3. **Create GitHub release manually** through the GitHub web interface

## Versioning Guidelines

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version when you make incompatible API changes
- **MINOR** version when you add functionality in a backwards compatible manner  
- **PATCH** version when you make backwards compatible bug fixes

### Examples:

- `v1.0.0` - Initial stable release
- `v1.0.1` - Bug fix release
- `v1.1.0` - New feature release
- `v2.0.0` - Breaking changes
- `v1.1.0-beta` - Beta release
- `v2.0.0-rc1` - Release candidate

## Troubleshooting

### Workflow Fails

If the release workflow fails:

1. Check the GitHub Actions tab for error details
2. Ensure all tests pass locally: `go test ./...`
3. Verify the tag format matches the expected pattern
4. Check that the repository has the necessary permissions

### Missing Binaries

If some platform binaries are missing:

1. Check the build matrix in `.github/workflows/release.yml`
2. Verify the GOOS/GOARCH combinations are valid
3. Check for build failures in the workflow logs

### Version Not Updated

If the binary shows the wrong version:

1. Verify the ldflags in the workflow are correct
2. Check that the version variable in `main.go` is properly defined
3. Ensure you're using the binary from the release, not a local build

## Security

All release binaries include SHA256 checksums. Users should verify downloads:

```bash
# Download both the binary and checksum
curl -LO https://github.com/jdcasey/myshift-go/releases/download/v1.0.0/myshift-v1.0.0-linux-amd64.tar.gz
curl -LO https://github.com/jdcasey/myshift-go/releases/download/v1.0.0/myshift-v1.0.0-linux-amd64.tar.gz.sha256

# Verify checksum
sha256sum -c myshift-v1.0.0-linux-amd64.tar.gz.sha256
``` 