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

## Choose the task

| Goal | Route |
|---|---|
| Reuse an existing image | List/search → get → keep the canonical registry address |
| Build with local source files | Local build → returned `pushCommand` → validated push receipt |
| Build Dockerfile text remotely | Remote build → terminal image status → inspect final image |
| Save a sandbox rootfs | Storage gate below → commit → terminal status; mounted data is separate |

## Workflow: build locally, publish, then use the image

1. Confirm Docker context/daemon, Dockerfile/build context, unique short name, project and write intent. Local mode reserves a remote record/quota but does not push.

```bash
bohr image build --mode local --dockerfile ./Dockerfile --context .   --name research-env:v1 --project-id <project_id> -o json
```

2. Preserve the local-build result's `data.localImage`, `imageUrl`, `id` and `pushCommand`. With `uploaded=false`, the image is only built locally. Run the actual returned `pushCommand` in the same Docker daemon, Bohrium identity/profile and project; do not reconstruct a registry path.
3. A successful push validates its final `data.id`, `imageUrl`, `status=1`, `digest` and `sizeBytes`. Use that final `imageUrl` with Job `--image_address` or Node `--image_path`; an image record ID or local tag is not the registry address.
4. The local prepare record is not a remote build task: do not poll it with `status`/`build-log`. Remote build/commit uses the separate `0/1 → 2/3/4` lifecycle described below; its success status `2` must not be applied to a local push receipt.
5. Complete with the published address and digest plus the selected consumer's requirements. If build/push is interrupted or prepare outcome is uncertain, reconcile the same name/record using the reference before reserving another one. A built local image alone is unfinished publication.

Choose `build --mode local` for a Docker build with local context; run its
returned push command separately. Default `build --mode remote` sends Dockerfile
text only to the remote builder. `commit` snapshots a running sandbox container's
root filesystem on the server; do not assume mounted volumes are included.

Discover an existing image before creating one:

```bash
bohr image list --page 1 --size 50 -o json
bohr image search "lammps cuda" -o json
bohr image search "DeePMD" --by-name -o json
bohr image get <image_id> -o json
bohr image packages <image_id> -o json
```

`list` uses `--page` (default `1`) and `--size` (default `100`, accepted `1-512`). The service can cap a page at 100; follow the returned `data.pagination.page_size` and `has_more` rather than assuming the requested size was honored. `--limit` is a hidden compatibility spelling with the same default.

Search defaults to an AND intersection of installed software capabilities. Use repeatable `--capability`/`--framework` filters for explicit software requirements, or `--by-name` with a positional keyword for repository/tag name search; name mode cannot combine with those filters. Search pagination uses `--page` and `--page-size`.

`packages` reads extracted software facts grouped in the returned `data.packages` map. Supply either a catalog image ID or `--image <full_registry_reference>`, never both. Scanner facts are evidence for selection, not proof that the workload has run successfully.

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
