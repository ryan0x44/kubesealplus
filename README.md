# Kubeseal Plus

A kubeseal wrapper which makes working with Sealed Secrets and Helm a breeze.

Kubeseal Plus uses an opinionated configuration and template structure:

* For a desired Secret with name `password-secret`
* For an 'environment' called `production`
* Kubeseal Plus:
  1. assumes the Kubernetes Context is also called `production`
  2. expects a Helm Values file will have the value `.Values.environment` = `production`
  3. writes to a `SealedSecret` file stored in `templates/secret-password.production.yaml`
  4. wraps the `SealedSecret` using Helm templating, with a condition of `if eq .Values.environment "production"`

__Example SealedSecret__

Here's an example YAML manifest:

```
{{- if eq .Values.environment "production" }}
apiVersion: bitnami.com/v1alpha1
kind: SealedSecret
metadata:
    creationTimestamp: null
    name: example-secret
    namespace: example
spec:
    encryptedData:
        MESSAGE: aGVsbG8gd29ybGQK
    template:
        data: null
        metadata:
            creationTimestamp: null
            name: example-secret
            namespace: example
{{- end }}
```

Note that Kubeseal Plus will fail if the first and last line do not match this exact template, and the remainder of the file does not parse as valid YAML.

__Usage__

Rotate secrets in an existing SealedSecret:

```
kubesealplus rotate templates/secret-password.production.yaml`
```

You'll be prompted to input secret values for each existing key;
* Enter a value and press return (newline) to complete the value
* When a string literal is used, white space will be trimmed including leading
  and trailing spaces, tabs, and newline characters (per Go's strings.TrimSpace)
* Once all values are entered, you'll be asked to confirm that what you entered
  is correct
* Filename's will be auto-detected: if the string literal starts with / and then
  resolves to a valid file, the contents of that file will be used as the value
* File contents for a provided filename will be used exactly (including spaces)
* To use a filename as a string literal, prefix that filename with a space,
  and the auto-detection will not run (as first char will not be `/`, and the
  leading space will be trimmed)

__Roadmap/Future Usage__

Configure 'Kubeseal Plus keyserver' for `production` environment:

```
kubesealplus config production keyserver https://production.example.com
```

__Further Information__

See:

* [Sealed Secrets Github Repo](https://github.com/bitnami-labs/sealed-secrets)
* [Helm Templates](https://helm.sh/docs/chart_best_practices/templates/)
