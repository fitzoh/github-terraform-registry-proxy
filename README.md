# Github Terraform Registry Proxy
An easy way to serve your GitHub hosted Terraform Modules through the Terraform registry API.

## Configuration
Let's say you're running your registry proxy at `https://my-registry-proxy.com` and you want to access an terraform module hosted at `github.com/my-github-org/terraform-aws-something-module` (specifically tag `v1.2.3`).

```hcl-terraform
module "my-module" {
  source = "my-registry-proxy.com/my-github-org/something-module/aws"
  version = "1.2.3"
}
```

As documented [here](https://www.terraform.io/docs/registry/modules/use.html#private-registry-module-sources), the source string uses the form `<HOSTNAME>/<NAMESPACE>/<NAME>/<PROVIDER>`.
The hostname is of course the hostname of the registry, the namespace maps to the github org hosting the module, and the name and provider combine to form the repo name using the convention defined [here](https://www.terraform.io/docs/registry/modules/publish.html#requirements) (`terraform-<PROVIDER>-<NAME>`).
Modules must be tagged to be used with the registry.

## Access Tokens
If you want to access private GitHub repos the registry needs a way to authenticate with GitHub.
Currently this is accomplished by having the Terraform CLI provide a GitHub access token which is used in subsequent requests to the GitHub API.

To enable authenticated access to the GitHub API, acquire a [personal access token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token) and add it to a [credentials block](https://www.terraform.io/docs/commands/cli-config.html#credentials-1) in the [Terraform CLI Configuration File](https://www.terraform.io/docs/commands/cli-config.html).

```hcl
credentials "your.hostname"{
  token = "<your-access-token>"
}
```
