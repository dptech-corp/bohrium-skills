---
name: bohrium-project
compatibility: "Requires bohr CLI (npm @dptech-corp/bohr-cli); verified with 2.9.19. Account operations require Bohrium authentication and network access."
description: >-
  Create, list, inspect, or delete Bohrium projects and list their members with
  `bohr project`. Use for 项目管理, project IDs, project ownership, creation-time
  monthly/total cost limits, member visibility, and project deletion. Use File
  commands for share/ data and resource commands for project-owned compute.
---

# Bohrium Project

Resolve projects and IDs before using them in Dataset, File, Job, Node, Image,
Notebook, or Sandbox workflows:

```bash
bohr project list -o json
bohr project get <project_id> -o json
bohr project members <project_id> -o json
```

Project commands are currently free. Create is a write and accepts cost limits
at creation time:

```bash
bohr project create --name research-project \
  --total_cost_limit 1000 --month_cost_limit 200 -o json
```

Treat `0` according to the platform response rather than assuming it means a
particular unlimited or disabled policy. Preserve the returned project ID; a
project name, member user ID, and project ID are not interchangeable.

Delete is destructive and can affect access to project-owned resources. Inspect
the project and members, identify dependent share paths or compute resources,
then obtain exact-target authorization:

```bash
bohr project delete <project_id> --yes -o json
```

The current CLI cannot rename a project, change limits after creation, or add,
remove, promote, demote, or recover members. Do not translate those upstream
service operations into invented flags.

Read [`references/commands.md`](references/commands.md) for every command,
creation input, project-ID handoff, and deletion boundary.
