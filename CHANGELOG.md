# Changelog

## [2.3.0](https://github.com/servasec/servasec/compare/v2.2.0...v2.3.0) (2026-07-17)


### 🧩 Features

* **frontend:** Onboarding tour for new users ([cf609e7](https://github.com/servasec/servasec/commit/cf609e7aa4c1b2ca08857ce2a9d824286b988792))
* **migrations:** Add FK constraints + soft delete support ([1bc04ce](https://github.com/servasec/servasec/commit/1bc04ce672cf9882ebf338c87eb85d845173f4f0))
* **models:** Add soft delete (gorm.DeletedAt) to User, Group, Team ([2788063](https://github.com/servasec/servasec/commit/27880638a75493fae32b32b6d7a1650227e94817))
* **parsers:** Add bandit, gosec, kube-bench, kubescape, npm-audit, osv-scanner, tfsec parsers ([c777e08](https://github.com/servasec/servasec/commit/c777e08ac701ee74e2397fcf277fb9c926809a25))


### 🐛 Bug fixes

* **parsers:** Fix grype detection signature, bandit CWE handling ([6e73bc5](https://github.com/servasec/servasec/commit/6e73bc5da9bb180f2c07e9be0c86415144bf5eb2))
* **security:** Add DB transactions, batch insert, file size limits ([fb92da1](https://github.com/servasec/servasec/commit/fb92da193872321f27b534f5940bdbce81effc2a))
* **security:** Harden auth, finding, team, webhook, policy controllers ([b439e9e](https://github.com/servasec/servasec/commit/b439e9e468380149def6802de77d20c74287ca46))
* **seeder:** Remove hardcoded fallback, fatalf on missing admin ([c501428](https://github.com/servasec/servasec/commit/c5014286a80eab37fb4b4e25e09c8e20ea63525e))

## [2.2.0](https://github.com/servasec/servasec/compare/v2.1.4...v2.2.0) (2026-07-07)


### 🧩 Features

* Upgrade process reviewed ([3020fab](https://github.com/servasec/servasec/commit/3020fab9b2a695a47175d406c51815fbff6180cd))
* Upgrade process reviewed pt2 ([f3672fc](https://github.com/servasec/servasec/commit/f3672fce23c044512be2d52b860107d4f30e4708))


### 🐛 Bug fixes

* **helm:** Caddy removed from helm chart; traefik is now the official ingress controller we support ([c773911](https://github.com/servasec/servasec/commit/c7739114e18015ed586c850f05418575520fc28c))

## [2.1.4](https://github.com/servasec/servasec/compare/v2.1.3...v2.1.4) (2026-07-06)


### 🐛 Bug fixes

* **backend:** Json security + opengrep false positive ([4052cfb](https://github.com/servasec/servasec/commit/4052cfbbea24a12dd5febb48f8e625fc693441cb))


### 🤖 Continuous integration

* Final workflows trigger ([f8b1b80](https://github.com/servasec/servasec/commit/f8b1b803db575491ceb0e1c7776aff8a10713e45))

## [2.1.3](https://github.com/servasec/servasec/compare/v2.1.2...v2.1.3) (2026-07-06)


### 🤖 Continuous integration

* Use custom Github APP for release workflow ([ace3941](https://github.com/servasec/servasec/commit/ace3941c98a22a3b1c1f4c7bd832939df3025aad))

## [2.1.2](https://github.com/servasec/servasec/compare/v2.1.1...v2.1.2) (2026-07-06)


### 🤖 Continuous integration

* Workflow trigger only on version tag ([a7c41a3](https://github.com/servasec/servasec/commit/a7c41a359bbf2af80ccfb2b009f1c452fcd10abe))
* Workflow trigger only on version tag ([2874590](https://github.com/servasec/servasec/commit/28745901db497f934a8fce51f6903336c507d42f))

## [2.1.1](https://github.com/servasec/servasec/compare/v2.1.0...v2.1.1) (2026-07-06)


### 🤖 Continuous integration

* Fix docker image workflows ([0f40e08](https://github.com/servasec/servasec/commit/0f40e0807d246f4eea23a14bececd40d065ee9f1))
* Fix docker image workflows ([8077636](https://github.com/servasec/servasec/commit/8077636663f8b82f28a8249c32c397ab5eb5c27c))

## [2.1.0](https://github.com/servasec/servasec/compare/v2.0.0...v2.1.0) (2026-07-06)


### 🧩 Features

* **kubernetes:** Helm chart 0.1.0 ([c1e8040](https://github.com/servasec/servasec/commit/c1e80409f81aeaa6daf2b3525a6de0664aeacaaf))


### 🐛 Bug fixes

* **backend:** Sql unique version index ([b7dd87b](https://github.com/servasec/servasec/commit/b7dd87badc84b67a60e79b25acaf88486f7096c1))
* Caddy service missing var ([4e7778e](https://github.com/servasec/servasec/commit/4e7778edc718a90f37456f2a26d056ca2fc80da5))
* Dead env var removed ([34cd9c3](https://github.com/servasec/servasec/commit/34cd9c335268d57b184b49741f957b1d97149c60))


### 🤖 Continuous integration

* Docker-pro wokflow secrets to env ([5576414](https://github.com/servasec/servasec/commit/557641464b0342332111e772cfb9162ac00fef0f))
* Docker-pro workflow fix ([9a5cb59](https://github.com/servasec/servasec/commit/9a5cb59b0e4bb409fbd5b3c5a25d203ed0854e0f))
* Fix security-reports ([3e198a6](https://github.com/servasec/servasec/commit/3e198a65c86a3893f461ced21e74e01dc261fa03))
* Ghcr images ([d80533b](https://github.com/servasec/servasec/commit/d80533b7046f76f363a48e3627de639ced75cb97))
* Helm chart release ([429e67a](https://github.com/servasec/servasec/commit/429e67a2a8d4f50c9c26724146d950d36e8a69a9))

## [2.0.0](https://github.com/servasec/servasec/compare/v1.0.0...v2.0.0) (2026-07-02)


### ⚠ BREAKING CHANGES

* **backend:** SSC_ADMIN_PASSWORD & CSRF_SECRET mandatories in prod env + GORM AutoMigrate deleted

### 🧩 Features

* **backend:** SSC_ADMIN_PASSWORD & CSRF_SECRET mandatories in prod env + GORM AutoMigrate deleted ([94dfcb8](https://github.com/servasec/servasec/commit/94dfcb8b738df07306f5fa671e4adf4babdf268e))
* New database migration system with goose & upgrade script + md ([47d8abf](https://github.com/servasec/servasec/commit/47d8abfb09e406908e3ae12cf10dbe5505fb342d))


### 🐛 Bug fixes

* **frontend:** Dashboard bar chart color ([bac0287](https://github.com/servasec/servasec/commit/bac02870cde337c9aa19b2708ba647083ad2c814))

## [1.0.0](https://github.com/servasec/servasec/compare/v0.3.1...v1.0.0) (2026-07-02)


### ⚠ BREAKING CHANGES

* **backend:** /applications/by-slug/:slug replaced by /groups/:groupPath/applications/:slug

### 🧩 Features

* **backend:** /applications/by-slug/:slug replaced by /groups/:groupPath/applications/:slug ([79ca7d8](https://github.com/servasec/servasec/commit/79ca7d8bdd302a74cd606cac86df00109806d5ba))
* **backend:** Implement openapi annotations ([bd940cb](https://github.com/servasec/servasec/commit/bd940cbd0311c7e74676f01060c5b2a2037d0812))
* **backend:** Implement openapi annotations ([894e32f](https://github.com/servasec/servasec/commit/894e32f751bda0dda181bf169de9ea9f02f739e8))
* Better ingest methods for CI/CD processes ([e24911c](https://github.com/servasec/servasec/commit/e24911c14387f4538a1511b185509948dd64a652))
* Sarif parser ([036dfbc](https://github.com/servasec/servasec/commit/036dfbca61cbe7c433e0a56cc5f3c7450e9b310b))
* Sarif parser ([9c1c00e](https://github.com/servasec/servasec/commit/9c1c00e6ed37790a1eb228f94ee190c54e2c4864))


### 🐛 Bug fixes

* API ingest processes + user api key ([f9d741e](https://github.com/servasec/servasec/commit/f9d741e9cfc95fa2d76de60d4420923c73ea34cb))
* **frontend:** Scans & findings pages filters rework ([93a378f](https://github.com/servasec/servasec/commit/93a378fce2a1b129db42e2568ff0d4c12ad6f9bd))
* **frontend:** Toast UI revamp ([b22d465](https://github.com/servasec/servasec/commit/b22d465d4616172864ec7ca9c8fc7080c317b28b))

## [0.3.1](https://github.com/servasec/servasec/compare/v0.3.0...v0.3.1) (2026-06-30)


### 🤖 Continuous integration

* Fix release please target branch ([55b22e9](https://github.com/servasec/servasec/commit/55b22e989b27ad02e5cf338936885eb078196684))
* Fix release please target branch ([56ea036](https://github.com/servasec/servasec/commit/56ea036de678eb006f8c64ee81d6d1abe12ebcdf))
* Release please target branch in action input ([0471dd9](https://github.com/servasec/servasec/commit/0471dd988dd84598be4044db0522e11b1f5250e6))
* Release please target branch in action input ([7a0f669](https://github.com/servasec/servasec/commit/7a0f669560e601e28cdb5f2e714c8d09735573ae))

## [0.3.0](https://github.com/servasec/servasec/compare/v0.2.0...v0.3.0) (2026-06-30)


### 🧩 Features

* Add audit log middleware ([dd1816a](https://github.com/servasec/servasec/commit/dd1816a8ba571375eaea9843994ddf228dbddfdc))
* Add audit log page, user search component, and risk scoring widgets to frontend ([e7e0305](https://github.com/servasec/servasec/commit/e7e030590174ce6e173e76bf4bf4edfe3ed53b1d))
* Add Gitleaks, TruffleHog, and Trivy parsers with tests ([db0c505](https://github.com/servasec/servasec/commit/db0c50538a703d9a5d7cb4071a09e1c17542ba3f))
* Add license-based feature gating system ([711f235](https://github.com/servasec/servasec/commit/711f235d4fa4e723a802a9622fd81284440c8e1f))
* Add MCP server with SSE and Streamable HTTP ([3d65701](https://github.com/servasec/servasec/commit/3d657017feda4f641549aaa5076cd22e64a1f022))
* Add Nuclei (DAST) scanner parser with tests ([d264ec4](https://github.com/servasec/servasec/commit/d264ec459c4dda60c050f52194577b582cbd5507))
* Add OAuth 2.0 authorization server for MCP ([c4b7311](https://github.com/servasec/servasec/commit/c4b7311869c55f9fbf8f0421aec003c1ef40dfba))
* Add Podman Quadlet deployment ([66e17ea](https://github.com/servasec/servasec/commit/66e17eabda7ebec5dd3f9a5ea85c18d07c2d6021))
* Add risk scoring engine with EPSS integration ([c9ed358](https://github.com/servasec/servasec/commit/c9ed358755f6672cfcfb6df6cbe85e7028128882))
* Add the possibility to enable/disable scanner type for administrators ([c0cb08a](https://github.com/servasec/servasec/commit/c0cb08acbacf32c9a4c568327a2af854ac9797ca))
* Complete base stack ([931e088](https://github.com/servasec/servasec/commit/931e0889a87b12e7adcc0c2bc563133dcbe36ec1))
* Deduplicate findings on ingest via dedupe_hash ([b26ac4d](https://github.com/servasec/servasec/commit/b26ac4d748aba845b9b7a364fe2539c6428950e0))
* **infra:** Route OAuth, well-known, and MCP traffic through Caddy ([22c20d8](https://github.com/servasec/servasec/commit/22c20d80e30fd9b6e05744a4cd2af0198e597351))
* Polcies & webhooks frontend ([a03e3ab](https://github.com/servasec/servasec/commit/a03e3ab55977440569f4638897eb5e7ef244eb76))
* Polcies & webhooks frontend ([766a663](https://github.com/servasec/servasec/commit/766a66320fd8a1b208ed8bdd2903a9e22fca7889))
* Policies + alerting ([9074114](https://github.com/servasec/servasec/commit/9074114dd1e58abbfe8f121a19998d1e1fa2d09e))
* Pro feat exported ([cd639f4](https://github.com/servasec/servasec/commit/cd639f4cf9a9d532c0d4c62870328b254052ae2e))
* Sso implementation ([cdc380d](https://github.com/servasec/servasec/commit/cdc380dec33aaf036ed7b7ac3f3713a615235bc9))
* Wire up features, risk scoring, MCP, and audit log in main; add computed fields to models ([39490f5](https://github.com/servasec/servasec/commit/39490f535eb8e5046879e896c0aef48b49c2d581))


### 🐛 Bug fixes

* Add Bearer token support to CheckPolicy and add resource-level access middleware ([dc31c5c](https://github.com/servasec/servasec/commit/dc31c5c055f8bf8cc8723dc322abdcdba6147d96))
* Add target-branch main to release-please config ([19f23f8](https://github.com/servasec/servasec/commit/19f23f8c41fa0dcc886597ea43d52bccdad6fac3))
* **dedup:** Dedup didn't take appVersion in criteria ([2d821b2](https://github.com/servasec/servasec/commit/2d821b2f3ec2f7a69b79a7a4471d09e5686d47fa))
* Global darkmode fix ([0710c67](https://github.com/servasec/servasec/commit/0710c674a891311f3c09ddc1073652d545ec4edd))
* Mcp & middleware security fix ([16fcafd](https://github.com/servasec/servasec/commit/16fcafd822725bc4dd25410d125a10a9de31e6f0))
* Seeder & ingest errors on build ([a5d9b22](https://github.com/servasec/servasec/commit/a5d9b22a32371036d9c79035606014d270bdf532))
* Servasec-pro remediation ([169b0f9](https://github.com/servasec/servasec/commit/169b0f94dab18d05ec6e9edc2f4d783b6f997df2))
* Sso controller ([f0aee5c](https://github.com/servasec/servasec/commit/f0aee5c69c432143c107cc265e84e7528a4c77e2))


### 🤖 Continuous integration

* Fix release workflow ([0329596](https://github.com/servasec/servasec/commit/03295961004027073779f91e10778aac2c073bea))
* Fix release workflow ([804aedf](https://github.com/servasec/servasec/commit/804aedf3390c5f008bbc1495d4af28b651853b1b))
* Fix semver release ([d6294ed](https://github.com/servasec/servasec/commit/d6294ed8dabf8d2b4cec8133bb6b9588573530bd))
* Release please custom config ([646e9bc](https://github.com/servasec/servasec/commit/646e9bcb6428f0ff0485b59074fb587f7b90f93d))
* **release:** Fix custom note ([f3fac0d](https://github.com/servasec/servasec/commit/f3fac0d53369d3ba07dba11899034812ad1b212c))
* SCA+SAST workflow ([4184552](https://github.com/servasec/servasec/commit/41845524312a8e44514e759e4652e28178c4b17e))
* Use release please ([f5bbeb6](https://github.com/servasec/servasec/commit/f5bbeb6cc1bf57a40e50adad38d597c1014f7a24))
