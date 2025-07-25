# Agents

## Get DERP map updates

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/derp-map \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /derp-map`

### Responses

| Status | Meaning                                                                  | Description         | Schema |
|--------|--------------------------------------------------------------------------|---------------------|--------|
| 101    | [Switching Protocols](https://tools.ietf.org/html/rfc7231#section-6.2.2) | Switching Protocols |        |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## User-scoped tailnet RPC connection

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/tailnet \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /tailnet`

### Responses

| Status | Meaning                                                                  | Description         | Schema |
|--------|--------------------------------------------------------------------------|---------------------|--------|
| 101    | [Switching Protocols](https://tools.ietf.org/html/rfc7231#section-6.2.2) | Switching Protocols |        |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Authenticate agent on AWS instance

### Code samples

```shell
# Example request using curl
curl -X POST http://coder-server:8080/api/v2/workspaceagents/aws-instance-identity \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`POST /workspaceagents/aws-instance-identity`

> Body parameter

```json
{
  "document": "string",
  "signature": "string"
}
```

### Parameters

| Name   | In   | Type                                                                             | Required | Description             |
|--------|------|----------------------------------------------------------------------------------|----------|-------------------------|
| `body` | body | [agentsdk.AWSInstanceIdentityToken](schemas.md#agentsdkawsinstanceidentitytoken) | true     | Instance identity token |

### Example responses

> 200 Response

```json
{
  "session_token": "string"
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                   |
|--------|---------------------------------------------------------|-------------|--------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [agentsdk.AuthenticateResponse](schemas.md#agentsdkauthenticateresponse) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Authenticate agent on Azure instance

### Code samples

```shell
# Example request using curl
curl -X POST http://coder-server:8080/api/v2/workspaceagents/azure-instance-identity \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`POST /workspaceagents/azure-instance-identity`

> Body parameter

```json
{
  "encoding": "string",
  "signature": "string"
}
```

### Parameters

| Name   | In   | Type                                                                                 | Required | Description             |
|--------|------|--------------------------------------------------------------------------------------|----------|-------------------------|
| `body` | body | [agentsdk.AzureInstanceIdentityToken](schemas.md#agentsdkazureinstanceidentitytoken) | true     | Instance identity token |

### Example responses

> 200 Response

```json
{
  "session_token": "string"
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                   |
|--------|---------------------------------------------------------|-------------|--------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [agentsdk.AuthenticateResponse](schemas.md#agentsdkauthenticateresponse) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Authenticate agent on Google Cloud instance

### Code samples

```shell
# Example request using curl
curl -X POST http://coder-server:8080/api/v2/workspaceagents/google-instance-identity \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`POST /workspaceagents/google-instance-identity`

> Body parameter

```json
{
  "json_web_token": "string"
}
```

### Parameters

| Name   | In   | Type                                                                                   | Required | Description             |
|--------|------|----------------------------------------------------------------------------------------|----------|-------------------------|
| `body` | body | [agentsdk.GoogleInstanceIdentityToken](schemas.md#agentsdkgoogleinstanceidentitytoken) | true     | Instance identity token |

### Example responses

> 200 Response

```json
{
  "session_token": "string"
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                   |
|--------|---------------------------------------------------------|-------------|--------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [agentsdk.AuthenticateResponse](schemas.md#agentsdkauthenticateresponse) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Patch workspace agent app status

### Code samples

```shell
# Example request using curl
curl -X PATCH http://coder-server:8080/api/v2/workspaceagents/me/app-status \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`PATCH /workspaceagents/me/app-status`

> Body parameter

```json
{
  "app_slug": "string",
  "icon": "string",
  "message": "string",
  "needs_user_attention": true,
  "state": "working",
  "uri": "string"
}
```

### Parameters

| Name   | In   | Type                                                         | Required | Description |
|--------|------|--------------------------------------------------------------|----------|-------------|
| `body` | body | [agentsdk.PatchAppStatus](schemas.md#agentsdkpatchappstatus) | true     | app status  |

### Example responses

> 200 Response

```json
{
  "detail": "string",
  "message": "string",
  "validations": [
    {
      "detail": "string",
      "field": "string"
    }
  ]
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                           |
|--------|---------------------------------------------------------|-------------|--------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [codersdk.Response](schemas.md#codersdkresponse) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Get workspace agent external auth

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/me/external-auth?match=string&id=string \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/me/external-auth`

### Parameters

| Name     | In    | Type    | Required | Description                       |
|----------|-------|---------|----------|-----------------------------------|
| `match`  | query | string  | true     | Match                             |
| `id`     | query | string  | true     | Provider ID                       |
| `listen` | query | boolean | false    | Wait for a new token to be issued |

### Example responses

> 200 Response

```json
{
  "access_token": "string",
  "password": "string",
  "token_extra": {},
  "type": "string",
  "url": "string",
  "username": "string"
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                   |
|--------|---------------------------------------------------------|-------------|--------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [agentsdk.ExternalAuthResponse](schemas.md#agentsdkexternalauthresponse) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Removed: Get workspace agent git auth

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/me/gitauth?match=string&id=string \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/me/gitauth`

### Parameters

| Name     | In    | Type    | Required | Description                       |
|----------|-------|---------|----------|-----------------------------------|
| `match`  | query | string  | true     | Match                             |
| `id`     | query | string  | true     | Provider ID                       |
| `listen` | query | boolean | false    | Wait for a new token to be issued |

### Example responses

> 200 Response

```json
{
  "access_token": "string",
  "password": "string",
  "token_extra": {},
  "type": "string",
  "url": "string",
  "username": "string"
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                   |
|--------|---------------------------------------------------------|-------------|--------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [agentsdk.ExternalAuthResponse](schemas.md#agentsdkexternalauthresponse) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Get workspace agent Git SSH key

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/me/gitsshkey \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/me/gitsshkey`

### Example responses

> 200 Response

```json
{
  "private_key": "string",
  "public_key": "string"
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                             |
|--------|---------------------------------------------------------|-------------|----------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [agentsdk.GitSSHKey](schemas.md#agentsdkgitsshkey) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Post workspace agent log source

### Code samples

```shell
# Example request using curl
curl -X POST http://coder-server:8080/api/v2/workspaceagents/me/log-source \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`POST /workspaceagents/me/log-source`

> Body parameter

```json
{
  "display_name": "string",
  "icon": "string",
  "id": "string"
}
```

### Parameters

| Name   | In   | Type                                                                     | Required | Description        |
|--------|------|--------------------------------------------------------------------------|----------|--------------------|
| `body` | body | [agentsdk.PostLogSourceRequest](schemas.md#agentsdkpostlogsourcerequest) | true     | Log source request |

### Example responses

> 200 Response

```json
{
  "created_at": "2019-08-24T14:15:22Z",
  "display_name": "string",
  "icon": "string",
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "workspace_agent_id": "7ad2e618-fea7-4c1a-b70a-f501566a72f1"
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                         |
|--------|---------------------------------------------------------|-------------|--------------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [codersdk.WorkspaceAgentLogSource](schemas.md#codersdkworkspaceagentlogsource) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Patch workspace agent logs

### Code samples

```shell
# Example request using curl
curl -X PATCH http://coder-server:8080/api/v2/workspaceagents/me/logs \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`PATCH /workspaceagents/me/logs`

> Body parameter

```json
{
  "log_source_id": "string",
  "logs": [
    {
      "created_at": "string",
      "level": "trace",
      "output": "string"
    }
  ]
}
```

### Parameters

| Name   | In   | Type                                               | Required | Description |
|--------|------|----------------------------------------------------|----------|-------------|
| `body` | body | [agentsdk.PatchLogs](schemas.md#agentsdkpatchlogs) | true     | logs        |

### Example responses

> 200 Response

```json
{
  "detail": "string",
  "message": "string",
  "validations": [
    {
      "detail": "string",
      "field": "string"
    }
  ]
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                           |
|--------|---------------------------------------------------------|-------------|--------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [codersdk.Response](schemas.md#codersdkresponse) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Get workspace agent reinitialization

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/me/reinit \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/me/reinit`

### Example responses

> 200 Response

```json
{
  "reason": "prebuild_claimed",
  "workspaceID": "string"
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                     |
|--------|---------------------------------------------------------|-------------|----------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [agentsdk.ReinitializationEvent](schemas.md#agentsdkreinitializationevent) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Get workspace agent by ID

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/{workspaceagent} \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/{workspaceagent}`

### Parameters

| Name             | In   | Type         | Required | Description        |
|------------------|------|--------------|----------|--------------------|
| `workspaceagent` | path | string(uuid) | true     | Workspace agent ID |

### Example responses

> 200 Response

```json
{
  "api_version": "string",
  "apps": [
    {
      "command": "string",
      "display_name": "string",
      "external": true,
      "group": "string",
      "health": "disabled",
      "healthcheck": {
        "interval": 0,
        "threshold": 0,
        "url": "string"
      },
      "hidden": true,
      "icon": "string",
      "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
      "open_in": "slim-window",
      "sharing_level": "owner",
      "slug": "string",
      "statuses": [
        {
          "agent_id": "2b1e3b65-2c04-4fa2-a2d7-467901e98978",
          "app_id": "affd1d10-9538-4fc8-9e0b-4594a28c1335",
          "created_at": "2019-08-24T14:15:22Z",
          "icon": "string",
          "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
          "message": "string",
          "needs_user_attention": true,
          "state": "working",
          "uri": "string",
          "workspace_id": "0967198e-ec7b-4c6b-b4d3-f71244cadbe9"
        }
      ],
      "subdomain": true,
      "subdomain_name": "string",
      "url": "string"
    }
  ],
  "architecture": "string",
  "connection_timeout_seconds": 0,
  "created_at": "2019-08-24T14:15:22Z",
  "directory": "string",
  "disconnected_at": "2019-08-24T14:15:22Z",
  "display_apps": [
    "vscode"
  ],
  "environment_variables": {
    "property1": "string",
    "property2": "string"
  },
  "expanded_directory": "string",
  "first_connected_at": "2019-08-24T14:15:22Z",
  "health": {
    "healthy": false,
    "reason": "agent has lost connection"
  },
  "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
  "instance_id": "string",
  "last_connected_at": "2019-08-24T14:15:22Z",
  "latency": {
    "property1": {
      "latency_ms": 0,
      "preferred": true
    },
    "property2": {
      "latency_ms": 0,
      "preferred": true
    }
  },
  "lifecycle_state": "created",
  "log_sources": [
    {
      "created_at": "2019-08-24T14:15:22Z",
      "display_name": "string",
      "icon": "string",
      "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
      "workspace_agent_id": "7ad2e618-fea7-4c1a-b70a-f501566a72f1"
    }
  ],
  "logs_length": 0,
  "logs_overflowed": true,
  "name": "string",
  "operating_system": "string",
  "parent_id": {
    "uuid": "string",
    "valid": true
  },
  "ready_at": "2019-08-24T14:15:22Z",
  "resource_id": "4d5215ed-38bb-48ed-879a-fdb9ca58522f",
  "scripts": [
    {
      "cron": "string",
      "display_name": "string",
      "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
      "log_path": "string",
      "log_source_id": "4197ab25-95cf-4b91-9c78-f7f2af5d353a",
      "run_on_start": true,
      "run_on_stop": true,
      "script": "string",
      "start_blocks_login": true,
      "timeout": 0
    }
  ],
  "started_at": "2019-08-24T14:15:22Z",
  "startup_script_behavior": "blocking",
  "status": "connecting",
  "subsystems": [
    "envbox"
  ],
  "troubleshooting_url": "string",
  "updated_at": "2019-08-24T14:15:22Z",
  "version": "string"
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                       |
|--------|---------------------------------------------------------|-------------|--------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [codersdk.WorkspaceAgent](schemas.md#codersdkworkspaceagent) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Get connection info for workspace agent

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/{workspaceagent}/connection \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/{workspaceagent}/connection`

### Parameters

| Name             | In   | Type         | Required | Description        |
|------------------|------|--------------|----------|--------------------|
| `workspaceagent` | path | string(uuid) | true     | Workspace agent ID |

### Example responses

> 200 Response

```json
{
  "derp_force_websockets": true,
  "derp_map": {
    "homeParams": {
      "regionScore": {
        "property1": 0,
        "property2": 0
      }
    },
    "omitDefaultRegions": true,
    "regions": {
      "property1": {
        "avoid": true,
        "embeddedRelay": true,
        "nodes": [
          {
            "canPort80": true,
            "certName": "string",
            "derpport": 0,
            "forceHTTP": true,
            "hostName": "string",
            "insecureForTests": true,
            "ipv4": "string",
            "ipv6": "string",
            "name": "string",
            "regionID": 0,
            "stunonly": true,
            "stunport": 0,
            "stuntestIP": "string"
          }
        ],
        "regionCode": "string",
        "regionID": 0,
        "regionName": "string"
      },
      "property2": {
        "avoid": true,
        "embeddedRelay": true,
        "nodes": [
          {
            "canPort80": true,
            "certName": "string",
            "derpport": 0,
            "forceHTTP": true,
            "hostName": "string",
            "insecureForTests": true,
            "ipv4": "string",
            "ipv6": "string",
            "name": "string",
            "regionID": 0,
            "stunonly": true,
            "stunport": 0,
            "stuntestIP": "string"
          }
        ],
        "regionCode": "string",
        "regionID": 0,
        "regionName": "string"
      }
    }
  },
  "disable_direct_connections": true,
  "hostname_suffix": "string"
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                         |
|--------|---------------------------------------------------------|-------------|--------------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [workspacesdk.AgentConnectionInfo](schemas.md#workspacesdkagentconnectioninfo) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Get running containers for workspace agent

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/{workspaceagent}/containers?label=string \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/{workspaceagent}/containers`

### Parameters

| Name             | In    | Type              | Required | Description        |
|------------------|-------|-------------------|----------|--------------------|
| `workspaceagent` | path  | string(uuid)      | true     | Workspace agent ID |
| `label`          | query | string(key=value) | true     | Labels             |

### Example responses

> 200 Response

```json
{
  "containers": [
    {
      "created_at": "2019-08-24T14:15:22Z",
      "id": "string",
      "image": "string",
      "labels": {
        "property1": "string",
        "property2": "string"
      },
      "name": "string",
      "ports": [
        {
          "host_ip": "string",
          "host_port": 0,
          "network": "string",
          "port": 0
        }
      ],
      "running": true,
      "status": "string",
      "volumes": {
        "property1": "string",
        "property2": "string"
      }
    }
  ],
  "devcontainers": [
    {
      "agent": {
        "directory": "string",
        "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
        "name": "string"
      },
      "config_path": "string",
      "container": {
        "created_at": "2019-08-24T14:15:22Z",
        "id": "string",
        "image": "string",
        "labels": {
          "property1": "string",
          "property2": "string"
        },
        "name": "string",
        "ports": [
          {
            "host_ip": "string",
            "host_port": 0,
            "network": "string",
            "port": 0
          }
        ],
        "running": true,
        "status": "string",
        "volumes": {
          "property1": "string",
          "property2": "string"
        }
      },
      "dirty": true,
      "error": "string",
      "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
      "name": "string",
      "status": "running",
      "workspace_folder": "string"
    }
  ],
  "warnings": [
    "string"
  ]
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                                                   |
|--------|---------------------------------------------------------|-------------|----------------------------------------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [codersdk.WorkspaceAgentListContainersResponse](schemas.md#codersdkworkspaceagentlistcontainersresponse) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Recreate devcontainer for workspace agent

### Code samples

```shell
# Example request using curl
curl -X POST http://coder-server:8080/api/v2/workspaceagents/{workspaceagent}/containers/devcontainers/{devcontainer}/recreate \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`POST /workspaceagents/{workspaceagent}/containers/devcontainers/{devcontainer}/recreate`

### Parameters

| Name             | In   | Type         | Required | Description        |
|------------------|------|--------------|----------|--------------------|
| `workspaceagent` | path | string(uuid) | true     | Workspace agent ID |
| `devcontainer`   | path | string       | true     | Devcontainer ID    |

### Example responses

> 202 Response

```json
{
  "detail": "string",
  "message": "string",
  "validations": [
    {
      "detail": "string",
      "field": "string"
    }
  ]
}
```

### Responses

| Status | Meaning                                                       | Description | Schema                                           |
|--------|---------------------------------------------------------------|-------------|--------------------------------------------------|
| 202    | [Accepted](https://tools.ietf.org/html/rfc7231#section-6.3.3) | Accepted    | [codersdk.Response](schemas.md#codersdkresponse) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Watch workspace agent for container updates

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/{workspaceagent}/containers/watch \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/{workspaceagent}/containers/watch`

### Parameters

| Name             | In   | Type         | Required | Description        |
|------------------|------|--------------|----------|--------------------|
| `workspaceagent` | path | string(uuid) | true     | Workspace agent ID |

### Example responses

> 200 Response

```json
{
  "containers": [
    {
      "created_at": "2019-08-24T14:15:22Z",
      "id": "string",
      "image": "string",
      "labels": {
        "property1": "string",
        "property2": "string"
      },
      "name": "string",
      "ports": [
        {
          "host_ip": "string",
          "host_port": 0,
          "network": "string",
          "port": 0
        }
      ],
      "running": true,
      "status": "string",
      "volumes": {
        "property1": "string",
        "property2": "string"
      }
    }
  ],
  "devcontainers": [
    {
      "agent": {
        "directory": "string",
        "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
        "name": "string"
      },
      "config_path": "string",
      "container": {
        "created_at": "2019-08-24T14:15:22Z",
        "id": "string",
        "image": "string",
        "labels": {
          "property1": "string",
          "property2": "string"
        },
        "name": "string",
        "ports": [
          {
            "host_ip": "string",
            "host_port": 0,
            "network": "string",
            "port": 0
          }
        ],
        "running": true,
        "status": "string",
        "volumes": {
          "property1": "string",
          "property2": "string"
        }
      },
      "dirty": true,
      "error": "string",
      "id": "497f6eca-6276-4993-bfeb-53cbbbba6f08",
      "name": "string",
      "status": "running",
      "workspace_folder": "string"
    }
  ],
  "warnings": [
    "string"
  ]
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                                                   |
|--------|---------------------------------------------------------|-------------|----------------------------------------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [codersdk.WorkspaceAgentListContainersResponse](schemas.md#codersdkworkspaceagentlistcontainersresponse) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Coordinate workspace agent

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/{workspaceagent}/coordinate \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/{workspaceagent}/coordinate`

### Parameters

| Name             | In   | Type         | Required | Description        |
|------------------|------|--------------|----------|--------------------|
| `workspaceagent` | path | string(uuid) | true     | Workspace agent ID |

### Responses

| Status | Meaning                                                                  | Description         | Schema |
|--------|--------------------------------------------------------------------------|---------------------|--------|
| 101    | [Switching Protocols](https://tools.ietf.org/html/rfc7231#section-6.2.2) | Switching Protocols |        |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Get listening ports for workspace agent

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/{workspaceagent}/listening-ports \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/{workspaceagent}/listening-ports`

### Parameters

| Name             | In   | Type         | Required | Description        |
|------------------|------|--------------|----------|--------------------|
| `workspaceagent` | path | string(uuid) | true     | Workspace agent ID |

### Example responses

> 200 Response

```json
{
  "ports": [
    {
      "network": "string",
      "port": 0,
      "process_name": "string"
    }
  ]
}
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                                                   |
|--------|---------------------------------------------------------|-------------|----------------------------------------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | [codersdk.WorkspaceAgentListeningPortsResponse](schemas.md#codersdkworkspaceagentlisteningportsresponse) |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Get logs by workspace agent

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/{workspaceagent}/logs \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/{workspaceagent}/logs`

### Parameters

| Name             | In    | Type         | Required | Description                                  |
|------------------|-------|--------------|----------|----------------------------------------------|
| `workspaceagent` | path  | string(uuid) | true     | Workspace agent ID                           |
| `before`         | query | integer      | false    | Before log id                                |
| `after`          | query | integer      | false    | After log id                                 |
| `follow`         | query | boolean      | false    | Follow log stream                            |
| `no_compression` | query | boolean      | false    | Disable compression for WebSocket connection |

### Example responses

> 200 Response

```json
[
  {
    "created_at": "2019-08-24T14:15:22Z",
    "id": 0,
    "level": "trace",
    "output": "string",
    "source_id": "ae50a35c-df42-4eff-ba26-f8bc28d2af81"
  }
]
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                      |
|--------|---------------------------------------------------------|-------------|-----------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | array of [codersdk.WorkspaceAgentLog](schemas.md#codersdkworkspaceagentlog) |

<h3 id="get-logs-by-workspace-agent-responseschema">Response Schema</h3>

Status Code **200**

| Name           | Type                                             | Required | Restrictions | Description |
|----------------|--------------------------------------------------|----------|--------------|-------------|
| `[array item]` | array                                            | false    |              |             |
| `» created_at` | string(date-time)                                | false    |              |             |
| `» id`         | integer                                          | false    |              |             |
| `» level`      | [codersdk.LogLevel](schemas.md#codersdkloglevel) | false    |              |             |
| `» output`     | string                                           | false    |              |             |
| `» source_id`  | string(uuid)                                     | false    |              |             |

#### Enumerated Values

| Property | Value   |
|----------|---------|
| `level`  | `trace` |
| `level`  | `debug` |
| `level`  | `info`  |
| `level`  | `warn`  |
| `level`  | `error` |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Open PTY to workspace agent

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/{workspaceagent}/pty \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/{workspaceagent}/pty`

### Parameters

| Name             | In   | Type         | Required | Description        |
|------------------|------|--------------|----------|--------------------|
| `workspaceagent` | path | string(uuid) | true     | Workspace agent ID |

### Responses

| Status | Meaning                                                                  | Description         | Schema |
|--------|--------------------------------------------------------------------------|---------------------|--------|
| 101    | [Switching Protocols](https://tools.ietf.org/html/rfc7231#section-6.2.2) | Switching Protocols |        |

To perform this operation, you must be authenticated. [Learn more](authentication.md).

## Removed: Get logs by workspace agent

### Code samples

```shell
# Example request using curl
curl -X GET http://coder-server:8080/api/v2/workspaceagents/{workspaceagent}/startup-logs \
  -H 'Accept: application/json' \
  -H 'Coder-Session-Token: API_KEY'
```

`GET /workspaceagents/{workspaceagent}/startup-logs`

### Parameters

| Name             | In    | Type         | Required | Description                                  |
|------------------|-------|--------------|----------|----------------------------------------------|
| `workspaceagent` | path  | string(uuid) | true     | Workspace agent ID                           |
| `before`         | query | integer      | false    | Before log id                                |
| `after`          | query | integer      | false    | After log id                                 |
| `follow`         | query | boolean      | false    | Follow log stream                            |
| `no_compression` | query | boolean      | false    | Disable compression for WebSocket connection |

### Example responses

> 200 Response

```json
[
  {
    "created_at": "2019-08-24T14:15:22Z",
    "id": 0,
    "level": "trace",
    "output": "string",
    "source_id": "ae50a35c-df42-4eff-ba26-f8bc28d2af81"
  }
]
```

### Responses

| Status | Meaning                                                 | Description | Schema                                                                      |
|--------|---------------------------------------------------------|-------------|-----------------------------------------------------------------------------|
| 200    | [OK](https://tools.ietf.org/html/rfc7231#section-6.3.1) | OK          | array of [codersdk.WorkspaceAgentLog](schemas.md#codersdkworkspaceagentlog) |

<h3 id="removed:-get-logs-by-workspace-agent-responseschema">Response Schema</h3>

Status Code **200**

| Name           | Type                                             | Required | Restrictions | Description |
|----------------|--------------------------------------------------|----------|--------------|-------------|
| `[array item]` | array                                            | false    |              |             |
| `» created_at` | string(date-time)                                | false    |              |             |
| `» id`         | integer                                          | false    |              |             |
| `» level`      | [codersdk.LogLevel](schemas.md#codersdkloglevel) | false    |              |             |
| `» output`     | string                                           | false    |              |             |
| `» source_id`  | string(uuid)                                     | false    |              |             |

#### Enumerated Values

| Property | Value   |
|----------|---------|
| `level`  | `trace` |
| `level`  | `debug` |
| `level`  | `info`  |
| `level`  | `warn`  |
| `level`  | `error` |

To perform this operation, you must be authenticated. [Learn more](authentication.md).
