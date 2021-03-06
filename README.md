# enforcer-action

A GitHub Action to gate deployments using [Rode](https://github.com/rode/rode).

## Use

Note that versions of this action prior to v0.3.0 are available as `rode/evaluate-policy-action`.

```yaml
  - name: Rode Enforcer
    uses: rode/enforcer-action@v0.3.0
    with:
      githubToken: ${{ secrets.GITHUB_TOKEN }}
      policyGroup: prod
      resourceUri: harbor.localhost/rode-demo/rode-demo-node-app@sha256:54221980d01768efc835708f037a716a11a6f2f7f9633c948896a7f39f859775
      rodeHost: rode.rode-demo.svc.cluster.local:50051
```

If the event type is `pull_request` or `pull_request_target`, the action will post a comment containing evaluation results on the pull request.

### Inputs

| Input          | Description                                                                                                            | Default |
|----------------|------------------------------------------------------------------------------------------------------------------------|---------|
| `accessToken`  | An access token that will be included in requests to Rode. Can be omitted if Rode isn't configured for authentication. | N/A     |
| `enforce`      | Controls whether the step should fail if the evaluation fails.                                                         | `true`  |
| `githubToken`  | A GitHub access token used to comment on pull requests. `${{ secrets.GITHUB_TOKEN }}` has the necessary permissions.   | N/A     |
| `policyGroup`  | The policy group to evaluate the resource against.                                                                     | N/A     |
| `resourceUri`  | The resource to evaluate policies against.                                                                             | N/A     |
| `rodeHost`     | Hostname of the Rode instance                                                                                          | N/A     |
| `rodeInsecure` | Disables transport security when communicating with Rode.                                                              | `false` |

### GitHub Environment

These settings are taken from the default GitHub Actions environment, but can also be set with environment variables or flags for local testing.

| Name                | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| `GITHUB_SERVER_URL` | URL of the GitHub instance                                                  |
| `GITHUB_REPOSITORY` | Repository slug of the form `${OWNER}/${REPO}`                              |
| `GITHUB_RUN_ID`     | The run id of the workflow.                                                 |
| `GITHUB_EVENT_NAME` | Name of the event that triggered the workflow.                              |
| `GITHUB_EVENT_PATH` | Absolute path to the JSON payload of the event that triggered the workflow. |

### Outputs

| Output       | Description                                 |
|--------------|---------------------------------------------|
| `pass`       | The boolean result of the policy evaluation |
| `reportPath` | A path to a summary of evaluation results   |


## Local Development

1. Run the action locally, configuring it with flags or environment variables: 
    ```shell
    go run main.go \
      --policy-group=prod \
      --resource-uri=harbor.localhost/rode-demo/rode-demo-node-app@sha256:54221980d01768efc835708f037a716a11a6f2f7f9633c948896a7f39f859775 \
      --rode-host=rode.rode-demo.svc.cluster.local:50051 \
      --rode-insecure-disable-transport-security \
      --enforce
    ```
1. Fix any formatting issues with `make fmt`
1. Run the tests with `make test`
