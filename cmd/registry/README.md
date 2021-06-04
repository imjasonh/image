# Read-only OCI layout registry

This registry serves a read-only registry API server backed by an OCI layout

Populate the layout:

```
$ crane pull ubuntu layout --format=oci
```

Run the server:

```
$ go run ./cmd/registry
Serving on 8080...
```

Hit the registry:

```
$ crane manifest localhost:8080/ubuntu@sha256:86ac87f73641c920fb42cc9612d4fb57b5626b56ea2a19b894d0673fd5b4f2e9
{
   "schemaVersion": 2,
   "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
   "config": {
      "mediaType": "application/vnd.docker.container.image.v1+json",
      "size": 3312,
      "digest": "sha256:7e0aa2d69a153215c790488ed1fcec162015e973e49962d438e18249d16fa9bd"
   },
   "layers": [
      {
         "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
         "size": 28539626,
...
```
