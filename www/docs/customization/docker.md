# Docker Images

GoReleaser can build and push Docker images.
Let's see how it works.

## How it works

You can declare multiple Docker images.
They will be matched against the binaries generated by your `builds` section and
packages generated by your `nfpms` section.

If you have only one item in the `builds` list,
the configuration can be as easy as adding the
name of your image to your `.goreleaser.yaml` file:

```yaml
dockers:
  - image_templates:
      - user/repo
```

<!-- md:templates -->

You also need to create a `Dockerfile` in your project's root directory:

```dockerfile
FROM scratch
ENTRYPOINT ["/mybin"]
COPY mybin /
```

This configuration will build and push a Docker image named `user/repo:tagname`.

### The Docker build context

Note that we are not building any go files in the Docker build phase, we are
merely copying the binary to a `scratch` image and setting up the `entrypoint`.

The idea is that you reuse the previously built binaries instead of building
them again when creating the Docker image.

The build context itself is a temporary directory which contains the binaries
and packages for the current platform, which you can `COPY` into your image.

A corollary of that is that **the context does not contain the source files**.
If you need to add some other file that is in your source directory, you'll need
to add it to the `extra_files` property, so it'll get copied into the context.

All that being said, your Docker build context will usually look like this:

```sh
temp-context-dir
├── Dockerfile
├── myprogram     # the binary
├── myprogram.rpm # linux package
├── myprogram.apk # linux package
└── myprogram.deb # linux package
```

`myprogram` would actually be your binary name, and the Linux package names
would follow their respective configuration's `name_templates`.

## Customization

Of course, you can customize a lot of things:

```yaml title=".goreleaser.yaml"
dockers:
  # You can have multiple Docker images.
  - #
    # ID of the image, needed if you want to filter by it later on (e.g. on custom publishers).
    id: myimg

    # GOOS of the built binaries/packages that should be used.
    # Default: 'linux'.
    goos: linux

    # GOARCH of the built binaries/packages that should be used.
    # Default: 'amd64'.
    goarch: amd64

    # GOARM of the built binaries/packages that should be used.
    # Default: '6'.
    goarm: ""

    # GOAMD64 of the built binaries/packages that should be used.
    # Default: 'v1'.
    goamd64: "v2"

    # IDs to filter the binaries/packages.
    #
    # Make sure to only include the IDs of binaries you want to `COPY` in your
    # Dockerfile.
    #
    # If you include IDs that don't exist or are not available for the current
    # architecture being built, the build of the image will be skipped.
    ids:
      - mybuild
      - mynfpm

    # Templates of the Docker image names.
    #
    # Templates: allowed.
    image_templates:
      - "myuser/myimage:latest"
      - "myuser/myimage:{{ .Tag }}"
      - "myuser/myimage:{{ .Tag }}-{{ .Env.FOOBAR }}"
      - "myuser/myimage:v{{ .Major }}"
      - "gcr.io/myuser/myimage:latest"

    # Skips the docker build.
    # Could be useful if you want to skip building the windows docker image on
    # linux, for example.
    #
    # This option is only available on GoReleaser Pro.
    # Templates: allowed.
    skip_build: false

    # Skips the docker push.
    # Could be useful if you also do draft releases.
    #
    # If set to auto, the release will not be pushed to the Docker repository
    #  in case there is an indicator of a prerelease in the tag, e.g. v1.0.0-rc1.
    #
    # Templates: allowed.
    skip_push: false

    # Path to the Dockerfile (from the project root).
    #
    # Default: 'Dockerfile'.
    # Templates: allowed.
    dockerfile: "{{ .Env.DOCKERFILE }}"

    # Use this instead of `dockerfile` if the contents of your Dockerfile are
    # supposed to go through the template engine as well.
    #
    # `dockerfile` is ignored when this is set.
    #
    # This feature is only available in GoReleaser Pro.
    # Templates: allowed.
    templated_dockerfile: "{{.Env.DOCKERFILE }}"

    # Set the "backend" for the Docker pipe.
    #
    # Valid options are: docker, buildx, podman.
    #
    # Podman is a GoReleaser Pro feature and is only available on Linux.
    #
    # Default: 'docker'.
    use: docker

    # Docker build flags.
    #
    # Templates: allowed.
    build_flag_templates:
      - "--pull"
      - "--label=org.opencontainers.image.created={{.Date}}"
      - "--label=org.opencontainers.image.title={{.ProjectName}}"
      - "--label=org.opencontainers.image.revision={{.FullCommit}}"
      - "--label=org.opencontainers.image.version={{.Version}}"
      - "--build-arg=FOO={{.Env.Bar}}"
      - "--platform=linux/arm64"

    # Extra flags to be passed down to the push command.
    push_flags:
      - --tls-verify=false

    # If your Dockerfile copies files other than binaries and packages,
    # you should list them here as well.
    # Note that GoReleaser will create the same structure inside a temporary
    # directory, so if you add `foo/bar.json` here, on your Dockerfile you can
    # `COPY foo/bar.json /whatever.json`.
    # Also note that the paths here are relative to the directory in which
    # GoReleaser is being run (usually the repository root directory).
    # This field does not support wildcards, you can add an entire directory here
    # and use wildcards when you `COPY`/`ADD` in your Dockerfile.
    extra_files:
      - config.yml

    # Additional templated extra files to add to the Docker image.
    # Those files will have their contents pass through the template engine,
    # and its results will be added to the build context the same way as the
    # extra_files field above.
    #
    # This feature is only available in GoReleaser Pro.
    # Templates: allowed.
    templated_extra_files:
      - src: LICENSE.tpl
        dst: LICENSE.txt
        mode: 0644
```

!!! warning

    Note that you will have to manually login into the Docker registries you
    want to push to — GoReleaser does not login by itself.

<!-- md:templates -->

!!! tip

    You can also create multi-platform images using the [docker_manifests](docker_manifest.md) config.

These settings should allow you to generate multiple Docker images,
for example, using multiple `FROM` statements,
as well as generate one image for each binary in your project or one image with multiple binaries, as well as
install the generated packages instead of copying the binary and configs manually.

## Generic Image Names

Some users might want to keep their image name as generic as possible.
That can be accomplished simply by adding template language in the definition:

```yaml title=".goreleaser.yaml"
project_name: foo
dockers:
  - image_templates:
      - "myuser/{{.ProjectName}}"
```

This will build and publish the following images:

- `myuser/foo`

<!-- md:templates -->

## Keeping docker images updated for current major

Some users might want to push docker tags `:v1`, `:v1.6`,
`:v1.6.4` and `:latest` when `v1.6.4` (for example) is built. That can be
accomplished by using multiple `image_templates`:

```yaml title=".goreleaser.yaml"
dockers:
  - image_templates:
      - "myuser/myimage:{{ .Tag }}"
      - "myuser/myimage:v{{ .Major }}"
      - "myuser/myimage:v{{ .Major }}.{{ .Minor }}"
      - "myuser/myimage:latest"
```

This will build and publish the following images:

- `myuser/myimage:v1.6.4`
- `myuser/myimage:v1`
- `myuser/myimage:v1.6`
- `myuser/myimage:latest`

With these settings you can hopefully push several Docker images
with multiple tags.

<!-- md:templates -->

## Publishing to multiple docker registries

Some users might want to push images to multiple docker registries. That can be
accomplished by using multiple `image_templates`:

```yaml title=".goreleaser.yaml"
dockers:
  - image_templates:
      - "docker.io/myuser/myimage:{{ .Tag }}"
      - "docker.io/myuser/myimage:latest"
      - "gcr.io/myuser/myimage:{{ .Tag }}"
      - "gcr.io/myuser/myimage:latest"
```

This will build and publish the following images to `docker.io` and `gcr.io`:

- `myuser/myimage:v1.6.4`
- `myuser/myimage:latest`
- `gcr.io/myuser/myimage:v1.6.4`
- `gcr.io/myuser/myimage:latest`

## Applying Docker build flags

Build flags can be applied using `build_flag_templates`.
The flags must be valid Docker build flags.

```yaml title=".goreleaser.yaml"
dockers:
  - image_templates:
      - "myuser/myimage"
    build_flag_templates:
      - "--pull"
      - "--label=org.opencontainers.image.created={{.Date}}"
      - "--label=org.opencontainers.image.title={{.ProjectName}}"
      - "--label=org.opencontainers.image.revision={{.FullCommit}}"
      - "--label=org.opencontainers.image.version={{.Version}}"
```

This will execute the following command:

```bash
docker build -t myuser/myimage . \
  --pull \
  --label=org.opencontainers.image.created=2020-01-19T15:58:07Z \
  --label=org.opencontainers.image.title=mybinary \
  --label=org.opencontainers.image.revision=da39a3ee5e6b4b0d3255bfef95601890afd80709 \
  --label=org.opencontainers.image.version=1.6.4
```

<!-- md:templates -->

## Use a specific builder with Docker buildx

If `buildx` is enabled, the `default` context builder will be used when building
the image. This builder is always available and backed by BuildKit in the
Docker engine. If you want to use a different builder, you can specify it using
the `build_flag_templates` field:

```yaml title=".goreleaser.yaml"
dockers:
  - image_templates:
      - "myuser/myimage"
    use: buildx
    build_flag_templates:
      - "--builder=mybuilder"
```

!!! tip

    Learn more about the [buildx builder instances](https://docs.docker.com/buildx/working-with-buildx/#work-with-builder-instances).

## Using Podman

<!-- md:pro -->

You can use [`podman`](https://podman.io) instead of `docker` by setting `use` to `podman` on your config:

```yaml title=".goreleaser.yaml"
dockers:
  - image_templates:
      - "myuser/myimage"
    use: podman
```

Note that GoReleaser will not install Podman for you, nor change any of its
configuration.

If you want to use it rootless, make sure to follow
[this guide](https://github.com/containers/podman/blob/main/docs/tutorials/rootless_tutorial.md).
