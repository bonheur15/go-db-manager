# Contributing to Go DB Manager

We welcome contributions to the Go DB Manager project! To ensure a smooth and collaborative experience for everyone, please follow these guidelines.

## How to Contribute

1.  **Fork the repository:** Start by forking the `go-db-manager` repository to your GitHub account.
2.  **Clone your forked repository:**
    ```bash
    git clone https://github.com/your-username/go-db-manager.git
    cd go-db-manager
    ```
3.  **Create a new branch:** For each new feature or bug fix, create a new branch. Use a descriptive name for your branch (e.g., `feature/add-user-auth`, `bugfix/fix-mysql-connection`).
    ```bash
    git checkout -b your-branch-name
    ```
4.  **Make your changes:** Implement your feature or bug fix. Ensure your code adheres to the existing coding style and conventions.
5.  **Write tests:** If you're adding new functionality, please write unit and/or integration tests to cover your changes. If you're fixing a bug, add a test that reproduces the bug and then passes with your fix.
6.  **Run tests and linters:** Before committing, make sure all existing tests pass and that your code passes linting checks.
    ```bash
    go test ./...
    go fmt ./...
    go vet ./...
    ```
7.  **Commit your changes:** Write clear and concise commit messages. A good commit message explains *what* changed and *why*.
    ```bash
    git commit -m "feat: Add new feature" # or "fix: Fix bug in X"
    ```
8.  **Push your branch:**
    ```bash
    git push origin your-branch-name
    ```
9.  **Create a Pull Request (PR):** Go to the original `go-db-manager` repository on GitHub and open a new Pull Request from your forked repository. Provide a clear description of your changes and reference any related issues.

## Code Style

*   Follow the standard Go formatting conventions (`go fmt`).
*   Keep functions concise and focused on a single responsibility.
*   Use meaningful variable and function names.
*   Add comments where the code's intent is not immediately obvious.

## Reporting Bugs

If you find a bug, please open an issue on GitHub. Provide as much detail as possible, including:

*   A clear and concise description of the bug.
*   Steps to reproduce the behavior.
*   Expected behavior.
*   Actual behavior.
*   Any relevant error messages or logs.
*   Your operating system and Go version.

## Suggesting Enhancements

We welcome ideas for new features or improvements. Please open an issue on GitHub to suggest an enhancement. Describe the feature, why it would be useful, and any potential implementation details.

## Pull Request Checklist

Before submitting your pull request, please ensure you have:

*   [ ] Forked the repository and created a new branch.
*   [ ] Implemented your changes and written appropriate tests.
*   [ ] Run `go test ./...`, `go fmt ./...`, and `go vet ./...`.
*   [ ] Written clear, concise commit messages.
*   [ ] Provided a detailed description in your pull request.
*   [ ] Referenced any related issues.

Thank you for contributing!