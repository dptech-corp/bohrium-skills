---
name: bohrium-image
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Discover, inspect, pull, push, build, commit, monitor, or delete Bohrium images with
  `bohr image`. Use for container images, 镜像管理, public/private image lookup,
  Docker pulls and pushes, Dockerfile builds, sandbox snapshots, build logs, image IDs,
  and image lifecycle states.
---

# Bohrium Image

Choose `build --mode local` for a Docker build with local context; run its
returned push command separately. Default `build --mode remote` sends Dockerfile
text only to the remote builder. `commit` snapshots a running sandbox container's
root filesystem on the server; do not assume mounted volumes are included.

Discover an existing image before creating one:

```bash
bohr image list --limit 50 -o json
bohr image search "DeePMD" -o json
bohr image get <image_id> -o json
```

`--limit` accepts `1-512` and defaults to `512`; use a smaller value when only
a bounded catalog sample is needed.

Pull copies a Bohrium image into the local Docker daemon. Confirm Docker is
installed/running and the address belongs to the Bohrium registry:

```bash
bohr image pull <bohrium_image_address> --timeout 30m -o json
```

Build with local context and then run the returned `pushCommand` with the same
Docker context/daemon, Bohrium profile and identity, project and short name:

```bash
bohr image build --mode local --dockerfile ./Dockerfile --context . \
  --name my-image:v1 --project-id 1234 -o json
```

This uses prepare to reserve a remote image record and quota; it is not offline.
Local build does not submit remote compute or require `--yes`, and does not push.
After local success, run `pushCommand`; do not poll `image status` or `build-log`
with the prepare record ID. Use a unique name and do not run same-name operations
concurrently. A known build failure attempts abort. If prepare does not return a
reliable result, inspect the state first: even an error can leave a record, so
do not assume it is absent or was cleaned up.

Alternatively, push an existing Docker daemon image with your AccessKey. The command invokes
the Docker CLI on `PATH`; it does not build an image or upload a Dockerfile or
context. Prepare the image yourself with `docker build`, `docker load`, or
`docker pull` first. The second argument is a
build/commit-style short image name, and project ID comes from `--project-id` or
`BOHR_PROJECT_ID`:

```bash
docker build -t my-image:v1 .
bohr image push my-image:v1 my-image:v1 --project-id 1234 --timeout 30m -o json
```

Use only the prepare response's canonical `imageUrl` for the registry target;
never construct a project/user path. The lifecycle and failure recovery are in
[`references/commands.md`](references/commands.md).

## Rootfs gate before a sandbox snapshot commit

Before every `bohr image commit`, run
`bohr sandbox storage inspect <sandbox_id> -o json` immediately before submit.
Do not rely on an 80% threshold after commit starts: commit apply may consume
temporary rootfs faster than reactive expansion becomes Ready. Use the
failure-data formula in [`references/commands.md`](references/commands.md),
resize first when needed, and continue only after `status=ready`. If the
expected uncompressed commit delta is unknown, do not treat it as zero or claim
the capacity is safe. Storage resize is grow-only and cannot shrink a rootfs.
This Agent preflight is not automatic CLI behavior and
does not apply to `bohr image build --dockerfile`.

Remote build and commit are resource operations charged from balance under the
Sandbox pricing catalog. Inspect price shape and authorize spending first:

```bash
bohr billing pricing --resource sandbox -o json
bohr image build --mode remote --dockerfile ./Dockerfile --name project/image:v1 \
  --project-id <project_id> --wait --yes -o json
bohr image commit --sandbox-id <sandbox_id> --name project/snapshot:v1 \
  --project-id <project_id> --wait --yes -o json
```

Use immutable names rather than a mutable `latest` tag. Remote build sends only the
selected Dockerfile text for a remote build, without local build context;
commit snapshots the selected running sandbox container's root filesystem on the server.
Neither uses a local Docker image. Confirm
the exact local file, sandbox, owning project, name, and spending authorization.

For remote builds and sandbox commits without `--wait`, capture the returned
image ID and follow it with:

```bash
bohr image build-log <image_id> --follow
bohr image status <image_id> -o json
bohr image dockerfile <image_id>
```

Creating (`0`) and pending (`1`) are non-terminal. Image status `2` is success,
status `3` is failed, and status `4` is canceled. `--wait` stops on `2/3/4`;
bound it with `--wait-timeout`. A local timeout does not prove the remote build
stopped.

Delete is destructive:

```bash
bohr image delete <image_id> --yes -o json
```

The current CLI has no share/unshare or metadata-update command. Read
[`references/commands.md`](references/commands.md) for the full command set,
flags, local/remote effects, billing, and lifecycle completion.

## Resource payment boundary

Resource actions marked balance-only keep their existing balance billing. The
global `--payment-source` option does not enable photon billing for these
resources; explicitly selecting `photon` on a balance-only action is rejected
before execution. Use the existing resource pricing/authorization workflow.
