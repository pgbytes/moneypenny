---
description: "Enforced Go testing standards for AI suggestions"
applyTo: "**/*_test.go"
version: "0.0.1"
---

# Go Testing Standards (ENFORCED)

**IMPORTANT**: These are **enforced testing standards** for this project. All test code MUST comply with these principles to pass review. Non-compliance will block merges.

## How to Apply These Standards

1. **Test behavior, not implementation** - Focus on observable outcomes, not internal mechanics
2. **Ensure test isolation** - Each test must be atomic and independent
3. **Test one concept per test** - Keep tests focused on a single logical behavior
4. **Follow AAA structure** - Arrange, Act, Assert with clear visual separation
5. **Use table-driven tests** - For multiple input scenarios with shared logic
6. **Mock external dependencies** - Use test doubles for isolation
7. **Write descriptive test names** - Clearly communicate intent and expected outcome

---

## Test Behavior, Not Implementation

**Purpose**: Tests should validate external contract (observable outputs), not internal mechanics (algorithms, private methods, data structures).

**Why**: Implementation-coupled tests break on refactoring even when behavior is correct, creating maintenance burden and discouraging improvement.

**Do**:
````go
// GOOD: Tests observable behavior through public API
func TestSort_GivenUnsortedSlice_ReturnsSortedSlice(t *testing.T) {
    input := []int{3, 1, 2}
    
    result := Sort(input)
    
    assert.Equal(t, []int{1, 2, 3}, result)
}
````

**Don't**:
````go
// BAD: Tests internal implementation details
func TestSort_UsesQuickSortAlgorithm(t *testing.T) {
    s := &Sorter{}
    
    s.Sort([]int{3, 1, 2})
    
    // Asserting against internal state/algorithm
    assert.True(t, s.usedQuickSort)
}
````

---

## Ensure Test Isolation (Atomic Tests)

**Purpose**: Each test must be self-contained, runnable in any order, and leave no side effects.

**Why**: Non-atomic tests cause cascading failures, complex debugging, and erode trust in the test suite. Atomicity enables parallel execution for fast CI/CD.

**Do**:
````go
// GOOD: Self-contained test with cleanup
func TestCreateUser_WithValidData_Succeeds(t *testing.T) {
    // Arrange: Create clean test database
    db := setupTestDB(t)
    t.Cleanup(func() { db.Close() })
    
    // Act
    user, err := CreateUser(db, "alice@example.com")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
}
````

**Don't**:
````go
// BAD: Relies on shared global state
var testDB *sql.DB // Shared across tests

func TestCreateUser(t *testing.T) {
    // No setup, depends on external initialization
    user, err := CreateUser(testDB, "alice@example.com")
    assert.NoError(t, err)
    // No cleanup - leaves database in modified state
}
````

---

## Test One Logical Concept Per Test

**Purpose**: Each test should verify a single behavior, scenario, or execution path.

**Why**: Focused tests provide precise failure diagnosis and clear intent. When a test fails, you immediately know what broke.

**Do**:
````go
// GOOD: Each test covers one distinct scenario
func TestWithdraw_WithSufficientFunds_Succeeds(t *testing.T) {
    account := NewAccount(100)
    
    err := account.Withdraw(50)
    
    assert.NoError(t, err)
    assert.Equal(t, 50, account.Balance())
}

func TestWithdraw_WithInsufficientFunds_ReturnsError(t *testing.T) {
    account := NewAccount(30)
    
    err := account.Withdraw(50)
    
    assert.Error(t, err)
    assert.Equal(t, 30, account.Balance()) // Balance unchanged
}
````

**Don't**:
````go
// BAD: Tests multiple unrelated concepts
func TestAccountOperations(t *testing.T) {
    account := NewAccount(100)
    
    // Testing withdrawal
    err := account.Withdraw(50)
    assert.NoError(t, err)
    
    // Testing deposit (different concept!)
    err = account.Deposit(30)
    assert.NoError(t, err)
    
    // Testing balance inquiry (yet another concept!)
    assert.Equal(t, 80, account.Balance())
}
````

---

## Follow AAA Structure (Arrange-Act-Assert)

**Purpose**: Consistent three-section structure improves readability and maintainability.

**Why**: Clear separation of setup, execution, and verification makes test intent immediately obvious.

**Do**:
````go
// GOOD: Clear AAA structure with visual separation
func TestProcessPayment_WithValidCard_ChargesCorrectAmount(t *testing.T) {
    // Arrange
    paymentProcessor := NewPaymentProcessor()
    card := &CreditCard{Number: "4111111111111111", CVV: "123"}
    
    // Act
    result, err := paymentProcessor.Process(card, 99.99)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "approved", result.Status)
    assert.Equal(t, 99.99, result.Amount)
}
````

**Don't**:
````go
// BAD: Mixed setup, execution, and verification
func TestProcessPayment(t *testing.T) {
    paymentProcessor := NewPaymentProcessor()
    result, err := paymentProcessor.Process(&CreditCard{Number: "4111111111111111"}, 99.99)
    assert.NoError(t, err)
    card := &CreditCard{Number: "4111111111111111", CVV: "123"} // Setup after Act!
    assert.Equal(t, "approved", result.Status)
    result2, _ := paymentProcessor.Process(card, 50.00) // Second Act!
    assert.Equal(t, 50.00, result2.Amount)
}
````

---

## Use Table-Driven Tests

**Purpose**: Test multiple input scenarios with minimal boilerplate using a data table.

**Why**: Reduces duplication, improves coverage, and makes adding new test cases trivial. Encourages systematic testing of edge cases.

**Do**:
````go
// GOOD: Table-driven test with subtests
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {name: "valid email", email: "user@example.com", wantErr: false},
        {name: "missing @", email: "userexample.com", wantErr: true},
        {name: "empty string", email: "", wantErr: true},
        {name: "missing domain", email: "user@", wantErr: true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Act
            err := ValidateEmail(tt.email)
            
            // Assert
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
````

**Don't**:
````go
// BAD: Repetitive individual tests
func TestValidateEmail_Valid(t *testing.T) {
    err := ValidateEmail("user@example.com")
    assert.NoError(t, err)
}

func TestValidateEmail_MissingAt(t *testing.T) {
    err := ValidateEmail("userexample.com")
    assert.Error(t, err)
}

func TestValidateEmail_Empty(t *testing.T) {
    err := ValidateEmail("")
    assert.Error(t, err)
}
// ... many more repetitive functions
````

---

## Mock External Dependencies

**Purpose**: Isolate the system under test from external systems (databases, APIs, filesystem) using test doubles.

**Why**: External dependencies make tests slow, flaky, and complex. Test doubles enable fast, deterministic unit tests.

**Stubs vs Mocks**:
- **Stub**: Provides canned responses for state verification
- **Mock**: Verifies interactions for behavior verification

**Do**:
````go
// GOOD: Using interface for dependency injection
type UserRepository interface {
    GetUserByID(id int) (*User, error)
}

type OrderService struct {
    userRepo UserRepository
}

func TestCreateOrder_WithValidUser_Succeeds(t *testing.T) {
    // Arrange: Use stub for state verification
    stubRepo := &StubUserRepo{
        user: &User{ID: 1, Name: "Alice"},
    }
    service := &OrderService{userRepo: stubRepo}
    
    // Act
    order, err := service.CreateOrder(1, "Product-X")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Alice", order.CustomerName)
}
````

**Don't**:
````go
// BAD: Direct dependency on concrete type
type OrderService struct {
    db *sql.DB // Cannot be mocked
}

func TestCreateOrder(t *testing.T) {
    // Requires real database connection
    db, _ := sql.Open("postgres", "connection-string")
    service := &OrderService{db: db}
    
    order, err := service.CreateOrder(1, "Product-X")
    assert.NoError(t, err)
}
````

---

## Write Descriptive Test Names

**Purpose**: Test name must communicate the unit under test, scenario, and expected behavior.

**Why**: When a test fails in CI, its name is often the first information seen. A clear name enables immediate understanding of what broke.

**Naming Convention**: `Test[Unit]_[Scenario]_[ExpectedBehavior]`

**Do**:
````go
// GOOD: Clear, descriptive names
func TestWithdraw_WithSufficientFunds_ReducesBalance(t *testing.T) { /*...*/ }
func TestWithdraw_WithInsufficientFunds_ReturnsError(t *testing.T) { /*...*/ }
func TestWithdraw_WithZeroAmount_ReturnsValidationError(t *testing.T) { /*...*/ }
func TestDeposit_WithNegativeAmount_ReturnsError(t *testing.T) { /*...*/ }
````

**Don't**:
````go
// BAD: Vague, uninformative names
func TestWithdraw(t *testing.T) { /*...*/ }
func TestWithdraw2(t *testing.T) { /*...*/ }
func TestAccountWorks(t *testing.T) { /*...*/ }
func Test1(t *testing.T) { /*...*/ }
````

---

## Quick Reference Checklist

**Before Writing a Test, Ask:**
- [ ] Am I testing behavior (public API) or implementation (internals)?
- [ ] Will this test pass if I refactor the implementation but keep the same behavior?
- [ ] Is this test isolated? Can it run in any order or in parallel?
- [ ] Does it test exactly one logical concept?
- [ ] Is the AAA structure clear with visual separation?
- [ ] Should this be part of a table-driven test?
- [ ] Are external dependencies mocked/stubbed?
- [ ] Does the test name clearly describe what, when, and what should happen?

**When a Test Fails, Verify:**
1. **Test Name** - Does it match the intended requirement?
2. **Test Implementation** - Is the assertion correct?
3. **Production Code** - Does it have the bug?

If all three disagree, one is wrong. Use this triangulation to debug effectively.
