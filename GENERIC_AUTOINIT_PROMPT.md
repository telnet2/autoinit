# Generic AutoInit Project Structure Prompt Template

A scalable prompt template for code agents to generate enterprise-level Go applications using autoinit as the structuring principle.

## Core Template Structure

```
Create a Go {PROJECT_TYPE} using 'github.com/telnet2/autoinit' following enterprise patterns. 

Architecture: {ARCHITECTURE_PATTERN}
Domain: {BUSINESS_DOMAIN}
Scale: {SCALE_LEVEL}

Required Infrastructure Components:
{INFRASTRUCTURE_COMPONENTS}

Required Business Components:
{BUSINESS_COMPONENTS}

Integration Requirements:
{INTEGRATION_REQUIREMENTS}

Configuration: {CONFIG_PATTERN}
Observability: {OBSERVABILITY_LEVEL}
Error Handling: {ERROR_STRATEGY}

Generate complete project structure following autoinit dependency discovery patterns, lifecycle hooks, and component composition principles. Include main application struct, proper initialization order, and {DEPLOYMENT_TARGET} deployment readiness.
```

## Template Variables Guide

### PROJECT_TYPE Options
- `microservice` - Single-responsibility service
- `monolith` - Multi-domain application  
- `cli-application` - Command-line tool
- `batch-processor` - Data processing pipeline
- `api-gateway` - Service orchestration layer
- `event-processor` - Event-driven system

### ARCHITECTURE_PATTERN Options
- `layered` - Presentation, Business, Data layers
- `hexagonal` - Ports and adapters pattern
- `clean-architecture` - Uncle Bob's clean architecture
- `event-driven` - Event sourcing and CQRS
- `plugin-based` - Extensible plugin system
- `domain-driven` - DDD bounded contexts

### SCALE_LEVEL Options
- `prototype` (1-5 components)
- `production` (5-20 components) 
- `enterprise` (20-100 components)
- `platform` (100+ components with multi-service coordination)

### INFRASTRUCTURE_COMPONENTS Template
```yaml
Database: [{database_type}, connection_pooling, migration_support]
Cache: [{cache_type}, clustering, persistence_options]  
MessageQueue: [{queue_type}, dead_letter_handling, retry_logic]
Logging: [structured_logging, log_aggregation, correlation_ids]
Metrics: [prometheus_metrics, custom_metrics, health_checks]
Tracing: [distributed_tracing, span_context, trace_correlation]
Security: [authentication, authorization, audit_logging, encryption]
Storage: [{storage_type}, backup_strategies, data_retention]
Network: [load_balancing, circuit_breakers, rate_limiting]
Scheduler: [cron_jobs, task_queues, job_persistence]
```

### BUSINESS_COMPONENTS Template  
```yaml
# Domain-specific - customize per project
UserManagement: [registration, profile_management, preferences]
ProductCatalog: [inventory, pricing, categories, search]
OrderProcessing: [cart, checkout, payment_integration, fulfillment]
NotificationService: [email, sms, push_notifications, templates]
ReportingService: [analytics, dashboards, data_export, scheduled_reports]
ContentManagement: [cms, media_handling, versioning, workflow]
```

### INTEGRATION_REQUIREMENTS Options
- `rest_apis` - HTTP REST interfaces
- `grpc_services` - gRPC communication
- `message_brokers` - Async messaging patterns
- `webhook_handlers` - Event webhooks processing  
- `third_party_apis` - External service integration
- `database_integration` - Multi-database support
- `file_processing` - File upload/processing pipelines

### CONFIG_PATTERN Options
- `yaml_driven` - YAML configuration with autoinit integration
- `environment_based` - 12-factor app environment variables
- `consul_integration` - Distributed configuration management
- `feature_flags` - Dynamic feature toggling
- `multi_environment` - Dev/staging/prod configuration layers

### OBSERVABILITY_LEVEL Options
- `basic` - Logs and basic metrics
- `intermediate` - Structured logging, metrics, health checks
- `advanced` - Full observability stack with tracing, alerting
- `enterprise` - Complete observability with SLA monitoring, incident response

### ERROR_STRATEGY Options
- `fail_fast` - Quick failure detection and propagation
- `graceful_degradation` - Service degradation under failure
- `circuit_breaker` - Prevent cascade failures
- `retry_with_backoff` - Intelligent retry mechanisms
- `error_aggregation` - Centralized error collection and analysis

## Example: Enterprise E-commerce Platform

```
Create a Go microservice using 'github.com/telnet2/autoinit' following enterprise patterns.

Architecture: clean-architecture
Domain: e-commerce
Scale: enterprise

Required Infrastructure Components:
Database: [postgresql, connection_pooling, migration_support]
Cache: [redis, clustering, persistence_options]
MessageQueue: [rabbitmq, dead_letter_handling, retry_logic]
Logging: [structured_logging, log_aggregation, correlation_ids]
Metrics: [prometheus_metrics, custom_metrics, health_checks]
Security: [jwt_authentication, rbac_authorization, audit_logging, tls_encryption]

Required Business Components:
UserManagement: [registration, profile_management, preferences]
ProductCatalog: [inventory, pricing, categories, search]
OrderProcessing: [cart, checkout, payment_integration, fulfillment]
NotificationService: [email, sms, push_notifications, templates]

Integration Requirements: rest_apis, message_brokers, third_party_apis, webhook_handlers
Configuration: yaml_driven
Observability: advanced  
Error Handling: circuit_breaker

Generate complete project structure following autoinit dependency discovery patterns, lifecycle hooks, and component composition principles. Include main application struct, proper initialization order, and kubernetes deployment readiness.
```

## Advanced Enterprise Extensions

### Multi-Service Coordination
For platform-scale applications:

```yaml
Service_Discovery: [consul, service_mesh, load_balancing]
Inter_Service_Communication: [grpc, service_contracts, versioning]
Data_Consistency: [saga_pattern, event_sourcing, eventual_consistency]  
Deployment: [blue_green, canary_releases, feature_toggles]
Monitoring: [service_topology, dependency_mapping, sla_monitoring]
```

### Component Dependency Patterns

```yaml
# Autoinit-specific patterns to include
Dependency_Discovery: [autoinit.As(), interface_resolution, component_registry]
Lifecycle_Management: [PreInit, PostInit, graceful_shutdown]
Configuration_Injection: [yaml_binding, environment_override, secret_management]
Health_Management: [component_health, dependency_health, readiness_probes]
Plugin_Architecture: [dynamic_loading, component_registration, hot_reloading]
```

## Prompt Engineering Best Practices

### 1. Specificity Layers
- **Layer 1**: Architecture and scale decisions
- **Layer 2**: Component specifications  
- **Layer 3**: Integration and deployment requirements
- **Layer 4**: Quality and operational requirements

### 2. Progressive Complexity
- Start with core application structure
- Add infrastructure components
- Layer in business logic components
- Apply cross-cutting concerns (security, observability)

### 3. Validation Criteria
Include explicit success criteria:
```yaml
Generated_Code_Must:
  - Use autoinit.AutoInit() for component initialization
  - Implement proper dependency discovery with autoinit.As()
  - Include lifecycle hooks for complex initialization
  - Support graceful shutdown and cleanup
  - Be production-ready with proper error handling
  - Include comprehensive logging and metrics
  - Support configuration via YAML and environment variables
  - Include health check endpoints
  - Follow Go best practices and idioms
```

## Usage Instructions

1. **Select Template Variables**: Choose appropriate values for your project
2. **Customize Components**: Add/remove components based on domain requirements  
3. **Apply to Code Agent**: Use the filled template as prompt
4. **Iterate and Refine**: Adjust based on generated output quality
5. **Validate Against Criteria**: Ensure generated code meets enterprise standards

This template approach enables consistent, scalable application generation while maintaining autoinit's declarative philosophy across projects of any complexity level.