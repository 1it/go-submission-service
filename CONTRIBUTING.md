# Contributing to Go Submission Service

We love your input! We want to make contributing to Go Submission Service as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

## Pull Requests

Pull requests are the best way to propose changes to the codebase. We actively welcome your pull requests:

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

## Any contributions you make will be under the MIT Software License

In short, when you submit code changes, your submissions are understood to be under the same [MIT License](http://choosealicense.com/licenses/mit/) that covers the project. Feel free to contact the maintainers if that's a concern.

## Report bugs using GitHub's [issue tracker](https://github.com/1it/go-submission-service/issues)

We use GitHub issues to track public bugs. Report a bug by [opening a new issue](https://github.com/1it/go-submission-service/issues/new).

## Write bug reports with detail, background, and sample code

**Great Bug Reports** tend to have:

- A quick summary and/or background
- Steps to reproduce
  - Be specific!
  - Give sample code if you can
- What you expected would happen
- What actually happens
- Notes (possibly including why you think this might be happening, or stuff you tried that didn't work)

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/1it/go-submission-service.git
   cd go-submission-service
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run tests:
   ```bash
   go test ./...
   ```

4. Run the service locally:
   ```bash
   make run
   ```

## Code Style

- Use `gofmt` to format your code
- Follow Go naming conventions
- Write clear, self-documenting code
- Add comments for complex logic
- Keep functions small and focused

## Testing

- Write unit tests for new functionality
- Ensure all tests pass before submitting PR
- Include integration tests for API endpoints
- Test error conditions and edge cases

## License

By contributing, you agree that your contributions will be licensed under its MIT License.