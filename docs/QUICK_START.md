# Quick Start

To get started with the `complytime` CLI, at least one plugin must be installed with a corresponding OSCAL [Component Definition](https://pages.nist.gov/OSCAL/resources/concepts/layer/implementation/component-definition/).

> Note: Some of these steps are manual and will become more automated as the project gets further in development.

## Step 1: Install ComplyTime

See [INSTALLATION.md](INSTALLATION.md)

## Step 2: Add configuration

After running `complytime list` for the first time, the complytime
directory should be created under $HOME/.config

```markdown
complytime
├── bundles
└── plugins
```

You will need an OSCAL Component Definition that defines an OSCAL Component for your target system and an OSCAL Component the corresponding
policy validation plugin. See `docs/samples/` for example configuration for the `myplugin` plugin.

```bash
cp docs/samples/sample-component-definition.json ~/.config/complytime/bundles
```

## Step 3: Install a plugin

Each plugin requires a plugin manifest. For more information about plugin discovery see [PLUGIN_GUIDE.md](PLUGIN_GUIDE.md).

```bash
cp myplugin ~/.config/complytime/plugins
checksum=$(sha256sum ~/.config/complytime/plugins/myplugin | cut -d ' ' -f 1 )
cat > ~/.config/complytime/plugins/c2p-myplugin-manifest.json << EOF
{
  "metadata": {
    "id": "myplugin",
    "description": "My complytime plugin",
    "version": "0.0.1",
    "types": ["pvp"]
  },
  "executablePath": "myplugin",
  "sha256": "$checksum"
}
EOF
```

### Using with the openscap-plugin

If using the openscap-plugin, there are two prerequisites:
- **openscap-scanner** package installed
- **scap-security-guide** package installed
