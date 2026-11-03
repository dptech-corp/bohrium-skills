# Platform snapshot

Pinned CPU / GPU template shortcuts and SKU discovery.

> Snapshot of platform-side values that may drift over time. Resource sizes are
> observed `cpuCount` / `memoryMB` from a create response and may change. For a
> live view, use the commands noted below.

## Image-override CPU profiles

Verified 2026-09-03; review by 2026-10-03. All six current templates are
public and CPU-only. The profile key is the stable CLI surface; its backing
template may change. The create response remains authoritative.

| `--cpu` profile | current public template | sku | CPU | memory |
| --- | --- | --- | ---: | ---: |
| `1c2g` (default) | `dynamic-image-replace` | `c1_m2_cpu` | 1 | 2 GiB |
| `2c4g` | `d-image-replace-c2-m4` | `c2_m4_cpu` | 2 | 4 GiB |
| `2c8g` | `d-image-replace-c2-m8` | `c2_m8_cpu` | 2 | 8 GiB |
| `4c8g` | `d-image-replace-c4-m8` | `c4_m8_cpu` | 4 | 8 GiB |
| `8c32g` | `d-image-replace-c8-m32` | `c8_m32_cpu` | 8 | 32 GiB |
| `16c32g` | `d-image-replace-c16-m32` | `c16_m32_cpu` | 16 | 32 GiB |

`bohr sandbox create --image <image-path>` without `--cpu` selects `1c2g`.
Use `bohr sandbox template list -o json` and `bohr sandbox machine list -o
json` for live values; report `templateID`, `cpuCount`, and `memoryMB` from the
create response when exact resources matter.

## GPU shortcuts

`bohr sandbox create --gpu` is a shortcut that selects a platform-managed GPU
template. GPU keys are platform-managed and may change over time — consult
`bohr sandbox template list` for the current list.

| `--gpu` key | Selects |
| --- | --- |
| `--gpu` (bare) or `--gpu 4090` | the default GPU template (NVIDIA 4090 class) |
| `--gpu 5090` | NVIDIA 5090 class |
| `--gpu l20` | NVIDIA L20 class |

`--gpu` is mutually exclusive with passing a positional template name or
`--cpu`.

**Out of stock.** There is no availability field to check before creating, so a
GPU model with no stock only shows up as a failed `create`. On an explicit
stock rejection, the CLI lists the other
`--gpu` keys in `error.hint.action` of the JSON envelope (and on stderr for
humans); retry with one of them. Account-level failures — authentication,
permission, rate limit, quota, insufficient balance — carry no such hint, since a
different model cannot fix them. Ordinary validation failures also carry no
alternative hint. A timeout or server failure may leave a sandbox already
created: reconcile the original request before attempting another create. This only covers the three `--gpu` shortcuts: a
GPU template passed via `--template` gets no alternative list, because the CLI
cannot tell which templates are GPU-backed. Do not assume `machine list` tells
you stock — it does not.

```bash
bohr sandbox create --gpu -o json
bohr sandbox exec <sandbox_id> nvidia-smi
```

## SKU catalog

SKU availability is dynamic. Use the live list instead of copying values from
documentation:

```bash
bohr sandbox machine list                # list default SKU category
bohr sandbox machine list -c gpu         # list GPU SKUs
bohr sandbox machine list -o json        # machine-readable SKU data
bohr sandbox machine list --sort price   # sort by a column (sku_name|cpu|memory|gpu|price)
bohr sandbox machine list --sort price -d # -d/--descending reverses the sort order
```

Pass the returned `sku_name` value to `bohr sandbox template create --sku-name`.
Quote SKU names that contain spaces or shell metacharacters.

## Example image

- Find registry image references with `bohr image ls` (paste an image path into
  `bohr sandbox template create --image`).
