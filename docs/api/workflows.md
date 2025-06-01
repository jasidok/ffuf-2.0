# API Testing Workflows with ffuf

This document provides example workflows for common API testing scenarios using ffuf. These workflows demonstrate how to use ffuf's features to effectively test APIs in real-world situations.

## Table of Contents

1. [API Endpoint Discovery](#api-endpoint-discovery)
2. [API Authentication Testing](#api-authentication-testing)
3. [API Parameter Fuzzing](#api-parameter-fuzzing)
4. [API Security Testing](#api-security-testing)
5. [GraphQL API Testing](#graphql-api-testing)
6. [API Performance Testing](#api-performance-testing)
7. [API Documentation Testing](#api-documentation-testing)

## API Endpoint Discovery

### Scenario
You have discovered an API and want to find all available endpoints.

### Workflow

1. **Basic Endpoint Discovery**

   ```bash
   # Use a wordlist of common API endpoints
   ffuf -u https://api.example.com/FUZZ -w /path/to/api_wordlist.txt -H "Content-Type: application/json"
   ```

2. **Recursive Endpoint Discovery**

   ```bash
   # Discover endpoints and then recursively fuzz discovered paths
   ffuf -u https://api.example.com/FUZZ -w /path/to/api_wordlist.txt -recursion -recursion-depth 2 -H "Content-Type: application/json"
   ```

3. **Version-Based Endpoint Discovery**

   ```bash
   # Test different API versions
   ffuf -u https://api.example.com/vFUZZ/users -w /path/to/versions.txt -H "Content-Type: application/json"
   ```

4. **Interactive Exploration**

   ```bash
   # Start in API mode
   ffuf -u https://api.example.com -api_mode
   
   # Press ENTER to enter interactive mode, then:
   > api start
   > method GET
   > url https://api.example.com
   > path /v1/users
   > send
   
   # Visualize the API structure
   > map html
   ```

## API Authentication Testing

### Scenario
You need to test different authentication methods for an API.

### Workflow

1. **API Key Testing**

   ```bash
   # Test different API keys
   ffuf -u https://api.example.com/v1/users -H "X-API-Key: FUZZ" -w /path/to/api_keys.txt -fc 401
   ```

2. **OAuth Token Testing**

   ```bash
   # Test different OAuth tokens
   ffuf -u https://api.example.com/v1/users -H "Authorization: Bearer FUZZ" -w /path/to/tokens.txt -fc 401
   ```

3. **Interactive Authentication Testing**

   ```bash
   # Start in API mode
   ffuf -u https://api.example.com -api_mode
   
   # Press ENTER to enter interactive mode, then:
   > api start
   > method POST
   > url https://api.example.com
   > path /v1/auth/login
   > header Content-Type application/json
   > body json {"username":"admin","password":"password"}
   > send
   
   # Extract token from response and use it
   > header Authorization Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   > path /v1/users
   > method GET
   > send
   ```

## API Parameter Fuzzing

### Scenario
You want to test different parameters and values for an API endpoint.

### Workflow

1. **Query Parameter Fuzzing**

   ```bash
   # Fuzz query parameters
   ffuf -u https://api.example.com/v1/search?FUZZ=test -w /path/to/params.txt
   
   # Fuzz parameter values
   ffuf -u https://api.example.com/v1/search?query=FUZZ -w /path/to/values.txt
   ```

2. **JSON Body Parameter Fuzzing**

   ```bash
   # Fuzz JSON body parameters
   ffuf -u https://api.example.com/v1/users -X POST -H "Content-Type: application/json" -d '{"FUZZ":"test"}' -w /path/to/params.txt
   
   # Fuzz JSON body values
   ffuf -u https://api.example.com/v1/users -X POST -H "Content-Type: application/json" -d '{"name":"FUZZ"}' -w /path/to/names.txt
   ```

3. **Multiple Parameter Fuzzing**

   ```bash
   # Fuzz multiple parameters simultaneously
   ffuf -u https://api.example.com/v1/search -X POST -H "Content-Type: application/json" -d '{"query":"FUZZW","filter":"FUZZX"}' -w /path/to/queries.txt:FUZZW -w /path/to/filters.txt:FUZZX
   ```

4. **Interactive Parameter Testing**

   ```bash
   # Start in API mode
   ffuf -u https://api.example.com -api_mode
   
   # Press ENTER to enter interactive mode, then:
   > api start
   > method POST
   > url https://api.example.com
   > path /v1/search
   > header Content-Type application/json
   > body query test
   > body filter category
   > send
   
   # Try different parameters
   > body clear
   > body json {"query":"test","sort":"date","limit":10}
   > send
   
   # Compare responses
   > diff
   ```

## API Security Testing

### Scenario
You want to test an API for common security vulnerabilities.

### Workflow

1. **BOLA/IDOR Testing**

   ```bash
   # Test for Broken Object Level Authorization
   ffuf -u https://api.example.com/v1/users/FUZZ/profile -w /path/to/ids.txt -H "Authorization: Bearer USER_A_TOKEN" -mc all
   ```

2. **Injection Testing**

   ```bash
   # Test for SQL injection
   ffuf -u https://api.example.com/v1/search -X POST -H "Content-Type: application/json" -d '{"query":"FUZZ"}' -w /path/to/sql_injections.txt -fr "error"
   
   # Test for NoSQL injection
   ffuf -u https://api.example.com/v1/users -X POST -H "Content-Type: application/json" -d '{"username":{"$ne":"FUZZ"}}' -w /path/to/nosql_payloads.txt
   ```

3. **Mass Assignment Testing**

   ```bash
   # Test for mass assignment vulnerabilities
   ffuf -u https://api.example.com/v1/users -X POST -H "Content-Type: application/json" -d '{"name":"John","email":"john@example.com","FUZZ":"value"}' -w /path/to/privileged_params.txt
   ```

4. **Rate Limiting Testing**

   ```bash
   # Test for rate limiting
   ffuf -u https://api.example.com/v1/login -X POST -H "Content-Type: application/json" -d '{"username":"admin","password":"password"}' -rate 50
   ```

## GraphQL API Testing

### Scenario
You need to test a GraphQL API.

### Workflow

1. **GraphQL Introspection**

   ```bash
   # Perform GraphQL introspection
   ffuf -u https://api.example.com/graphql -X POST -H "Content-Type: application/json" -d '{"query":"{ __schema { types { name fields { name } } } }"}' -fr "error"
   ```

2. **GraphQL Query Fuzzing**

   ```bash
   # Fuzz GraphQL queries
   ffuf -u https://api.example.com/graphql -X POST -H "Content-Type: application/json" -d '{"query":"{ user(id: \"FUZZ\") { name email } }"}' -w /path/to/ids.txt
   ```

3. **GraphQL Mutation Testing**

   ```bash
   # Test GraphQL mutations
   ffuf -u https://api.example.com/graphql -X POST -H "Content-Type: application/json" -d '{"query":"mutation { updateUser(id: \"1\", input: { name: \"FUZZ\" }) { id name } }"}' -w /path/to/names.txt
   ```

4. **Interactive GraphQL Testing**

   ```bash
   # Start in API mode
   ffuf -u https://api.example.com -api_mode
   
   # Press ENTER to enter interactive mode, then:
   > api start
   > method POST
   > url https://api.example.com
   > path /graphql
   > header Content-Type application/json
   > body json {"query":"{ user(id: \"1\") { name email } }"}
   > send
   
   # Try a different query
   > body json {"query":"{ users { id name } }"}
   > send
   
   # Compare responses
   > diff
   ```

## API Performance Testing

### Scenario
You want to test the performance of an API under load.

### Workflow

1. **Basic Performance Testing**

   ```bash
   # Test with high concurrency
   ffuf -u https://api.example.com/v1/users -X GET -H "Authorization: Bearer YOUR_TOKEN" -c 50
   ```

2. **Rate-Limited Testing**

   ```bash
   # Test with specific request rate
   ffuf -u https://api.example.com/v1/users -X GET -H "Authorization: Bearer YOUR_TOKEN" -rate 10
   ```

3. **Timeout Testing**

   ```bash
   # Test with different timeouts
   ffuf -u https://api.example.com/v1/users -X GET -H "Authorization: Bearer YOUR_TOKEN" -timeout 5
   ```

4. **Response Time Analysis**

   ```bash
   # Filter by response time
   ffuf -u https://api.example.com/v1/users/FUZZ -w /path/to/ids.txt -H "Authorization: Bearer YOUR_TOKEN" -ft 500
   ```

## API Documentation Testing

### Scenario
You want to test an API against its documentation (OpenAPI/Swagger).

### Workflow

1. **OpenAPI Specification Testing**

   ```bash
   # First, parse the OpenAPI specification
   ffuf -u https://api.example.com/openapi.json -X GET -o openapi.json
   
   # Then test endpoints from the specification
   ffuf -u https://api.example.com/FUZZ -w openapi_endpoints.txt -H "Authorization: Bearer YOUR_TOKEN"
   ```

2. **Interactive Documentation Testing**

   ```bash
   # Start in API mode
   ffuf -u https://api.example.com -api_mode
   
   # Press ENTER to enter interactive mode, then:
   > api start
   > method GET
   > url https://api.example.com
   > path /openapi.json
   > send
   
   # The response contains the API documentation
   # Now test endpoints from the documentation
   > path /v1/users
   > send
   
   # Visualize the API structure
   > map html
   ```

## Conclusion

These workflows demonstrate how to use ffuf for various API testing scenarios. You can adapt and combine these workflows to suit your specific testing needs.

For more detailed information on ffuf's API testing capabilities, refer to the [API Testing Guide](guide.md).