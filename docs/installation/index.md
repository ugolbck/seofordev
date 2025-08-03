# Installation Overview

seofor.dev is available on any Unix machine with architectures arm64 or amd64. So whether you're on MacOS, Linux, or Winsows (WSL), you're all good. We do not support native Windows installation.

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

Once this is done, you're all set, congratulations! 🎉

## Running the CLI

Start by checking that the CLI is properly installed with the following command:

```bash
seo --version
```

and you should see something like that:

_seo version v0.1.11_

All good 👌 Now run the TUI (Text-based User Interface) by running

```bash
seo
```

## Need Help?

If you encounter any issues during installation, please:

1. Send me a DM on X (Twitter) [@ugo_builds](https://x.com/ugo_builds)
2. Review the [GitHub Issues](https://github.com/ugolbck/seofordev/issues) or create a new one