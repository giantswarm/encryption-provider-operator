# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.2] - 2023-03-06

### Changed

- Add finalizer before reconciliation
- Use patch to add/remove finalizers insted of update

## [0.3.1] - 2023-03-03

### Added

- Add the use of the runtime/default seccomp profile. Allow required volume types in PSP so that pods can still be admitted.

## [0.3.0] - 2023-01-31

### Changed

- Upgrade CAPI dependency to `v1.2.7` and use `v1beta1` CAPI CRDs.

### Fixed

- Force go module to use `golang.org/x/text v0.3.8` to fix `CVE-2022-32149`.

## [0.2.3] - 2022-08-05

## [0.2.2] - 2022-06-21

### Changed

- Bump `encryption-config-hasher` to 0.1.2.

## [0.2.1] - 2022-06-21

### Fixed

- Fix typo in helm values (`fromRelase` became `fromRelease`).

## [0.2.0] - 2022-06-14

### Changed

- Add to azure app collection.

## [0.1.0] - 2022-01-31

### Added

- Add main reconciliation loop logic.
- Implement key rotation logic.

[Unreleased]: https://github.com/giantswarm/giantswarm/compare/v0.3.2...HEAD
[0.3.2]: https://github.com/giantswarm/giantswarm/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/giantswarm/giantswarm/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/giantswarm/compare/v0.2.3...v0.3.0
[0.2.3]: https://github.com/giantswarm/giantswarm/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/giantswarm/giantswarm/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/giantswarm/giantswarm/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/giantswarm/giantswarm/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/encryption-provider-operator/releases/tag/v0.1.0
