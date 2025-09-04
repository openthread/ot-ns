[![Build Status][build-actions-svg]][build-actions] [![Lint Status][lint-actions-svg]][lint-actions] [![Test Status][test-actions-svg]][test-actions] [![codecov][codecov-svg]][codecov-url]

---

# OpenThread Network Simulator

OpenThread Network Simulator (OTNS) simulates Thread networks using OpenThread POSIX instances and provides visualization and management of those simulated networks.

More information about Thread can be found at [threadgroup.org](http://threadgroup.org/). Thread is a registered trademark of the Thread Group, Inc.

To learn more about OpenThread, visit [openthread.io](https://openthread.io).

# Features

Note: this is version 2.x of OTNS. It offers additional features compared to version 1:

- Support for more accurate RF simulation of OpenThread nodes. This uses the OpenThread platform `ot-rfsim`, which specifically supports RF simulation for OT nodes. This C code is included.
- Selectable radio (RF propagation) models with tunable RF parameters.
- Run-time tunable radio parameters on each individual OT node. For example, CSL parameters or Rx sensitivity.
- Control of logging display from OT-node, using `log` and `watch` CLI commands. Logging to file per OT-node. The logging output can include any enabled OT-node log items.
- Detailed logging options for RF operations (at log-level 'trace') performed in the simulated radio, at 1 us resolution.
- Reproducible simulations by selection of a seed value for all pseudo-random number generators.
- See packets in flight: animations in the GUI with a duration scaled to the actual time duration of a packet in flight (works at low simulation speed only).
- Support for easily adding various Thread node types (1.1, 1.2, 1.3, 1.4, 1.4 Border Router).
- New graphical displays for overall node type statistics, and energy usage (beta - contribution by [Vinggui](https://github.com/Vinggui)).
- Extended set of Python scripts for unit testing, examples, and case studies.
- Key Performance Indicators (KPI) module that tracks counters and statistics for all nodes.
- Loading/saving of network topologies in YAML files.
- Custom startup scripts with OT CLI commands, defined in a YAML file.
- Additional ["WPAN-TAP"](https://exegin.com/wp-content/uploads/ieee802154_tap.pdf) PCAP format that captures channel information.
- Various UI look & feel improvements.

[build-actions-svg]: https://github.com/openthread/ot-ns/workflows/Build/badge.svg?branch=main&event=push
[build-actions]: https://github.com/openthread/ot-ns/actions?query=workflow%3ABuild+branch%3Amain+event%3Apush
[lint-actions-svg]: https://github.com/openthread/ot-ns/workflows/Lint/badge.svg?branch=main&event=push
[lint-actions]: https://github.com/openthread/ot-ns/actions?query=workflow%3ALint+branch%3Amain+event%3Apush
[test-actions-svg]: https://github.com/openthread/ot-ns/workflows/Test/badge.svg?branch=main&event=push
[test-actions]: https://github.com/openthread/ot-ns/actions?query=workflow%3ATest+branch%3Amain+event%3Apush
[codecov-svg]: https://codecov.io/gh/openthread/ot-ns/branch/main/graph/badge.svg
[codecov-url]: https://codecov.io/gh/openthread/ot-ns

# Getting started

See [GUIDE](GUIDE.md) to get started with a local install of OTNS.

See [OTNS CLI Reference](cli/README.md) for the OTNS CLI commands.

To do a quick try-out of OTNS without installing any software locally, you can fetch and run the playground Docker image:

```bash
docker run -it -p 8997-9000:8997-9000 openthread/otns-playground
```

# Contributing

We would love for you to contribute to OTNS and help make it even better than it is today! See our [Contributing Guidelines](CONTRIBUTING.md) for more information.

Contributors are required to abide by our [Code of Conduct](CODE_OF_CONDUCT.md) and [Coding Conventions and Style Guide](CONTRIBUTING.md#coding-conventions-and-style). See [AUTHORS](AUTHORS) for the list of present authors.

# Versioning

OTNS follows the [Semantic Versioning guidelines](http://semver.org/) for release cycle transparency and to maintain backwards compatibility.

# License

OTNS is released under the [BSD 3-Clause license](LICENSE). See the [`LICENSE`](LICENSE) file for more information.

Please only use the OpenThread name and marks when accurately referencing this software distribution. Do not use the marks in a way that suggests you are endorsed by or otherwise affiliated with Nest, Google, or The Thread Group.

# Need Help?

OpenThread support is available on GitHub:

- Bugs and feature requests — [submit to the Issue Tracker](https://github.com/openthread/ot-ns/issues)
- Community Discussion - [ask questions, share ideas, and engage with other community members](https://github.com/openthread/openthread/discussions)
