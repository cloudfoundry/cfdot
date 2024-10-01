---
title: Running from a BOSH deployed VM
expires_at : never
tags: [diego-release, cfdot]
---

## Running from a BOSH-deployed VM

`cfdot` is most useful in the context of a running Diego deployment.  If you
use the [`generate-deployment-manifest`](https://github.com/cloudfoundry/diego-release/blob/master/scripts/generate-deployment-manifest)
script in diego-release to generate your Diego manifest, `cfdot` is already
available on the BOSH-deployed Diego VMs. To use it:

```bash
bosh ssh <DIEGO_JOB>/<INDEX>
cfdot
```

The `cfdot` pre-start script installs the `setup` script into `/etc/profile.d`.
This `setup` script does 3 things:

- Exports environment variables to target the BBS API in the deployment.
- Puts the `cfdot` binary on the `PATH`.
- Puts a `jq` binary on the `PATH`.
