# Index remote `/share` and `/personal` storage before repeated searches

`/share` and `/personal` can be network-backed mounts. Repeated recursive
`find`, `du`, `grep -R`, or content-searching `rg` calls over them can consume
substantial network bandwidth and make both the agent and other users appear
stuck.

## Decision rule

1. Start with an exact path or one narrow, depth-limited lookup when the likely
   directory is known.
2. If the task needs repeated broad **filename** searches, build one path-only
   index for the narrowest useful subtree and search that local index.
3. Do not index file contents or hashes. Read only the exact live files needed
   after locating their paths.
4. Before deleting, overwriting, moving, changing permissions, or making a
   billing/resource decision, verify the selected path against the live mount
   with `stat`/`realpath`. A cached index is discovery data, not authority.

Do not blindly index all of `/share` or `/personal` when a project or user
subdirectory is known. Never follow symlinks out of the selected root and never
cross into another mount.

## Lightweight path index

The following Linux-shell example keeps a NUL-delimited, path-only cache under
the sandbox-local `/tmp`. It uses mode `0700`, never writes alongside user
files, refreshes after 10 minutes, and keys the cache by both the canonical root
and its current mount identity.

```bash
root="$(realpath -e -- /share/my-project)" || exit 1
case "$root" in
  /share|/share/*|/personal|/personal/*) ;;
  *) echo "refusing non-remote-storage root: $root" >&2; exit 2 ;;
esac

cache_dir=/tmp/bohr-storage-index
umask 077
mkdir -p -- "$cache_dir"
chmod 700 -- "$cache_dir"

mount_sig="$(findmnt -n -o TARGET,SOURCE,FSTYPE -T "$root" 2>/dev/null)" || exit 1
cache_key="$(printf '%s\0%s' "$root" "$mount_sig" | sha256sum | awk '{print $1}')"
index="$cache_dir/$cache_key.paths0"
now="$(date +%s)"
mtime="$(stat -c %Y -- "$index" 2>/dev/null || printf 0)"

if [ ! -s "$index" ] || [ $((now - mtime)) -ge 600 ]; then
  tmp="$(mktemp "$cache_dir/.paths.XXXXXX")" || exit 1
  if find -P "$root" -xdev -mindepth 1 -printf '%P\0' >"$tmp"; then
    chmod 600 -- "$tmp"
    mv -f -- "$tmp" "$index"
  else
    rm -f -- "$tmp"
    exit 1
  fi
fi

# Search locally. `--` prevents a query beginning with '-' becoming an option.
needle='input.dat'
grep -zFi -- "$needle" "$index" | tr '\0' '\n' | sed -n '1,50p'
```

If `findmnt` is unavailable, use a stable, non-secret mount identifier from
`stat -c '%d' -- "$root"` in the cache key. Do not copy mount credentials,
tokens, or mount options into the index.

## Refresh and invalidation

Refresh immediately, without waiting for the TTL, when:

- the agent created, uploaded, downloaded, renamed, moved, or removed relevant
  files;
- the mount was remounted or its identity changed;
- a cached match fails the live `stat` check; or
- a lookup miss conflicts with a path the user says currently exists.

Network filesystems do not guarantee a useful directory mtime for every remote
change, so directory mtime alone is not a valid invalidation signal. A ten
minute TTL is only a default: shorten it for actively changing workflows, and
discard the cache at the end of the sandbox or when it is no longer needed.

## Safety boundaries

- Keep the cache sandbox-local and private. Do not put it on `/share` or
  `/personal`, commit it into an image, upload it, or expose it through a public
  port.
- Store relative paths only. Do not add file contents, hashes, credentials,
  signed URLs, ownership metadata, or secrets to the index.
- Do not reuse an index across users, mounts, sandboxes, or canonical roots.
- Treat path names as potentially sensitive. Delete the local cache when the
  task is complete if it is no longer needed.
- A cache miss never proves a file does not exist. Refresh once and then perform
  a narrow live lookup before reporting absence.
