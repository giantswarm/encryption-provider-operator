# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.0] - 2025-10-09

### Changed

- Bump encryption-config-hasher to v0.2.0.

## [0.5.1] - 2024-08-13

### Fixed

- Disable logger development mode to avoid panicking

## [0.5.0] - 2024-01-17

### Added

- Consider new control-plane label.
- Add `global.podSecurityStandards.enforced` value for PSS migration.

### Changed

- Configure `gsoci.azurecr.io` as the default container image registry.

## [0.4.0] - 2023-07-13

### Fixed

- Add necessary values for PSS policy warnings.

### Changed

- Push to vsphere app collection.
- Don't push to openstack app collection.

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

[Unreleased]: https://github.com/giantswarm/encryption-provider-operator/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/giantswarm/encryption-provider-operator/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/giantswarm/encryption-provider-operator/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/giantswarm/encryption-provider-operator/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/giantswarm/giantswarm/compare/v0.4.0...v0.4.0
[0.4.0]: https://github.com/giantswarm/giantswarm/compare/v0.4.0...v0.4.0
[0.4.0]: https://github.com/giantswarm/giantswarm/compare/v0.4.0...v0.4.0
[0.4.0]: https://github.com/giantswarm/giantswarm/compare/v0.4.0...v0.4.0
[0.4.0]: https://github.com/giantswarm/giantswarm/compare/v0.3.2...v0.4.0
[0.3.2]: https://github.com/giantswarm/giantswarm/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/giantswarm/giantswarm/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/giantswarm/giantswarm/compare/v0.2.3...v0.3.0
[0.2.3]: https://github.com/giantswarm/giantswarm/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/giantswarm/giantswarm/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/giantswarm/giantswarm/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/giantswarm/giantswarm/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/encryption-provider-operator/releases/tag/v0.1.0
