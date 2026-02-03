# Installation Overview

seofor.dev is available on any Unix machine with architectures arm64 or amd64. So whether you're on MacOS, Linux, or Windows (WSL), you're all good. We do not support native Windows installation.

## Install script

seofor.dev CLI can be installed with the following command:

```bash
curl -sSfL https://seofor.dev/install.sh | bash
```

The script will:
- determine what OS and Arch you have
- download the proper binary
- install the binary (you may be prompted for your root user password)
- create the persistent config file at _~/.seo/config.yml_

Once this is done, you're all set, congratulations!

## Running the CLI

Start by checking that the CLI is properly installed with the following command:

```bash
seo --version
```

and you should see something like that:

_seo version v3.0.0_

All good. Now run the CLI by running

```bash
seo
```

## Available Commands

```bash
seo audit run              # Run localhost SEO audit
seo audit list             # List audit history
seo config                 # Show CLI configuration
seo index submit           # Submit URLs to search engines via IndexNow
seo --help                 # Show all commands
```

## Need Help?

If you encounter any issues during installation, please:

1. Send me a DM on X (Twitter) [@ugo_builds](https://x.com/ugo_builds)
2. Review the [GitHub Issues](https://github.com/ugolbck/seofordev/issues) or create a new one
