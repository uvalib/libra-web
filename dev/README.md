# Running libra-web locally

Everything runs on this machine. No UVA or AWS services are contacted.

Requirements: Go, Node, and Docker or Podman (with compose). With rootless Podman, start its API
socket first: `systemctl --user start podman.socket`.

```
dev/local.sh up        # start postgres, meilisearch and local S3; apply DB migrations
dev/local.sh seed      # create a sample draft work for the dev user (prints its edit URL)
dev/local.sh start     # run stub services + backend (:8085) + frontend dev server; ctrl-c stops all
```

Then open http://localhost:8080. You are logged in as `dev1x` with the `admin` role. Override
with `DEV_USER`, `DEV_ROLE` (user | registrar | admin), `DEV_FIRST` and `DEV_LAST`.

Other commands: `reindex` (rebuild the search index from the local easystore), `down` (stop
containers, keep data), `reset` (stop containers and delete all local data).

## What replaces what

| Real service | Local stand-in |
|---|---|
| Postgres (libraweb DB) | `postgres` container on port 55432, migrated from `backend/db/migrations` |
| EasyStore proxy | backend talks to easystore directly (`-deveasystoredb` / `-deveasystorebucket`, needs a `-tags service` build): objects and fields in the `easystore` DB, files in a local S3 bucket |
| S3 | `s3mock` container behind the `s3proxy` nginx on port 9000 (maps `<bucket>.localhost` to path-style); download links are signed URLs to `easystore-local.localhost:9000` |
| Search index | `meilisearch` container on port 7700 |
| User service, deposit-auth, ORCID, audit, metrics | `dev/stubsvc` on port 8095 (every user lookup returns the same name; audits and metrics are empty) |
| Event bus | disabled (`-devbus`) |
| NetBadge login | `-devuser` / `-devrole` |

## Limitations

- Works are normally created by other services. Use `seed`, or the admin "register" page.
- Nothing updates the search index when works change, because that normally happens off the event
  bus. Run `dev/local.sh reindex` to refresh the dashboard and search lists.
- ORCID and deposit-auth (SIS status) features don't work.
