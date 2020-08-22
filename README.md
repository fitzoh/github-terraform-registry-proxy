# Github Terraform Tegistry Proxy
An easy way to serve your GitHub through the Terraform registry API

Don't try to use this yet.

## Access Tokens

If you want authenticated access to private github repos, you'll need to pass in an access token.
The [Terraform CLI Configuration File](https://www.terraform.io/docs/commands/cli-config.html) allows you to specify [per host credentials](https://www.terraform.io/docs/commands/cli-config.html#credentials-1): 

```hcl
credentials "your.hostname"{
  token = "your-github-token"
}
```
