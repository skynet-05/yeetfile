# Kamal config (https://kamal-deploy.org)

`.deploy.template.yml` is used to populate deployment configs for specific
environments (i.e. `deploy.dev.yml`, `deploy.prod.yml`, etc) using the
`deploy.sh` script in the root of the repo. This ensures that all .env
variables are used by the container, and that the deployment reaches the
correct server IPs.

Common deployment configuration is in `deploy.yml`.

An SSH-only yml can be added to further configure a specific env deployment
using a file named `ssh.[env].yml` (i.e. `ssh.prod.yml`). This should follow
the SSH config formatting described here:

https://kamal-deploy.org/docs/configuration/ssh/
