[![Build Status][build-actions-svg]][build-actions]
[![Lint Status][lint-actions-svg]][lint-actions]
[![Test Status][test-actions-svg]][test-actions]
[![codecov][codecov-svg]][codecov-url]
---

# OpenThread Network Simulator 2

OpenThread Network Simulator 2 (OTNS2) simulates Thread networks using OpenThread POSIX instances
and provides visualization and management of those simulated networks.

Note: this is a fork of the [OpenThread OTNS project](https://github.com/openthread/ot-ns), made by 
[IoTconsultancy.nl](https://www.iotconsultancy.nl/). It offers additional features including:

* Support for more accurate RF simulation of OpenThread nodes. This uses the OpenThread platform 
  `ot-rfsim`, which specifically supports RF simulation, for OT nodes. This C code is included.
* Selectable radio (RF propagation) models with tunable RF parameters.
* Tunable radio parameters on each individual OT node. For example, CSL parameters or Rx sensitivity.
* Control of logging display from OT-node, using `log` and `watch` CLI commands. Logging to file per 
  OT-node. The logging output can include all enabled OT-node log items.
* Detailed logging options for RF operations (at log-level 'trace') performed in the simulated radio,
  at 1 us resolution.
* Animations with duration scaled to the actual time duration of a packet in flight (at low simulation 
  speed only).
* Support for easily adding various node types (1.1, 1.2, 1.3, 1.4, 1.4 Border Router)
* New graphical displays for overall node type statistics, and energy usage (beta - 
  contribution by [Vinggui](https://github.com/Vinggui)).
* Extensive set of Python scripts for unit testing, examples, and case studies.
* Various UI look & feel improvements.

More information about Thread can be found at [threadgroup.org](http://threadgroup.org/). 
Thread is a registered trademark of the Thread Group, Inc.

To learn more about OpenThread, visit [openthread.io](https://openthread.io).

[build-actions-svg]: https://github.com/EskoDijk/ot-ns/workflows/Build/badge.svg?branch=main&event=push
[build-actions]: https://github.com/EskoDijk/ot-ns/actions?query=workflow%3ABuild+branch%3Amain+event%3Apush
[lint-actions-svg]: https://github.com/EskoDijk/ot-ns/workflows/Lint/badge.svg?branch=main&event=push
[lint-actions]: https://github.com/EskoDijk/ot-ns/actions?query=workflow%3ALint+branch%3Amain+event%3Apush
[test-actions-svg]: https://github.com/EskoDijk/ot-ns/workflows/Test/badge.svg?branch=main&event=push
[test-actions]: https://github.com/EskoDijk/ot-ns/actions?query=workflow%3ATest+branch%3Amain+event%3Apush
[codecov-svg]: https://codecov.io/gh/EskoDijk/ot-ns/branch/main/graph/badge.svg
[codecov-url]: https://codecov.io/gh/EskoDijk/ot-ns

## Get started
See [GUIDE](GUIDE.md) to get started. 

See [OTNS CLI Reference](cli/README.md) for the OTNS CLI commands.

## Contributing

We would love for you to contribute to OTNS and help make it even better than it is today!
See our [Contributing Guidelines](CONTRIBUTING.md) for more information.

Contributors are required to abide by our [Code of Conduct](CODE_OF_CONDUCT.md) and 
[Coding Conventions and Style Guide](CONTRIBUTING.md#coding-conventions-and-style).

## Version

OTNS follows the [Semantic Versioning guidelines](http://semver.org/) for release cycle transparency 
and to maintain backwards compatibility. 

## License

OTNS is released under the [BSD 3-Clause license](LICENSE). See the [`LICENSE`](LICENSE) file for more information.
