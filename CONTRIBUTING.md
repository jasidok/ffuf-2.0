# Contributing to ffuf

We welcome contributions to ffuf! This document provides guidelines for contributing to the project.

## Ways to Contribute

### 1. Code Contributions

- **Bug fixes**: Fix issues reported in GitHub Issues
- **Feature development**: Implement new features from the roadmap
- **Performance improvements**: Optimize existing functionality
- **Testing**: Add or improve test coverage

### 2. Documentation

- **User guides**: Improve existing documentation or create new guides
- **API documentation**: Document new APIs or improve existing docs
- **Examples**: Add practical examples and use cases
- **Translations**: Help translate documentation

### 3. Community Support

- **Answer questions**: Help users in GitHub Discussions
- **Bug triage**: Help validate and reproduce bug reports
- **Feature discussion**: Participate in feature planning discussions
- **Code review**: Review pull requests from other contributors

### 4. Testing and Quality Assurance

- **Manual testing**: Test new features and report issues
- **Automated testing**: Improve test coverage and CI/CD
- **Performance testing**: Benchmark and profile performance
- **Security testing**: Help identify security issues

## Getting Started

### Development Environment Setup

1. **Fork and clone the repository**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/ffuf.git
   cd ffuf
   ```

2. **Install Go** (version 1.16 or later):
   ```bash
   # Check Go version
   go version
   
   # Install dependencies
   go mod download
   ```

3. **Build and test**:
   ```bash
   # Build the project
   go build
   
   # Run tests
   go test ./...
   
   # Run specific package tests
   go test ./pkg/ffuf
   ```

4. **Verify your setup**:
   ```bash
   # Test the built binary
   ./ffuf -V
   ```

### Project Structure

```
ffuf/
├── main.go              # Main entry point
├── help.go              # Help text and usage information
├── pkg/                 # Core packages
│   ├── ffuf/           # Core ffuf functionality
│   ├── api/            # API testing features
│   ├── filter/         # Response filtering
│   ├── input/          # Input handling (wordlists, etc.)
│   ├── output/         # Output formatting
│   ├── runner/         # HTTP request execution
│   └── interactive/    # Interactive mode
├── docs/               # Documentation
│   ├── api/           # API testing documentation
│   ├── GETTING_STARTED.md
│   └── FAQ.md
└── _img/              # Images and assets
```

## Contribution Guidelines

### Code Style

1. **Follow Go conventions**:
    - Use `gofmt` to format your code
    - Follow effective Go practices
    - Write clear, self-documenting code

2. **Documentation**:
    - Add godoc comments for all exported functions
    - Document complex internal functions
    - Update relevant documentation files

3. **Testing**:
    - Write tests for new functionality
    - Ensure all tests pass before submitting
    - Maintain or improve test coverage

### Commit Guidelines

1. **Commit message format**:
   ```
   type(scope): description
   
   Longer explanation if needed
   
   Fixes #issue_number
   ```

2. **Types**:
    - `feat`: New feature
    - `fix`: Bug fix
    - `docs`: Documentation changes
    - `style`: Code style changes (formatting, etc.)
    - `refactor`: Code refactoring
    - `test`: Adding or updating tests
    - `chore`: Build process or auxiliary tool changes

3. **Examples**:
   ```
   feat(api): add GraphQL query fuzzing support
   fix(filter): handle empty response body correctly
   docs(api): add examples for JWT authentication
   ```

### Pull Request Process

1. **Before submitting**:
    - Ensure your code follows the style guidelines
    - Run all tests and ensure they pass
    - Update documentation if needed
    - Test your changes thoroughly

2. **Pull request description**:
    - Clearly describe what your PR does
    - Reference any related issues
    - Include examples of how to test your changes
    - Add screenshots for UI changes

3. **Review process**:
    - Be responsive to feedback
    - Make requested changes promptly
    - Discuss any disagreements constructively
    - Ensure CI checks pass

### Issue Guidelines

#### Bug Reports

When reporting bugs, include:

1. **ffuf version**: Output of `ffuf -V`
2. **Operating system**: OS and version
3. **Command used**: The exact command that caused the issue
4. **Expected behavior**: What you expected to happen
5. **Actual behavior**: What actually happened
6. **Reproduction steps**: Step-by-step instructions to reproduce
7. **Additional context**: Any other relevant information

**Example bug report**:

```markdown
**ffuf version**: v2.1.0
**OS**: Ubuntu 20.04
**Command**: `ffuf -w wordlist.txt -u https://example.com/FUZZ -mc all`

**Expected**: Should show all responses regardless of status code
**Actual**: Only shows 200 responses

**Steps to reproduce**:
1. Run the command above
2. Notice only 200 responses are shown
3. Expected 404, 500, etc. responses as well

**Additional context**: This worked in v2.0.0
```

#### Feature Requests

When requesting features, include:

1. **Problem description**: What problem are you trying to solve?
2. **Proposed solution**: How would you like it to work?
3. **Use case**: Provide specific examples of when this would be useful
4. **Alternatives**: Have you considered any alternative solutions?

**Example feature request**:

```markdown
**Problem**: It's difficult to test APIs that require custom authentication schemes

**Proposed solution**: Add support for custom authentication plugins

**Use case**: 
- Testing APIs with proprietary auth schemes
- Supporting enterprise authentication systems
- Extending auth without modifying core code

**Alternatives**: 
- Currently using custom headers, but it's not flexible enough
- Could write external scripts, but integration would be poor
```

## Specific Contribution Areas

### API Testing Features

We're particularly interested in contributions to API testing functionality:

1. **New authentication methods**: OAuth, SAML, custom schemes
2. **Protocol support**: WebSocket, gRPC, more GraphQL features
3. **Response parsing**: Better JSON/XML parsing and filtering
4. **Integration**: Tools like Postman, Swagger, OpenAPI
5. **Security testing**: OWASP API Top 10 checks

### Performance Improvements

Areas where performance contributions are welcome:

1. **HTTP client optimization**: Connection pooling, HTTP/2
2. **Memory usage**: Reduce allocations, improve garbage collection
3. **Concurrency**: Better worker pool implementation
4. **I/O optimization**: Faster file reading, better buffering

### Documentation Improvements

Documentation contributions we need:

1. **More examples**: Real-world use cases and scenarios
2. **Video tutorials**: Screen recordings of common workflows
3. **Best practices**: Security considerations, performance tips
4. **Integration guides**: Using ffuf with other security tools

## Code Review Process

### For Contributors

1. **Self-review**: Review your own PR before requesting review
2. **Tests**: Ensure all tests pass and add new tests if needed
3. **Documentation**: Update docs for any user-facing changes
4. **Backwards compatibility**: Avoid breaking existing functionality

### For Reviewers

1. **Be constructive**: Focus on the code, not the person
2. **Explain reasoning**: Help contributors understand feedback
3. **Suggest alternatives**: Don't just point out problems
4. **Approve promptly**: Don't let good PRs sit waiting

## Release Process

1. **Feature freeze**: No new features in release candidates
2. **Testing**: Thorough testing of release candidates
3. **Documentation**: Update all relevant documentation
4. **Changelog**: Maintain detailed changelog
5. **Versioning**: Follow semantic versioning

## Community Guidelines

### Code of Conduct

We follow a simple code of conduct:

1. **Be respectful**: Treat everyone with respect and professionalism
2. **Be inclusive**: Welcome contributors of all backgrounds and skill levels
3. **Be constructive**: Focus on helping the project and community grow
4. **Be patient**: Remember that everyone is learning

### Communication

1. **GitHub Issues**: For bug reports and feature requests
2. **GitHub Discussions**: For questions and general discussion
3. **Pull Requests**: For code contributions and reviews
4. **Email**: For private or security-related matters

## Recognition

We recognize contributors in several ways:

1. **Contributors file**: All contributors are listed in CONTRIBUTORS.md
2. **Release notes**: Significant contributions are mentioned in releases
3. **GitHub recognition**: We use GitHub's contributor recognition features

## Questions?

If you have questions about contributing:

1. **Check existing issues**: Someone may have asked the same question
2. **GitHub Discussions**: Ask in the community discussions
3. **Documentation**: Review this guide and other docs
4. **Direct contact**: Reach out to maintainers for sensitive matters

Thank you for contributing to ffuf! Your efforts help make the tool better for the entire security community.