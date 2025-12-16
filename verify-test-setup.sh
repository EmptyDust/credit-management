#!/bin/bash

echo "=================================="
echo "Test Infrastructure Verification"
echo "=================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✓${NC} $1"
        return 0
    else
        echo -e "${RED}✗${NC} $1 (missing)"
        return 1
    fi
}

check_dir() {
    if [ -d "$1" ]; then
        echo -e "${GREEN}✓${NC} $1/"
        return 0
    else
        echo -e "${RED}✗${NC} $1/ (missing)"
        return 1
    fi
}

passed=0
failed=0

echo "Test Utilities:"
check_dir "test-utils" && ((passed++)) || ((failed++))
check_file "test-utils/database.go" && ((passed++)) || ((failed++))
check_file "test-utils/fixtures.go" && ((passed++)) || ((failed++))
check_file "test-utils/assertions.go" && ((passed++)) || ((failed++))
check_file "test-utils/helpers.go" && ((passed++)) || ((failed++))
check_file "test-utils/README.md" && ((passed++)) || ((failed++))
check_file "test-utils/go.mod" && ((passed++)) || ((failed++))

echo ""
echo "Auth Service Tests:"
check_dir "auth-service/tests" && ((passed++)) || ((failed++))
check_file "auth-service/tests/auth_test.go" && ((passed++)) || ((failed++))

echo ""
echo "Activity Service Tests:"
check_dir "credit-activity-service/tests" && ((passed++)) || ((failed++))
check_file "credit-activity-service/tests/activity_test.go" && ((passed++)) || ((failed++))
check_file "credit-activity-service/tests/participant_test.go" && ((passed++)) || ((failed++))
check_file "credit-activity-service/tests/attachment_test.go" && ((passed++)) || ((failed++))

echo ""
echo "Test Automation:"
check_file "Makefile" && ((passed++)) || ((failed++))
check_file "run-tests.sh" && ((passed++)) || ((failed++))
check_dir ".github/workflows" && ((passed++)) || ((failed++))
check_file ".github/workflows/test.yml" && ((passed++)) || ((failed++))

echo ""
echo "Documentation:"
check_file "TESTING.md" && ((passed++)) || ((failed++))
check_file "TEST_SUMMARY.md" && ((passed++)) || ((failed++))

echo ""
echo "=================================="
echo "Summary: ${GREEN}${passed} passed${NC}, ${RED}${failed} failed${NC}"
echo "=================================="

if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✓ Test infrastructure is complete!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some components are missing${NC}"
    exit 1
fi
