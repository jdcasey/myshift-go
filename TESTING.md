# Testing Strategy for MyShift-Go

This document outlines the testing strategy for the MyShift-Go project, providing guidance on how to write, maintain, and extend tests while keeping the test suite manageable and reliable.

## 📋 **Testing Philosophy**

Our testing strategy follows these principles:

1. **Fast feedback**: Unit tests should run quickly (< 1s for the entire suite)
2. **Reliable**: Tests should not be flaky or dependent on external services
3. **Maintainable**: Test code should be as clean and readable as production code
4. **Comprehensive**: Focus on business logic, edge cases, and error conditions
5. **Isolated**: Each test should be independent and not affect others

## 🏗️ **Testing Architecture**

### **Layer 1: Unit Tests (Fast & Isolated)**
- **Purpose**: Test individual functions and components in isolation
- **Speed**: < 10ms per test
- **Dependencies**: All external dependencies mocked
- **Coverage**: Business logic, validation, data processing

### **Layer 2: Integration Tests (Medium Speed)**
- **Purpose**: Test component interactions with mocked external services
- **Speed**: < 100ms per test
- **Dependencies**: Real internal components, mocked external APIs
- **Coverage**: Command workflows, configuration loading

### **Layer 3: Contract Tests (Optional)**
- **Purpose**: Verify API request/response structures
- **Speed**: Slower, run less frequently
- **Dependencies**: Can hit real APIs (for verification only)
- **Coverage**: API compatibility, data structures

## 🎯 **What We Test**

### ✅ **High Priority (Must Test)**
- **Business logic**: Core functionality like shift finding, time calculations
- **Input validation**: Parameter checking, error conditions
- **Configuration**: File loading, validation, error handling
- **API interactions**: Mocked PagerDuty client calls
- **Error handling**: Expected error conditions and messages

### ⚠️ **Medium Priority (Should Test)**
- **Output formatting**: CLI output verification
- **CLI parsing**: Flag handling and validation
- **Integration workflows**: End-to-end command execution

### ⬇️ **Low Priority (Nice to Have)**
- **Performance**: Benchmarks for critical paths
- **Edge cases**: Unusual but possible scenarios
- **User experience**: Help text, error messages

## 🛠️ **Testing Patterns & Tools**

### **Interface-Based Mocking**

We use interfaces to enable easy mocking:

```go
// Define interface for external dependencies
type PagerDutyClient interface {
    FindUserByEmail(email string) (*myshift.User, error)
    GetOnCalls(params url.Values) ([]myshift.OnCall, error)
    // ... other methods
}

// Commands use the interface, not concrete types
type NextCommand struct {
    client PagerDutyClient  // Interface, not *pagerduty.Client
}
```

### **Table-Driven Tests**

Use table-driven tests for multiple scenarios:

```go
func TestNextCommand_Execute(t *testing.T) {
    tests := []struct {
        testName   string
        userEmail  string
        setupMock  func(*MockPagerDutyClient, time.Time)
        wantErr    bool
        wantOutput []string
    }{
        {
            testName:  "user not found",
            userEmail: "nonexistent@example.com",
            setupMock: func(mock *MockPagerDutyClient, now time.Time) {
                // No users added
            },
            wantErr: true,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.testName, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### **Test Fixtures**

Use fixtures for common test setup:

```go
func NewTestFixture() *TestFixture {
    now := time.Now()
    mockClient := NewMockPagerDutyClient()
    
    // Common setup
    mockClient.AddUser("USER001", "John Doe", "john@example.com")
    
    return &TestFixture{
        MockClient: mockClient,
        Now:        now,
    }
}
```

### **Dependency Injection for Testing**

Make functions testable by injecting dependencies:

```go
// Production code uses a variable that can be overridden
var configPathsFunc = getConfigPaths

func Load() (*myshift.Config, error) {
    for _, path := range configPathsFunc() {
        // ... implementation
    }
}

// Test code overrides the variable
func TestLoad(t *testing.T) {
    originalFunc := configPathsFunc
    configPathsFunc = func() []string {
        return []string{"/path/to/test/config.yaml"}
    }
    defer func() { configPathsFunc = originalFunc }()
    
    // Test with mocked paths
}
```

## 📁 **Test Organization**

```
myshift-go/
├── internal/
│   ├── commands/
│   │   ├── next.go
│   │   ├── next_test.go          # Unit tests for NextCommand
│   │   ├── testutil_test.go      # Shared test utilities and mocks
│   │   └── ...
│   ├── config/
│   │   ├── config.go
│   │   ├── config_test.go        # Unit tests for config loading
│   │   └── ...
│   └── pagerduty/
│       ├── client.go
│       ├── interfaces.go         # Interfaces for mocking
│       └── client_test.go        # (Future) HTTP client tests
└── testdata/                     # Test fixtures and sample data
    ├── configs/
    │   ├── valid.yaml
    │   └── invalid.yaml
    └── api_responses/
        ├── users.json
        └── oncalls.json
```

## 🚀 **Running Tests**

### **All Tests**
```bash
go test ./...
```

### **Specific Package**
```bash
go test ./internal/commands -v
```

### **With Coverage**
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### **Benchmarks**
```bash
go test ./... -bench=.
```

### **Race Detection**
```bash
go test ./... -race
```

## 📝 **Writing New Tests**

### **1. Command Tests**

When adding a new command:

```go
func TestNewCommand_Execute(t *testing.T) {
    tests := []struct {
        testName   string
        // ... test parameters
        setupMock  func(*MockPagerDutyClient, time.Time)
        wantErr    bool
        wantOutput []string
    }{
        // ... test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.testName, func(t *testing.T) {
            fixture := NewTestFixture()
            tt.setupMock(fixture.MockClient, fixture.Now)
            
            cmd := NewYourCommand(fixture.MockClient)
            
            // Capture output if needed
            oldStdout := os.Stdout
            r, w, _ := os.Pipe()
            os.Stdout = w
            
            err := cmd.Execute(/* parameters */)
            
            w.Close()
            os.Stdout = oldStdout
            output, _ := io.ReadAll(r)
            
            // Verify results
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
            }
            
            // Check output if no error
            if !tt.wantErr {
                for _, expected := range tt.wantOutput {
                    if !strings.Contains(string(output), expected) {
                        t.Errorf("Expected output to contain %q", expected)
                    }
                }
            }
        })
    }
}
```

### **2. Configuration Tests**

When adding configuration features:

```go
func TestNewConfigFeature(t *testing.T) {
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "test.yaml")
    
    configContent := `
key: "value"
number: 42
`
    
    err := os.WriteFile(configPath, []byte(configContent), 0644)
    if err != nil {
        t.Fatalf("Failed to create test config: %v", err)
    }
    
    // Test loading and validation
    config, err := loadFromFile(configPath)
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }
    
    // Verify values
    // ...
}
```

### **3. Mock Client Extensions**

When adding new PagerDuty API methods:

```go
// Add to PagerDutyClient interface
type PagerDutyClient interface {
    // ... existing methods
    NewMethod(param string) (*Result, error)
}

// Add to MockPagerDutyClient
func (m *MockPagerDutyClient) NewMethod(param string) (*Result, error) {
    m.NewMethodCalls = append(m.NewMethodCalls, param)
    
    // Return test data based on input
    if result, exists := m.NewMethodResults[param]; exists {
        return result, nil
    }
    
    return nil, fmt.Errorf("not found: %s", param)
}
```

## 🎯 **Best Practices**

### **✅ Do:**
- **Use descriptive test names**: `TestNextCommand_Execute_UserNotFound`
- **Use clear field names**: Use `testName` instead of `name` in test structs to avoid confusion with user data
- **Test one thing per test**: Single responsibility
- **Use table-driven tests**: For multiple scenarios
- **Mock external dependencies**: Keep tests fast and reliable
- **Test error conditions**: Not just happy paths
- **Use temp directories**: For file operations
- **Clean up resources**: Use `defer` for cleanup

### **❌ Don't:**
- **Test implementation details**: Focus on behavior
- **Use real external services**: In unit tests
- **Write overly complex tests**: Keep them simple
- **Test trivial getters/setters**: Focus on business logic
- **Ignore test maintenance**: Keep tests updated with code changes
- **Skip error scenarios**: Test both success and failure

## 🔧 **Common Patterns**

### **Testing Time-Dependent Code**
```go
// Use time.Now() in production
func (n *NextCommand) Execute(...) {
    now := time.Now()
    // ... use now
}

// For more complex time testing, consider dependency injection:
type TimeProvider interface {
    Now() time.Time
}

type RealTimeProvider struct{}
func (r RealTimeProvider) Now() time.Time { return time.Now() }

// Command uses TimeProvider
type NextCommand struct {
    client   PagerDutyClient
    timeProvider TimeProvider
}

// Test with fixed time
type FixedTimeProvider struct{ t time.Time }
func (f FixedTimeProvider) Now() time.Time { return f.t }
```

### **Testing File Operations**
```go
func TestConfigLoad(t *testing.T) {
    tmpDir := t.TempDir() // Automatically cleaned up
    configPath := filepath.Join(tmpDir, "config.yaml")
    
    // Create test file
    err := os.WriteFile(configPath, []byte("content"), 0644)
    require.NoError(t, err) // Using testify for cleaner assertions
    
    // Test loading
    // ...
}
```

### **Output Verification**
```go
func captureOutput(t *testing.T, fn func()) string {
    oldStdout := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    
    fn()
    
    w.Close()
    os.Stdout = oldStdout
    
    output, err := io.ReadAll(r)
    require.NoError(t, err)
    
    return string(output)
}
```

## 📊 **Coverage Guidelines**

- **Target**: 80%+ overall coverage
- **Priority**: 95%+ for business logic (commands package)
- **Acceptable**: 70%+ for infrastructure (config, client)
- **Monitor**: Use `go test -coverprofile` regularly

## 🚀 **CI/CD Integration**

The test suite is integrated into CI/CD pipelines:

```yaml
# .github/workflows/pr-checks.yml
- name: Run tests
  run: go test -v -race -coverprofile=coverage.out ./...

- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v4
  with:
    file: ./coverage.out
```

## 🔧 **Maintenance**

### **Regular Tasks**
- **Review test coverage**: Monthly
- **Update test data**: When APIs change
- **Refactor test utilities**: When duplicated
- **Performance check**: Run benchmarks quarterly

### **When Code Changes**
- **Update tests first**: TDD approach
- **Verify all tests pass**: Before committing
- **Update mocks**: When interfaces change
- **Review test relevance**: Remove obsolete tests

## 📚 **Additional Resources**

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Table-Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Test Examples](https://golang.org/pkg/testing/#hdr-Examples)
- [Testify Framework](https://github.com/stretchr/testify) (optional)

---

This testing strategy ensures that the MyShift-Go codebase remains reliable, maintainable, and easy to extend while keeping the test suite fast and focused on business value. 