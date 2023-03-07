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

Note that Kubeseal Plus will fail if the first and last line do not match this 
exact template, and the remainder of the file does not parse as valid YAML.

## Usage

### Rotate

Rotate secrets in an existing SealedSecret:

```
kubesealplus rotate templates/secret-password.production.yaml
```

You'll be prompted to input secret values for each existing key then confirm 
before the file is written to;
* Enter a value and press return (newline) to complete the value
* If you want to skip rotating some keys, don't enter anything for that key and
  press return (newline)
* When a string literal is used as the value, white space will be trimmed 
  including leading and trailing spaces, tabs, and newline characters (per Go's
  strings.TrimSpace)
* Once all values are entered, you'll be asked to confirm that what you entered
  is correct
* Filename's will be auto-detected: if the string literal resolves to a valid 
  file path, the contents of that file will be used as the value
* File contents for a provided filename will be used exactly (including spaces)
* To use a filename as a string literal, prefix that filename with a space,
  and the auto-detection will not run (as first char will not be `/`, and the
  leading space will be trimmed)

### Config

Configure the Sealed Secret public key/cert URL for the `production` 
environment:

```
kubesealplus config production cert https://production.example.com
```

Note that we will automatically append `/v1/cert.pem` as a suffix to this if it
is not present.

## Sharing your Public Cert/Key

Per the `config ... cert ...` usage instructions above, this tool can fetch the
public cert/key of
your Sealed Secrets deployment from a URL.

Because Sealed Secrets will rotate this regularly, a simple way to make the 
latest version available (if your cluster is using the Traefik ingress
controller with their CRDs) is to create an [IngressRoute](https://doc.traefik.io/traefik/routing/providers/kubernetes-crd/) e.g.:

```
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
spec:
  routes:
  - match: Host(`example.com`) && (Method(`GET`) || Method(`HEAD`)) && Path(`/v1/cert.pem`)
    kind: Rule
    services:
      - name: sealed-secrets-controller
        namespace: kube-system
        port: http
```

Please take caution with the above as it will potentially expose your Sealed 
Secrets controller to the open internet.

## Support for auth via Cloudflare Access

Kubeseal Plus also supports automatically authenticating via Cloudflare Access
when fetching certificates from a URL protected by it. For example, you may
protect the IngressRoute (as outlined above) using Cloudflare Access.

Under the hood this uses [cloudflared](https://github.com/cloudflare/cloudflared)
as a library, so it works the same as as the `cloudflared login` command 
documented [here](https://developers.cloudflare.com/cloudflare-one/tutorials/cli/#authenticate-a-session-from-the-command-line).

## Further Information

See:

* [Sealed Secrets Github Repo](https://github.com/bitnami-labs/sealed-secrets)
* [Helm Templates](https://helm.sh/docs/chart_best_practices/templates/)
