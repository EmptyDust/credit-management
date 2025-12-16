#!/bin/bash

# Test Runner Script for Credit Management System
# This script provides an easy way to run tests across all services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_THRESHOLD=60
DOCKER_REQUIRED=true

# Functions
print_header() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

check_requirements() {
    print_header "Checking Requirements"

    # Check Go
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed"
        exit 1
    fi
    print_success "Go is installed: $(go version)"

    # Check Docker
    if [ "$DOCKER_REQUIRED" = true ]; then
        if ! command -v docker &> /dev/null; then
            print_error "Docker is not installed"
            exit 1
        fi

        if ! docker info &> /dev/null; then
            print_error "Docker daemon is not running"
            exit 1
        fi
        print_success "Docker is running"
    fi

    # Check test directories
    for dir in auth-service/tests credit-activity-service/tests test-utils; do
        if [ ! -d "$dir" ]; then
            print_warning "Directory $dir not found"
        else
            print_success "Directory $dir exists"
        fi
    done

    echo ""
}

run_service_tests() {
    local service=$1
    local test_dir="${service}/tests"

    print_header "Running ${service} Tests"

    if [ ! -d "$test_dir" ]; then
        print_warning "No tests found for $service"
        return 0
    fi

    cd "$test_dir"

    if go test -v -timeout 10m ./...; then
        print_success "${service} tests passed"
        cd - > /dev/null
        return 0
    else
        print_error "${service} tests failed"
        cd - > /dev/null
        return 1
    fi
}

run_all_tests() {
    print_header "Running All Tests"

    local failed=0

    run_service_tests "auth-service" || ((failed++))
    run_service_tests "credit-activity-service" || ((failed++))

    echo ""
    if [ $failed -eq 0 ]; then
        print_success "All tests passed!"
        return 0
    else
        print_error "$failed service(s) failed tests"
        return 1
    fi
}

run_coverage() {
    print_header "Running Tests with Coverage"

    mkdir -p coverage

    # Auth service coverage
    echo "Generating coverage for auth-service..."
    cd auth-service
    go test -coverprofile=../coverage/auth-coverage.out ./... || true
    cd ..

    # Activity service coverage
    echo "Generating coverage for credit-activity-service..."
    cd credit-activity-service
    go test -coverprofile=../coverage/activity-coverage.out ./... || true
    cd ..

    # Generate coverage reports
    print_header "Coverage Summary"

    for service in auth activity; do
        if [ -f "coverage/${service}-coverage.out" ]; then
            echo ""
            echo "${service}-service:"
            go tool cover -func=coverage/${service}-coverage.out | tail -n 1

            # Check if coverage meets threshold
            coverage=$(go tool cover -func=coverage/${service}-coverage.out | grep total | awk '{print $3}' | sed 's/%//')
            if (( $(echo "$coverage >= $COVERAGE_THRESHOLD" | bc -l) )); then
                print_success "Coverage ($coverage%) meets threshold ($COVERAGE_THRESHOLD%)"
            else
                print_warning "Coverage ($coverage%) below threshold ($COVERAGE_THRESHOLD%)"
            fi
        fi
    done

    # Generate HTML reports
    echo ""
    print_header "Generating HTML Coverage Reports"

    for service in auth activity; do
        if [ -f "coverage/${service}-coverage.out" ]; then
            go tool cover -html=coverage/${service}-coverage.out -o coverage/${service}-coverage.html
            print_success "Generated coverage/${service}-coverage.html"
        fi
    done
}

run_specific_test() {
    local service=$1
    local test_name=$2

    print_header "Running Specific Test: $test_name in $service"

    cd "${service}/tests"
    go test -v -run "$test_name" ./...
    cd - > /dev/null
}

install_dependencies() {
    print_header "Installing Dependencies"

    for dir in test-utils auth-service credit-activity-service user-service; do
        if [ -d "$dir" ]; then
            echo "Installing dependencies for $dir..."
            cd "$dir"
            go mod tidy
            go mod download
            cd ..
            print_success "Dependencies installed for $dir"
        fi
    done
}

clean_artifacts() {
    print_header "Cleaning Test Artifacts"

    rm -rf coverage/
    find . -name "*.test" -delete
    find . -name "*.out" -delete

    print_success "Cleaned test artifacts"
}

show_usage() {
    cat << EOF
Usage: $0 [COMMAND] [OPTIONS]

Commands:
    all             Run all tests (default)
    auth            Run auth-service tests only
    activity        Run credit-activity-service tests only
    user            Run user-service tests only
    coverage        Run tests with coverage report
    specific        Run a specific test (requires -s SERVICE -t TEST)
    install         Install test dependencies
    clean           Clean test artifacts
    verify          Verify test setup
    help            Show this help message

Options:
    -s SERVICE      Specify service (auth-service, credit-activity-service)
    -t TEST         Specify test name (e.g., TestLoginWithUsername)
    -v              Verbose output
    --no-docker     Skip Docker requirement check

Examples:
    $0 all                                    # Run all tests
    $0 auth                                   # Run auth service tests
    $0 coverage                               # Run with coverage
    $0 specific -s auth-service -t TestLogin  # Run specific test
    $0 verify                                 # Verify setup
EOF
}

# Main script logic
main() {
    local command="${1:-all}"
    shift || true

    # Parse options
    while [ $# -gt 0 ]; do
        case "$1" in
            --no-docker)
                DOCKER_REQUIRED=false
                shift
                ;;
            -s)
                SERVICE="$2"
                shift 2
                ;;
            -t)
                TEST="$2"
                shift 2
                ;;
            -v)
                set -x
                shift
                ;;
            *)
                echo "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done

    # Execute command
    case "$command" in
        all)
            check_requirements
            run_all_tests
            ;;
        auth)
            check_requirements
            run_service_tests "auth-service"
            ;;
        activity)
            check_requirements
            run_service_tests "credit-activity-service"
            ;;
        user)
            check_requirements
            run_service_tests "user-service"
            ;;
        coverage)
            check_requirements
            run_coverage
            ;;
        specific)
            if [ -z "$SERVICE" ] || [ -z "$TEST" ]; then
                print_error "Service and test name required for specific test"
                echo "Usage: $0 specific -s SERVICE -t TEST"
                exit 1
            fi
            check_requirements
            run_specific_test "$SERVICE" "$TEST"
            ;;
        install)
            install_dependencies
            ;;
        clean)
            clean_artifacts
            ;;
        verify)
            check_requirements
            ;;
        help)
            show_usage
            ;;
        *)
            echo "Unknown command: $command"
            show_usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
