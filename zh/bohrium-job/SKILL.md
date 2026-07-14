---
name: bohrium-job
description: "Manage Bohrium compute jobs via bohr CLI or open.bohrium.com API. Use when: user asks about submitting/listing/killing/deleting compute jobs on Bohrium, checking job logs, or monitoring job status. NOT for: node management, image management, or project management."
---

# SKILL: Bohrium 任务 (Job) 管理

## 概述

管理 Bohrium 平台的计算任务。**优先使用 `bohr` CLI**，仅在 CLI 不支持的高级操作时回退到 API。

## 认证配置

BOHR_ACCESS_KEY 和 PROJECT_ID 从 OpenClaw 配置文件 `~/.openclaw/openclaw.json` 中读取：

```json
"bohrium-job": {
  "enabled": true,
  "apiKey": "YOUR_BOHR_ACCESS_KEY",
  "env": {
    "BOHR_ACCESS_KEY": "YOUR_BOHR_ACCESS_KEY",
    "PROJECT_ID": "YOUR_PROJECT_ID"
  }
}
```

OpenClaw 会自动将 `env` 中的变量注入到运行环境。Skill 只要求配置 `BOHR_ACCESS_KEY`；兼容旧 CLI 所需的映射由辅助脚本内部处理。

直接调用 `bohr` CLI 时，先把 `BOHR_ACCESS_KEY` 映射给旧 CLI 读取的 `ACCESS_KEY`：

```bash
export ACCESS_KEY="$BOHR_ACCESS_KEY"
```

## 前置条件：安装 bohr CLI

```bash
# macOS
/bin/bash -c "$(curl -fsSL https://dp-public.oss-cn-beijing.aliyuncs.com/bohrctl/1.0.0/install_bohr_mac_curl.sh)"

# Linux
/bin/bash -c "$(curl -fsSL https://dp-public.oss-cn-beijing.aliyuncs.com/bohrctl/1.0.0/install_bohr_linux_curl.sh)"

# 安装后确认
source ~/.bashrc  # 或 source ~/.zshrc
export PATH="$HOME/.bohrium:$PATH"
export ACCESS_KEY="$BOHR_ACCESS_KEY"
bohr version
```

> 安装脚本会自动配置 `OPENAPI_HOST` 和 `TIEFBLUE_HOST` 环境变量。若未生效，手动设置：
> ```bash
> export OPENAPI_HOST=https://open.bohrium.com
> export TIEFBLUE_HOST=https://tiefblue.dp.tech
> ```

---

## 提交任务

### 方式 1：命令行参数（推荐）

```bash
bohr job submit \
  -m "registry.dp.tech/dptech/deepmd-kit:3.1.1" \
  -t "c4_m15_1 * NVIDIA T4" \
  -c "python train.py" \
  -p ./input_dir/ \
  --project_id YOUR_PROJECT_ID \
  -n "my-job-name"
```

**参数说明：**

| 参数 | 缩写 | 必填 | 说明 |
|------|------|------|------|
| `--image_address` | `-m` | 是 | 镜像地址（必须用完整地址，如 `registry.dp.tech/dptech/xxx:tag`） |
| `--machine_type` | `-t` | 是 | 机器规格 |
| `--command` | `-c` | 是 | 执行命令 |
| `--input_directory` | `-p` | 否 | 输入文件目录（默认 `./`） |
| `--project_id` | | 是 | 项目 ID |
| `--job_name` | `-n` | 否 | 任务名称 |
| `--log_file` | `-l` | 否 | 日志文件路径 |
| `--result_path` | `-r` | 否 | 完成后自动下载结果到指定目录（仅 `/data`、`/personal`、`/share`） |
| `--job_group_id` | `-g` | 否 | 任务组 ID（需先通过 `bohr job_group create` 创建） |
| `--max_run_time` | | 否 | 最大运行时间（分钟），超时自动终止并回收文件 |
| `--max_reschedule_times` | | 否 | 异常中断后自动重试次数 |
| `--nnode` | | 否 | 并行计算节点数（默认 1） |

### 方式 2：配置文件（适合复杂场景）

```bash
bohr job submit -i job.json -p ./input_dir/
```

**job.json 完整示例：**

```json
{
  "job_name": "my-training-job",
  "command": "python train.py --epochs 10",
  "log_file": "train.log",
  "backward_files": ["model.pt", "results/"],
  "project_id": YOUR_PROJECT_ID,
  "machine_type": "c4_m15_1 * NVIDIA T4",
  "image_address": "registry.dp.tech/dptech/deepmd-kit:3.1.1",
  "job_type": "container",
  "disk_size": 50,
  "dataset_path": ["/bohr/my-dataset/v1"],
  "result_path": "/personal",
  "max_reschedule_times": 2,
  "max_run_time": 120,
  "nnode": 1
}
```

**job.json 字段说明：**

| 字段 | 说明 | 示例 |
|------|------|------|
| `job_name` | 任务名称 | `"DeePMD-kit test"` |
| `command` | 执行命令，用**相对路径** | `"cd se_e2_a && dp train input.json"` |
| `log_file` | 运行期间可实时查看的日志文件 | `"train.log"` |
| `backward_files` | 需要下载的结果文件，为空则保留全部 | `["model.pt", "results/"]` |
| `project_id` | 项目 ID | `YOUR_PROJECT_ID` |
| `machine_type` | 机器规格 | `"c4_m15_1 * NVIDIA T4"` |
| `image_address` | 完整镜像地址（非镜像名） | `"registry.dp.tech/dptech/deepmd-kit:2.1.5-cuda11.6"` |
| `job_type` | 必须为 `"container"` | `"container"` |
| `dataset_path` | 挂载的数据集路径数组 | `["/bohr/my-dataset/v1"]` |
| `result_path` | 自动回收结果到数据盘 | `"/personal"` |
| `max_run_time` | 最大运行时间（分钟） | `120` |
| `max_reschedule_times` | 异常中断重试次数 | `2` |
| `disk_size` | 磁盘大小(GB) | `50` |

### 关键注意事项

| 项目 | 说明 |
|------|------|
| **工作目录** | Bohrium 自动 `cd` 到解压后的输入目录，cmd 中用**相对路径** |
| **不要** `cd /root/input` | 实际路径是 `/home/input_lbg-{userId}-{jobId}/`，不可预测 |
| **image_address 格式** | 必须用完整地址 `registry.dp.tech/dptech/xxx:tag`，不能只写镜像名 |
| **machine_type 格式** | CPU: `c2_m4_cpu`；GPU: `c4_m15_1 * NVIDIA T4`、`c8_m32_1 * NVIDIA V100` |
| **WAF 拦截** | 若 command 含 shell 关键字被阿里云 WAF 拦截(405)，将命令写入脚本，cmd 改为 `bash run.sh` |
| **-p 参数大文件** | 提交后无输出提示时，检查 `-p` 目录下是否有大的隐藏文件导致压缩慢 |
| **自动重试** | 设置 `max_reschedule_times` 可在异常中断后自动重试（全量重跑） |
| **job_type** | 必须设为 `"container"`，自 2023 年起不再支持虚拟机任务 |

---

## 查看任务

```bash
bohr job list -n 10              # 最近 10 个
bohr job list -n 5 --json        # JSON 格式输出
bohr job list -r                 # 只看运行中
bohr job list -f                 # 只看失败的
bohr job list -i                 # 只看已完成
bohr job list -p                 # 只看等待中
bohr job list -j YOUR_JOB_GROUP_ID        # 查看指定 job group 下的任务
```

### 查看任务详情

```bash
bohr job describe -j YOUR_JOB_ID --json    # JSON 格式详情
bohr job describe -j YOUR_JOB_ID -l        # 完整详细信息
```

### 查看/下载日志

```bash
bohr job log -j YOUR_JOB_ID                # 查看日志
bohr job log -j YOUR_JOB_ID -o ./logs/     # 下载日志到本地目录
```

### 下载任务结果

```bash
bohr job download -j YOUR_JOB_ID -o ./results/
bohr job_group download -j YOUR_JOB_GROUP_ID -o ./results/   # 下载整组结果
```

---

## 管理任务

```bash
bohr job terminate YOUR_JOB_ID             # 终止任务（保留结果文件，状态变为 completed）
bohr job kill YOUR_JOB_ID                  # 强制停止（不保留结果文件，不删除任务记录）
bohr job delete YOUR_JOB_ID                # 删除任务（状态变为 failed，结果文件删除，记录消失）
bohr job terminate YOUR_JOB_ID YOUR_JOB_ID_2    # 批量终止
bohr job delete YOUR_JOB_ID YOUR_JOB_ID_2       # 批量删除
```

**terminate vs kill vs delete：**

| 操作 | 结果文件 | 任务记录 | 状态 |
|------|---------|---------|------|
| `terminate` | 保留 | 保留 | → completed |
| `kill` | 不保留 | 保留 | → failed |
| `delete` | 删除 | 删除 | 从列表消失 |

---

## 任务组管理

```bash
bohr job_group list -n 10 --json                    # 列出任务组
bohr job_group list -s 2026-01-01 -e 2026-03-14     # 按日期范围过滤
bohr job_group create -n "experiment-v1" -p YOUR_PROJECT_ID      # 创建任务组
bohr job_group terminate YOUR_JOB_GROUP_ID                    # 终止任务组
bohr job_group delete YOUR_JOB_GROUP_ID                       # 删除任务组
bohr job_group download -j YOUR_JOB_GROUP_ID -o ./results/    # 下载任务组结果
```

> **注意**：`bohr job_group create` 创建的 `job_group_id` 用于 `bohr job submit -g <id>` 将多个任务归入同一组。此 ID 与 Web 界面上的任务组 ID 不同，仅适用于 CLI 提交。

---

## 科学计算软件示例

### DeePMD-kit 训练任务

```json
{
  "job_name": "DeePMD-kit test",
  "command": "cd se_e2_a && dp train input.json > tmp_log 2>&1 && dp freeze -o graph.pb",
  "log_file": "se_e2_a/tmp_log",
  "backward_files": ["se_e2_a/lcurve.out", "se_e2_a/graph.pb"],
  "project_id": YOUR_PROJECT_ID,
  "machine_type": "c4_m15_1 * NVIDIA T4",
  "job_type": "container",
  "image_address": "registry.dp.tech/dptech/deepmd-kit:2.1.5-cuda11.6"
}
```

### LAMMPS 分子动力学

```json
{
  "job_name": "lammps_tutorial",
  "command": "mpirun -n 32 lmp_mpi -i in.shear > log",
  "log_file": "log",
  "backward_files": [],
  "project_id": YOUR_PROJECT_ID,
  "machine_type": "c32_m64_cpu",
  "job_type": "container",
  "image_address": "registry.dp.tech/dptech/lammps:29Sep2021"
}
```

### ABACUS 第一性原理

```json
{
  "job_name": "ABACUS test",
  "command": "OMP_NUM_THREADS=1 mpirun -np 8 abacus > log",
  "log_file": "log",
  "backward_files": [],
  "project_id": YOUR_PROJECT_ID,
  "machine_type": "c16_m32_cpu",
  "job_type": "container",
  "image_address": "registry.dp.tech/dptech/abacus:3.0.0"
}
```

### GROMACS 分子模拟

```json
{
  "job_name": "bohrium-gmx-example",
  "command": "bash rungmx.sh > log",
  "log_file": "log",
  "backward_files": [],
  "project_id": YOUR_PROJECT_ID,
  "machine_type": "c16_m62_1 * NVIDIA T4",
  "job_type": "container",
  "image_address": "registry.dp.tech/dptech/gromacs:2022.2"
}
```

### CP2K 量子化学

```json
{
  "job_name": "CP2K_Si_opt",
  "command": "source /cp2k-7.1/tools/toolchain/install/setup && mpirun -n 16 --allow-run-as-root --oversubscribe cp2k.popt -i input.inp -o output.log",
  "log_file": "output.log",
  "backward_files": ["output.log"],
  "project_id": YOUR_PROJECT_ID,
  "machine_type": "c16_m32_cpu",
  "job_type": "container",
  "image_address": "registry.dp.tech/dptech/cp2k:7.1"
}
```

---

## 容器镜像与虚拟机镜像映射

| 软件 | 虚拟机镜像 | 容器镜像地址 |
|------|-----------|-------------|
| DeePMD-kit | LBG_DeePMD-kit_2.1.4_v1 | `registry.dp.tech/dptech/deepmd-kit:2.1.5-cuda11.6` |
| DPGEN | LBG_DP-GEN_0.10.6_v3 | `registry.dp.tech/dptech/dpgen:0.10.6` |
| LAMMPS | LBG_LAMMPS_stable_23Jun2022_v1 | `registry.dp.tech/dptech/lammps:29Sep2021` |
| GROMACS | gromacs-dp:2020.2 | `registry.dp.tech/dptech/gromacs:2022.2` |
| Quantum-Espresso | LBG_Quantum-Espresso_7.1 | `registry.dp.tech/dptech/quantum-espresso:7.1` |
| CP2K | - | `registry.dp.tech/dptech/cp2k:7.1` |
| ABACUS | - | `registry.dp.tech/dptech/abacus:3.0.0` |
| 基础镜像 (CPU) | LBG_Common_v1/v2 | `registry.dp.tech/dptech/ubuntu:20.04-py3.10` |
| 基础镜像 (GPU) | LBG_base_image_ubun20.04 | `registry.dp.tech/dptech/ubuntu:20.04-py3.10-cuda11.6` |
| Intel OneAPI | LBG_oneapi_2021_v1 | `registry.dp.tech/dptech/ubuntu:20.04-py3.10-intel2022-cuda11.6` |

---

## API 补充（CLI 不支持的操作）

以下操作 bohr CLI 不支持，需通过 API 完成：

```python
import os, requests

AK = os.environ.get("BOHR_ACCESS_KEY", "")
BASE = "https://open.bohrium.com/openapi/v1"
HEADERS = {"Authorization": f"Bearer {AK}"}

# 按状态过滤任务列表 (0=等待, 1=运行中, 2=完成, 3=调度中, -1=失败)
r = requests.get(f"{BASE}/job/list", headers=HEADERS,
    params={"page": 1, "pageSize": 10, "status": 1})

# 查看任务配置（含文件访问 token）
r = requests.get(f"{BASE}/job/view/conf/{job_id}", headers=HEADERS)
# 返回: {state, baseDir, tempDir, token, host, expires}
# token + host 可用于 tiefblue 访问任务文件

# 查看快照
r = requests.get(f"{BASE}/job/{job_id}/snapshot", headers=HEADERS)

# 修改任务名
requests.post(f"{BASE}/job/{job_id}/modify",
    headers={**HEADERS, "Content-Type": "application/json"},
    json={"name": "new-name"})

# 修改任务组名
requests.post(f"{BASE}/job_group/{job_group_id}/modify",
    headers={**HEADERS, "Content-Type": "application/json"},
    json={"name": "new-group-name"})
```

### 直接用 API 提交任务（不经 bohr CLI）

`bohr job submit` 底层不是单个接口，而是**按顺序组合调用多个接口**：① 建任务组（可选）→ ② 建 job 拿上传凭证 → ③ 把输入 zip 上传到 tiefblue → ④ 正式提交调度。

```python
import os, json, base64, requests

AK = os.environ["BOHR_ACCESS_KEY"]
PROJECT_ID = int(os.environ["PROJECT_ID"])
BASE = "https://open.bohrium.com/openapi/v4"
H = {"Authorization": f"Bearer {AK}", "Content-Type": "application/json"}

# ① 创建任务组（已有组可跳过，直接用其 groupId 作为 BohrGroupId）
group_id = requests.post(f"{BASE}/job_group/add", headers=H,
    json={"projectId": PROJECT_ID, "name": "my-batch"}).json()["data"]["groupId"]

# ② 创建 job，拿上传凭证
d = requests.post(f"{BASE}/job/create", headers=H,
    json={"name": "my-job", "BohrGroupId": group_id, "projectId": PROJECT_ID}).json()["data"]
job_id, store_host, store_path, token = d["jobId"], d["storeHost"], d["storePath"], d["token"]

# ③ 把 input.zip 上传到 tiefblue（<~50MB 用二进制直传；更大走 /api/upload/multipart/*）
object_key = store_path + "input.zip"
xsp = base64.b64encode(json.dumps({
    "path": object_key,
    "option": {"contentDisposition": 'attachment; filename="input.zip"'},
}).encode()).decode()
with open("input.zip", "rb") as f:
    requests.post(f"{store_host}/api/upload/binary",
        headers={"Authorization": f"Bearer {token}", "X-Storage-Param": xsp},
        data=f.read())

# ④ 正式提交调度（jobType 固定 "container"，ossPath 为数组）
resp = requests.post(f"{BASE}/job/add", headers=H, json={
    "jobId": job_id, "jobType": "container", "projectId": PROJECT_ID,
    "jobGroupId": group_id, "jobName": "my-job",
    "scassType": "c2_m4_cpu",
    "imageName": "registry.dp.tech/dptech/ubuntu:20.04-py3.10",
    "cmd": "bash run.sh", "platform": "ali",
    "ossPath": [object_key], "inputFileType": 3, "inputFileMethod": 1, "nnode": 1,
}).json()
print(resp)  # {"code": 0, "data": {"jobId": ..., "bohrJobId": ..., "jobGroupId": ...}}
```

> **验证状态（2026-07-16，正式 AK 于 open.bohrium.com）**：①②③④ 四步已端到端验证通过——④ `job/add` 返回 `{code:0, data:{jobId, bohrJobId, jobGroupId}}`，任务进入调度。用 `v4` 版本、`ossPath` 传**数组**、`jobType` 固定 `container`。返回的 `jobId` 可用于 `POST /openapi/v4/job/kill/{jobId}` 终止。

---

## 任务状态码

| status | 含义 | bohr CLI 显示 |
|--------|------|---------------|
| 0 | 等待中 | Pending |
| 1 | 运行中 | Running |
| 2 | 已完成 | Finished |
| 3 | 调度中 | Scheduling |
| -1 | 失败 | Failed |

## 常见机器规格

| machine_type | 说明 |
|--------------|------|
| `c2_m4_cpu` | 2 核 4G 内存 CPU |
| `c4_m8_cpu` | 4 核 8G 内存 CPU |
| `c8_m32_cpu` | 8 核 32G 内存 CPU |
| `c16_m32_cpu` | 16 核 32G 内存 CPU |
| `c32_m64_cpu` | 32 核 64G 内存 CPU |
| `c4_m15_1 * NVIDIA T4` | 4 核 15G + 1×T4 GPU |
| `c16_m62_1 * NVIDIA T4` | 16 核 62G + 1×T4 GPU |
| `c8_m32_1 * NVIDIA V100` | 8 核 32G + 1×V100 GPU |
| `c32_m128_4 * NVIDIA V100` | 32 核 128G + 4×V100 GPU |

## 常见错误与排查

| 问题 | 原因 | 解决 |
|------|------|------|
| `cd /root/input: No such file` | cmd 中用了绝对路径 | Bohrium 工作目录不可预测，用**相对路径** |
| `unsupported protocol scheme ""` | bohr CLI 缺少环境变量 | `export OPENAPI_HOST=https://open.bohrium.com && export TIEFBLUE_HOST=https://tiefblue.dp.tech` |
| `(200, '/account/login', None)` | 旧版 lbg (pip) 不支持 access_key | 用新版 Go CLI（`~/.bohrium/bohr`） |
| `AccessKey Invalid` | 直接调用 `bohr` 时未设置旧变量名 | `export ACCESS_KEY="$BOHR_ACCESS_KEY"` 后重试 |
| WAF 405 拦截 | cmd 含 shell 关键字被阿里云 WAF 拦截 | 将命令写入脚本文件，cmd 改为 `bash run.sh` |
| `Permission error` | Job 不属于当前用户 | 确认 BOHR_ACCESS_KEY 对应的用户 |
| `jobId` vs `jobGroupId` 搞混 | 两个是不同概念 | CLI 中 `kill/terminate/delete` 均使用 `jobId` |
| 提交后无输出 | `-p` 目录下有大隐藏文件 | 检查目录大小，排除不需要的大文件 |
| 同规格任务效率不同 | 算法波动 ~30% + 不同供应商资源差异 | 属正常现象 |
| 任务调度时间长 | 供应商缓存自定义镜像或资源不足 | 耐心等待或联系技术支持 |
| 结果文件撑满系统盘 | 输出写到了系统盘而非数据盘 | 调整脚本输出路径，结果写到工作目录内 |
| 异常中断 | 机器回收或物理机故障（低概率） | 设置 `max_reschedule_times` 自动重试 |
