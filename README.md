# Bus

![Status](https://img.shields.io/badge/status-planning-blue)
![License](https://img.shields.io/github/license/hyperifyio/bus)

Bus is a **modular, CLI-first accounting core** for Finnish small businesses.
It provides a **double-entry, append-only ledger** with VAT (ALV) reporting,
period closing, and audit trails compliant with Finnish accounting law. The
system is **offline-first** and **VCS-friendly** (Bus does not run Git
operations) and is designed to expose a matching REST API with OpenAPI.

## Table of contents
- [Start here](#start-here)
- [Features](#features)
- [Architecture at a glance](#architecture-at-a-glance)
- [Installation](#installation)
- [Usage](#usage)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Tests](#tests)
- [Support](#support)
- [Credits](#credits)
- [License](#license)
- [Project status](#project-status)

## Start here
- **Docs entry point**: `docs/README.md`
- **Design plan**: `docs/spec/accounting-finnish-smb.md`
- **Roadmap start (authoritative)**: `docs/roadmap/0.0.1.md`

## Features
- Double-entry, append-only ledger with audit trail.
- VAT (ALV) reporting by rate and period.
- Period close locks with adjustment entries.
- Sales and purchase invoicing tied to ledger entries.
- Offline-first storage with VCS-friendly formats.
- Modular services with CLI-first and REST API parity.

## Architecture at a glance
- **Core**: orchestrates modules and storage backends only.
- **Modules**: ledger, chart of accounts, invoicing, VAT, bank imports, budgets.
- **Transports**: CLI first, REST API later (OpenAPI).
- **Storage**: filesystem `.bus/` or database backend.

## Installation
Bus is in a **planning-first** stage with no runnable binaries yet.

To read the docs locally:
```shell
git clone https://github.com/hyperifyio/bus.git
cd bus
```

## Usage
Planned CLI examples (not implemented yet):
```shell
bus init "Acme Oy"
bus account add "3000 Sales"
bus entry post --date 2026-02-15 --lines "1100:1240" "3000:1000" "2930:240"
bus vat report 2026-02
```

## Roadmap
- See `docs/README.md` for the authoritative roadmap index.

## Contributing
Contributions are welcome. Please read `CONTRIBUTING.md` and follow the
workflow and traceability requirements described there. Participation is
governed by `CODE_OF_CONDUCT.md`.

## Tests
No automated test suite exists yet because there is no implementation.
When code lands, this section will document the standard test command.

## Support
Use GitHub Issues for bugs and feature requests:
`https://github.com/hyperifyio/bus/issues`

## Credits
Maintainer: Hyperify.io. Significant contributors are listed in git history.

## License
See `LICENSE`.

## Project status
Planning-only and pre-alpha. No production use yet.
