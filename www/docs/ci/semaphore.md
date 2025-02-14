# Semaphore

In [Semaphore 2.0](https://semaphoreci.com) each project starts with the
default pipeline specified in `.semaphore/semaphore.yml`.

```yaml
# .semaphore/semaphore.yml.
version: v1.0
name: Build
agent:
  machine:
    type: e1-standard-2
    os_image: ubuntu1804

blocks:
  - name: "Test"
    task:
      prologue:
        commands:
          # set go version
          - sem-version go 1.11
          - "export GOPATH=~/go"
          - "export PATH=/home/semaphore/go/bin:$PATH"
          - checkout

      jobs:
        - name: "Lint"
          commands:
            - go get ./...
            - go test ./...

# On Semaphore 2.0 deployment and delivery is managed with promotions,
# which may be automatic or manual and optionally depend on conditions.
promotions:
    - name: Release
       pipeline_file: goreleaser.yaml
       auto_promote_on:
         - result: passed
           branch:
             - "^refs/tags/v*"
```

Pipeline file in `.semaphore/goreleaser.yaml`:

```yaml
version: "v1.0"
name: GoReleaser
agent:
  machine:
    type: e1-standard-2
    os_image: ubuntu1804
blocks:
  - name: "Release"
    task:
      secrets:
        - name: goreleaser
      prologue:
        commands:
          - sem-version go 1.11
          - "export GOPATH=~/go"
          - "export PATH=/home/semaphore/go/bin:$PATH"
          - checkout
      jobs:
        - name: goreleaser
          commands:
            - curl -sfL https://goreleaser.com/static/run | bash
```

The following YAML file, `createSecret.yml` creates a new secret item that is
called GoReleaser with one environment variable, named `GITHUB_TOKEN`:

```yaml
apiVersion: v1alpha
kind: Secret
metadata:
  name: goreleaser
data:
  env_vars:
    - name: GITHUB_TOKEN
      value: "your token here"
```

Check [Managing Secrets](https://docs.semaphoreci.com/using-semaphore/secrets)
for more detailed documentation.
