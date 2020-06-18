# nginx

This example shows how `nginx` can be watched by `mon` and restarted if it becomes unhealthy.

```
# Specify --build to ensure everything is rebuilt
docker-compose up --build
```

## Notes

We do some "hackery" to make the nginx container self-break after some time...tl;dr don't use the `entrypoint.sh` for production, everything else is fine.
 