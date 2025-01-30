# complytime

[![OpenSSF Best Practices status](https://www.bestpractices.dev/projects/9761/badge)](https://www.bestpractices.dev/projects/9761)
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/complytime/complytime)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/complytime/complytime/badge)](https://scorecard.dev/viewer/?uri=github.com/complytime/complytime)

ComplyTime leverages [OSCAL](https://github.com/usnistgov/OSCAL/) to perform compliance assessment activities, using plugins for each stage of the lifecycle.

## Install

### Prerequisites

- **Go** version 1.20 or higher
- **Make** (optional, for using the `Makefile` if included)

#### Using with the openscap-plugin
- **openscap-scanner** package installed
- **scap-security-guide** package installed

### Clone the repository

```bash
git clone https://github.com/complytime/complytime.git
cd complytime
```

### Build Instructions
To compile complytime and openscap-plugin:

```bash
make build
```

The binaries can be found in the `bin/` directory in the local repo. Add it to your PATH and you are all set!

## Usage

### Install a plugin

_Note: This is currently a manual process_

```markdown
After running complytime for the first time, the complytime
directory should be created under $HOME/.config

complytime
├── bundles
└── plugins
```


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

### Create configuration

Plugin selection and input is configured through [OSCAL Component Definitions](https://pages.nist.gov/OSCAL-Reference/models/v1.1.3/component-definition/json-outline/).
See `docs/samples/` for example configuration for the `myplugin` plugin.

```bash
cp docs/samples/sample-component-definition.json ~/.config/complytime/bundles
```

### Run ComplyTime

#### Create an Assessment Plan
```bash
complytime plan "example"
cat assessment-plan.json
```

#### Generate artifacts and Run
```bash
complytime generate
complytime scan
```

## Community and Contribution

- [Contributing](./docs/CONTRIBUTING.md)
- [Style Guide](./docs/STYLE_GUIDE.md)
- [Code of Conduct](./docs/CODE_OF_CONDUCT.md)
