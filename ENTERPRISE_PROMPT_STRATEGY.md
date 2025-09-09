# Enterprise AutoInit Prompt Strategy for Code Agents

Based on testing with Coco, here's a refined strategy for generating enterprise-level autoinit applications using code agents.

## Key Findings from Testing

### ✅ What Code Agents Handle Well
- **Clean Architecture**: Generated proper layered structure (cmd, internal, domain)
- **Configuration Management**: Created comprehensive YAML-driven config
- **Enterprise Patterns**: Graceful shutdown, signal handling, dependency injection concepts
- **Project Structure**: Proper Go project layout with multiple packages
- **Domain Modeling**: Created entities, repositories, and service layers

### ❌ Current Limitations  
- **API Accuracy**: Invented non-existent autoinit APIs (`autoinit.Container`, `autoinit.As[T]`)
- **Timeout Issues**: Complex prompts timeout after 90-120 seconds
- **Incomplete Generation**: May not generate all requested components
- **Dependency Discovery**: Doesn't understand actual autoinit.As() patterns

## Recommended Strategy: Layered Prompt Decomposition

Instead of one complex prompt, use a **multi-stage generation approach**:

### Stage 1: Architecture Foundation
**Prompt**: Generate project structure, configuration, and main application skeleton

```
Create Go project structure for {DOMAIN} microservice using 'github.com/telnet2/autoinit'. Generate:
1) Clean architecture layout (cmd/, internal/, configs/)
2) Main application struct with autoinit.AutoInit() integration  
3) YAML configuration management
4) Basic HTTP server with health endpoint
5) Graceful shutdown handling
Single main.go with proper autoinit usage patterns.
```

### Stage 2: Infrastructure Components
**Prompt**: Add infrastructure layer components one by one

```
Add {COMPONENT} component to existing autoinit application:
1) Create {COMPONENT} struct with Init(context.Context) method
2) Add to main application struct with autoinit tag
3) Include configuration section in YAML
4) Add health check integration
Show only the component files and updated main application struct.
```

### Stage 3: Business Components  
**Prompt**: Add business logic with dependency discovery

```
Add {BUSINESS_COMPONENT} with autoinit dependency discovery:
1) Create service with Init(ctx, parent) method
2) Use autoinit.As(ctx, self, parent, &dependency) for finding {DEPENDENCIES}
3) Include business logic methods
4) Add to main application composition
Show component implementation with proper autoinit patterns.
```

### Stage 4: Integration Layer
**Prompt**: Add HTTP handlers, middleware, and external integrations

```
Add REST API layer to autoinit application:
1) HTTP handlers with dependency injection from autoinit
2) Middleware for logging, metrics, authentication
3) OpenAPI documentation generation
4) Integration with business services via autoinit discovery
Show handler implementations and routing setup.
```

## Refined Generic Template for Stage 1

```
Create Go {PROJECT_TYPE} project structure using 'github.com/telnet2/autoinit' following clean architecture.

Domain: {BUSINESS_DOMAIN}
Scale: {SCALE_LEVEL}

Core Requirements:
1) Project layout: cmd/main.go, internal/app/, internal/config/, configs/app.yaml
2) Main App struct with these embedded components:
   - Config *config.Config
   - Logger *logger.Component  
   - Database *database.Component
   - HTTPServer *http.ServerComponent

3) Each component implements Init(context.Context) method
4) Use autoinit.AutoInit(ctx, app) in main()
5) YAML configuration for all components
6) Health endpoint at /health returning component status
7) Graceful shutdown with signal handling
8) Production-ready error handling and logging

Generate complete, runnable single-file version first, then suggest multi-package structure.
```

## Component Template Library

### Infrastructure Components
```yaml
Database:
  - PostgreSQL: connection_pooling, migrations, health_checks
  - Redis: clustering, persistence, connection_management
  - MongoDB: replica_sets, indexing, connection_pooling

Messaging:
  - Kafka: consumer_groups, producer_configs, offset_management
  - RabbitMQ: exchanges, queues, dead_letter_handling
  - NATS: subjects, clustering, jetstream

Observability:
  - Prometheus: custom_metrics, histogram, counter, gauge
  - Jaeger: distributed_tracing, span_management
  - Structured_Logging: correlation_ids, log_levels, output_formats
```

### Business Components
```yaml
Authentication:
  - JWT: token_validation, refresh_tokens, blacklisting
  - OAuth2: provider_integration, scope_management
  - Session: session_storage, timeout_handling

User_Management:
  - Registration: email_verification, password_policies
  - Profile: profile_updates, preferences, privacy_settings
  - Permissions: role_based_access, resource_permissions
```

## Progressive Complexity Strategy

### Level 1: Prototype (1-3 components)
- Single main.go file
- Basic HTTP server
- Simple database integration
- **Success Rate**: 90%

### Level 2: Production (5-10 components)
- Multi-package structure
- Infrastructure + business components
- Configuration management
- **Success Rate**: 70%

### Level 3: Enterprise (10+ components)  
- Layered prompt approach
- Stage-by-stage generation
- Manual integration and testing
- **Success Rate**: 50% per stage, 80% overall with manual integration

## Quality Validation Checklist

After each generation stage:
```yaml
Code_Quality:
  - ✅ Uses actual autoinit.AutoInit() API (not invented APIs)
  - ✅ Components implement correct Init() signatures
  - ✅ Proper error handling and context usage
  - ✅ Go best practices and idioms followed

Architecture:
  - ✅ Clean separation of concerns
  - ✅ Dependency injection via autoinit patterns
  - ✅ Configuration externalized to YAML
  - ✅ Health checks and observability included

Enterprise_Readiness:
  - ✅ Graceful shutdown handling
  - ✅ Production logging and metrics
  - ✅ Error aggregation and monitoring
  - ✅ Security best practices followed
```

## Implementation Workflow

1. **Start Small**: Generate single-file prototype with core components
2. **Validate Foundation**: Ensure autoinit integration works correctly  
3. **Add Incrementally**: Use stage-by-stage prompts to add complexity
4. **Manual Integration**: Code agent output often needs manual refinement
5. **Test Continuously**: Validate each stage before proceeding
6. **Document Patterns**: Build organization-specific component library

This approach transforms autoinit from a simple dependency injection tool into a **enterprise application generation platform** while working within current code agent limitations.