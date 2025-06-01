# ffuf API Testing Documentation

This directory contains comprehensive documentation for using ffuf as an API testing tool. The documentation covers various aspects of API testing, from basic usage to advanced techniques.

## Contents

- **[API Testing Guide](guide.md)** - Complete guide covering all aspects of API testing with ffuf
- **[API Testing Workflows](workflows.md)** - Practical examples and workflows for common API testing scenarios
- **[API Testing Cheat Sheet](cheatsheet.md)** - Quick reference for commands, techniques, and best practices
- **[Integration Guide](integration.md)** - Documentation on integrating ffuf with other API testing tools and workflows

## Getting Started

If you're new to API testing with ffuf, start with the [API Testing Guide](guide.md) to learn the basics. Then, explore the [API Testing Workflows](workflows.md) for practical examples of how to use ffuf for API testing in real-world scenarios.

For quick reference during testing, refer to the [API Testing Cheat Sheet](cheatsheet.md).

If you're looking to integrate ffuf with other tools in your API testing workflow, check out the [Integration Guide](integration.md).

## Documentation Structure

### For Beginners

1. Start with the [API Testing Guide](guide.md) - covers basic concepts and setup
2. Try the examples in [API Testing Workflows](workflows.md) - practical hands-on scenarios
3. Keep the [API Testing Cheat Sheet](cheatsheet.md) handy for quick reference

### For Advanced Users

- Explore advanced techniques in the [API Testing Guide](guide.md)
- Learn about tool integrations in the [Integration Guide](integration.md)
- Contribute your own workflows to [API Testing Workflows](workflows.md)

## Quick Examples

### REST API Testing

```bash
# Basic endpoint discovery
ffuf -api-mode -u https://api.example.com/FUZZ

# JSON parameter fuzzing
ffuf -w params.txt -u https://api.example.com/endpoint -X POST -H "Content-Type: application/json" -d '{"FUZZ":"value"}'
```

### GraphQL API Testing

```bash
# GraphQL query fuzzing
ffuf -w graphql_queries.txt -u https://api.example.com/graphql -X POST -H "Content-Type: application/json" -d '{"query":"FUZZ"}'
```

### Authentication Testing

```bash
# JWT token testing
ffuf -w endpoints.txt -u https://api.example.com/FUZZ -H "Authorization: Bearer TOKEN" -mc all -fc 401,403
```

## Additional Resources

- **Main Documentation**: See the main [README](../../README.md) for general ffuf usage
- **Configuration**: Learn about [configuration files](../../README.md#configuration-files) for API testing setups
- **Community**: Join discussions about API testing techniques
  in [GitHub Discussions](https://github.com/ffuf/ffuf/discussions)

## Contributing to API Documentation

We welcome contributions to improve the API testing documentation:

1. **Examples**: Submit real-world API testing examples
2. **Techniques**: Share new API testing techniques or improvements
3. **Integrations**: Document integrations with new tools or platforms
4. **Clarifications**: Help improve existing documentation clarity

For contribution guidelines, see the main project [CONTRIBUTORS.md](../../CONTRIBUTORS.md).
