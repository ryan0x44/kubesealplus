# Kubeseal Plus

A kubeseal wrapper which makes working with Sealed Secrets and Helm a breeze.

Kubeseal Plus uses an opinionated configuration and template structure:

* For a desired Secret with name `password-secret`
* For an 'environment' called `production`
* Kubeseal Plus:
  1. assumes the Kubernetes Context is also called `production`
  2. expects a Helm Values file will have the value `.Values.environment` = `production`
  3. writes to a `SealedSecret` file stored in `templates/secret-password.production.yaml`
  4. wraps the `SealedSecret` using Helm templating, with a condition of `if eq .Values.environment production`

__Example SealedSecret__

Here's an example YAML manifest:

```

```

Note that Kubeseal Plus will fail if the first and last line do not match this exact template, and the remainder of the file does not parse as valid YAML.

__Usage__

Rotate secrets in an existing SealedSecret:

```
kubesealplus rotate templates/secret-password.production.yaml`
```

__Roadmap/Future Usage__

Configure 'Kubeseal Plus keyserver' for `production` environment:

```
kubesealplus config production keyserver https://production.example.com
```

__Further Information__

See:

* [Sealed Secrets Github Repo](https://github.com/bitnami-labs/sealed-secrets)
* [Helm Templates](https://helm.sh/docs/chart_best_practices/templates/)
