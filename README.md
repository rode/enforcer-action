# evaluate-policy-action

A GitHub Action to evaluate a policy against a specified resource.
It's intended to be used along with [Rode](https://github.com/rode/rode) in order to gate deployments.

## Use

```yaml
  - name: Evaluate Policy
    uses: rode/evaluate-policy-action@v0.1.0
    with:
      policyName: Sample Harbor Policy
      resourceUri: harbor.localhost/rode-demo/rode-demo-node-app@sha256:54221980d01768efc835708f037a716a11a6f2f7f9633c948896a7f39f859775
      rodeHost: rode.rode-demo.svc.cluster.local:50051
```

### Inputs

| Input          | Description                                                                                              | Default |
|----------------|----------------------------------------------------------------------------------------------------------|---------|
| `enforce`      | Indicates that failing evaluation should fail the build                                                  | `true`  |
| `policyId`     | The policy id to evaluate against. Exactly one of policyName or policyId should be set                   | N/A     |
| `policyName`   | Name of the policy to evaluate against the resource. Exactly one of policyName or policyId should be set | N/A     |
| `resourceUri`  | The resource to evaluate policy against                                                                  | N/A     |
| `rodeHost`     | Hostname of the Rode instance                                                                            | N/A     |
| `rodeInsecure` | When set, the connection to Rode will not use TLS                                                        | `false` |

### Outputs

| Output | Description                                 |
|--------|---------------------------------------------|
| `pass` | The boolean result of the policy evaluation |

## Local Development

1. Configuration for the action is sourced from the environment, the easiest way to run locally is to set the following environment variables,
   or place them in a file called `.env`:
    ```
    ENFORCE=true
    POLICY_ID=cd84b7e2-8fbf-429c-97b7-6b2d7ce9b64a
    RESOURCE_URI=harbor.localhost/rode-demo/rode-demo-node-app@sha256:54221980d01768efc835708f037a716a11a6f2f7f9633c948896a7f39f859775
    RODE_HOST=rode.rode-demo.svc.cluster.local:50051
    RODE_INSECURE=true
    ```
1. Then `env $(cat .env | xargs) go run main.go` or simply `go run` if the variables are already set
1. Fix any formatting issues with `make fmt`
1. Run the tests with `make test`