## [2.8.2](https://github.com/AtomiCloud/sulfone.boron/compare/v2.8.1...v2.8.2) (2026-03-17)


### 🐛 Bug Fixes 🐛

* **executor:** unzip container name collision in parallel sessions ([#30](https://github.com/AtomiCloud/sulfone.boron/issues/30)) ([7659981](https://github.com/AtomiCloud/sulfone.boron/commit/765998112c0fa4d2a5229295c7622bbe72313ebb))
* **executor:** unzip container name collision in parallel sessions ([6ec8f37](https://github.com/AtomiCloud/sulfone.boron/commit/6ec8f37adfa8c465c01f2fb160019dc67b178af7))

## [2.8.1](https://github.com/AtomiCloud/sulfone.boron/compare/v2.8.0...v2.8.1) (2026-03-13)


### 🐛 Bug Fixes 🐛

* **executor:** prevent nil dereference in PullImages ([2eff2c2](https://github.com/AtomiCloud/sulfone.boron/commit/2eff2c2d0e8d260f11ad017aa9b3a7d521f7bb9c))

## [2.8.0](https://github.com/AtomiCloud/sulfone.boron/compare/v2.7.0...v2.8.0) (2026-03-12)


### ✨ Features ✨

* **executor:** add POST /executor/try endpoint for local testing ([10a00c2](https://github.com/AtomiCloud/sulfone.boron/commit/10a00c21c40f3d8045d414190bd01d7067719767))


### 🐛 Bug Fixes 🐛

* address coderabbit local review findings ([39e016f](https://github.com/AtomiCloud/sulfone.boron/commit/39e016f1edfddc400fc2f5da60fd11cac286a06e))
* **executor:** address CodeRabbit review comments ([2506063](https://github.com/AtomiCloud/sulfone.boron/commit/25060639a9c2bf19266f90a53f387f358885d825))
* **executor:** fix try endpoint and cleanup bugs ([5fc5719](https://github.com/AtomiCloud/sulfone.boron/commit/5fc5719acda703b59b16ff45ffcf8db88e66ee98))
* **executor:** health check already-running resolver containers ([b1fb574](https://github.com/AtomiCloud/sulfone.boron/commit/b1fb574ae559e5a007bc7a9e98472c8aeacde6bf))
* **server:** persist default source value to request struct ([8ea3b43](https://github.com/AtomiCloud/sulfone.boron/commit/8ea3b43b3cf721e0564095caa196c3956ee2c6ac))
* **executor:** preserve positional error slices, fix Clean() nil filter ([4329fea](https://github.com/AtomiCloud/sulfone.boron/commit/4329fea2895e4989bc0cddd2b713828ed4ffdda3))
* **server:** return realPath for consistency in validatePath ([e2cddd7](https://github.com/AtomiCloud/sulfone.boron/commit/e2cddd7ff497c5610dfd048038b0b6a9f5c5d83e))
* **executor:** use defer for unzip container cleanup ([6cb7d8d](https://github.com/AtomiCloud/sulfone.boron/commit/6cb7d8dc0517444e0d2d8e8e3bfddf704ab55e75))
* **executor:** use health check for blob extraction ([c553790](https://github.com/AtomiCloud/sulfone.boron/commit/c5537901249619bd2f4fc0f1c0a18b8f21d3e2c9))
* **test:** use t.Fatalf to prevent nil dereference in symlink test ([a886baf](https://github.com/AtomiCloud/sulfone.boron/commit/a886baf2ed28cdfb58c16e7cf5dd61904478db90))

## [2.7.0](https://github.com/AtomiCloud/sulfone.boron/compare/v2.6.0...v2.7.0) (2026-03-11)


### 📜 Documentation 📜

* add implementation plans for CU-86ewueggx ([3b25c87](https://github.com/AtomiCloud/sulfone.boron/commit/3b25c872a1eb421162f0fa8d846ac8fba9032707))
* add task spec for CU-86ewueggx ([d67e93c](https://github.com/AtomiCloud/sulfone.boron/commit/d67e93c108084b20d6c531b9640bbf7366948c77))
* update task spec for CU-86ewueggx - use doublestar glob library ([5b3e39e](https://github.com/AtomiCloud/sulfone.boron/commit/5b3e39e943a8fb9efba92854cf1517483e63b757))


### ✨ Features ✨

* **merger:** implement resolver support for conflict resolution ([e6c37d5](https://github.com/AtomiCloud/sulfone.boron/commit/e6c37d59869cb1e67de43f028e1b8f8a7bc88293))


### 🐛 Bug Fixes 🐛

* add defensive nil check for req.Files in callResolver error path ([4ce33c2](https://github.com/AtomiCloud/sulfone.boron/commit/4ce33c2579bd1574501dbc0b024a46b28c017bd3))
* **merger:** address CodeRabbit review feedback for resolver integration ([81248e1](https://github.com/AtomiCloud/sulfone.boron/commit/81248e17e858f271839a23388977ccafa87a9275))
* **merger:** address CodeRabbit review feedback for resolver integration ([b896ec9](https://github.com/AtomiCloud/sulfone.boron/commit/b896ec960218df9448c42dc523f301f5dad3963f))
* **merger:** use os.Chmod for file mode preservation [CU-86ewueggx] ([9cab6d7](https://github.com/AtomiCloud/sulfone.boron/commit/9cab6d7dd03857141b708c5d95e4b97e6684b1eb))

## [2.6.0](https://github.com/AtomiCloud/sulfone.boron/compare/v2.5.0...v2.6.0) (2026-03-09)


### 📜 Documentation 📜

* add task spec and implementation plan for CU-86ewrbr3t ([9b73155](https://github.com/AtomiCloud/sulfone.boron/commit/9b73155eea7c6f2a48e45ba976298b10b5930897))
* **spec:** add v1 feedback for resolver coordinator ([136b882](https://github.com/AtomiCloud/sulfone.boron/commit/136b882c8eca280e4588714a167c5c051affdc72))
* add v2 implementation plans for CU-86ewrbr3t ([896c2bd](https://github.com/AtomiCloud/sulfone.boron/commit/896c2bd378e6fbda3ea7572e40ed8cd13bebe33b))


### ✨ Features ✨

* **resolver:** implement resolver coordinator for warming and proxying ([9b71d5a](https://github.com/AtomiCloud/sulfone.boron/commit/9b71d5a320ad38bda220e7e45e1c246e41991ca0))


### 🐛 Bug Fixes 🐛

* address coderabbit local review findings ([78ce85b](https://github.com/AtomiCloud/sulfone.boron/commit/78ce85bdb5caeae15135970772a09bfbcd47ea7f))
* address CodeRabbit local review findings ([a71992a](https://github.com/AtomiCloud/sulfone.boron/commit/a71992a70d21eded9548f227a44e2760233689e5))
* **resolver:** address CodeRabbit review feedback ([ed29940](https://github.com/AtomiCloud/sulfone.boron/commit/ed299405201e4e7f489c6f45e283adf73b34ecb5))
* **resolver:** de-duplicate resolvers before warming ([6820579](https://github.com/AtomiCloud/sulfone.boron/commit/6820579af994b74cc8ab0f9e9be1e17597bf6690))

## [2.5.0](https://github.com/AtomiCloud/sulfone.boron/compare/v2.4.1...v2.5.0) (2026-03-09)


### ✨ Features ✨

* **cleanup:** add daemon shutdown and cleanup functionality ([acff0a7](https://github.com/AtomiCloud/sulfone.boron/commit/acff0a75e2542893b30b045427090730f92a5aea))

## [2.4.1](https://github.com/AtomiCloud/sulfone.boron/compare/v2.4.0...v2.4.1) (2026-02-26)


### 📜 Documentation 📜

* address CodeRabbit PR review feedback [CU-86ewk3qxh] ([a7badc7](https://github.com/AtomiCloud/sulfone.boron/commit/a7badc7f1e0354362a1fa75a9c67c8ba1a7c6f06))
* fix all remaining CodeRabbit feedback [CU-86ewk3qxh] ([7421aa3](https://github.com/AtomiCloud/sulfone.boron/commit/7421aa3bd038f4e4bbf4cbab159c59a36cfe52de))
* fix remaining CodeRabbit review feedback [CU-86ewk3qxh] ([1bc4b78](https://github.com/AtomiCloud/sulfone.boron/commit/1bc4b78dede5f126b850b9599a923273164710dd))


### 🐛 Bug Fixes 🐛

* address coderabbit local review findings ([672e779](https://github.com/AtomiCloud/sulfone.boron/commit/672e77941329301584e5f7b2bf7b1fdd2f1fe16d))

## [2.4.0](https://github.com/AtomiCloud/sulfone.boron/compare/v2.3.0...v2.4.0) (2026-02-23)


### 📜 Documentation 📜

* add Boron execution cluster architecture and operations documentation ([bba5cd4](https://github.com/AtomiCloud/sulfone.boron/commit/bba5cd457c7c0aae55a9d27cd991c68952c6fc3c))
* address CodeRabbit review feedback [CU-86ewk3qxh] ([240ebcd](https://github.com/AtomiCloud/sulfone.boron/commit/240ebcdc3026d63ee524ab671a7af5171ce5db64))
* fix capitalization inconsistency in table headers [CU-86ewk3qxh] ([71e8222](https://github.com/AtomiCloud/sulfone.boron/commit/71e82224bc4006e63bcbdd27aaa7282408598fe5))
* fix mermaid diagrams and improve docs [CU-86ewk3qxh] ([#23](https://github.com/AtomiCloud/sulfone.boron/issues/23)) ([67389b6](https://github.com/AtomiCloud/sulfone.boron/commit/67389b6a01a35ee8ec36ac322edbad0d549980e7))
* format documentation files [CU-86ewk3qxh] ([cc3b8aa](https://github.com/AtomiCloud/sulfone.boron/commit/cc3b8aaff50315693f422905e046b0d2b4b3f61c))


### ✨ Features ✨

* ignore kagent ([c1dd73d](https://github.com/AtomiCloud/sulfone.boron/commit/c1dd73d0fa1e4fedd060930ef9457cc4d7a98c24))

## [2.3.0](https://github.com/AtomiCloud/sulfone.boron/compare/v2.2.1...v2.3.0) (2026-01-27)


### ✨ Features ✨

* add version number configuration for boron deployment ([f996d05](https://github.com/AtomiCloud/sulfone.boron/commit/f996d0589bd2b5ccf0724e201f00e058451ba766))
* add version number configuration for boron deployment ([#22](https://github.com/AtomiCloud/sulfone.boron/issues/22)) ([005da45](https://github.com/AtomiCloud/sulfone.boron/commit/005da456531a0625a5af20a253b4a78732693faa))

## [2.2.1](https://github.com/AtomiCloud/sulfone.boron/compare/v2.2.0...v2.2.1) (2026-01-27)


### 🐛 Bug Fixes 🐛

* upgrade Docker client to v28 and fix network API types ([ce2e2cb](https://github.com/AtomiCloud/sulfone.boron/commit/ce2e2cb3e271962792d5ba57579e369f63881339))

## [2.2.0](https://github.com/AtomiCloud/sulfone.boron/compare/v2.1.1...v2.2.0) (2026-01-26)


### ✨ Features ✨

* allow key selection during deploy ([7915442](https://github.com/AtomiCloud/sulfone.boron/commit/79154421ab32dee830001fe4fc8c7c5f6bdb7bcb))


### 🐛 Bug Fixes 🐛

* **ci:** update all Docker API types for v27 ([f152d76](https://github.com/AtomiCloud/sulfone.boron/commit/f152d761fe4789425d93f97076e1269e0950eeb5))
* **ci:** update Docker API calls for v27 and fix shell script ([ff937c8](https://github.com/AtomiCloud/sulfone.boron/commit/ff937c85472ac76d4c522878347578d92f9342f0))
* **ci:** update Dockerfile to use Go 1.24 ([8fbbd88](https://github.com/AtomiCloud/sulfone.boron/commit/8fbbd886c200ce95da765e0ca2185886f96fda76))
* **default:** upgrade to use newest docker client ([ce169eb](https://github.com/AtomiCloud/sulfone.boron/commit/ce169eb1fb4d3a13f7eb2b318392c0307786aba1))

## [2.1.1](https://github.com/AtomiCloud/sulfone.boron/compare/v2.1.0...v2.1.1) (2025-06-29)


### 🐛 Bug Fixes 🐛

* **default:** incorrect reference ID for older templates ([620b607](https://github.com/AtomiCloud/sulfone.boron/commit/620b607945eef1250f2ac7a8336a88923e42da9e))
* upgraded tofu ([9113c47](https://github.com/AtomiCloud/sulfone.boron/commit/9113c473e2ded2b47a74f1150c6bd606bce636d8))

## [2.1.0](https://github.com/AtomiCloud/sulfone.boron/compare/v2.0.0...v2.1.0) (2025-05-10)


### ✨ Features ✨

* upgrade models to include template ref & empty template ([60abd10](https://github.com/AtomiCloud/sulfone.boron/commit/60abd10bb0bd98e0d145a48aa681b08231eeb15f))

## [2.0.0](https://github.com/AtomiCloud/sulfone.boron/compare/v1.2.0...v2.0.0) (2025-05-03)


### ✨ Features ✨

* add gatekeeper CI ([040d0da](https://github.com/AtomiCloud/sulfone.boron/commit/040d0da04c3da5fd5050333584af6e53d27c11fc))
* **breaking:** remove extensions ([a03194c](https://github.com/AtomiCloud/sulfone.boron/commit/a03194c8bdc87713d04965c3742eef4e2ec36739))
* upgrade infra to latest ([f6e71d8](https://github.com/AtomiCloud/sulfone.boron/commit/f6e71d84d20058203978c99f1fef92300ea7986c))


### 🐛 Bug Fixes 🐛

* incorrect docker setup version ([2dc50f4](https://github.com/AtomiCloud/sulfone.boron/commit/2dc50f4d3a30008cb8206cd8831f7bb47431e41f))
* incorrect path to docker build ([094a48c](https://github.com/AtomiCloud/sulfone.boron/commit/094a48ca9374f5efd0a2a9ef498ae0f776556128))

## [1.2.0](https://github.com/AtomiCloud/sulfone.boron/compare/v1.1.2...v1.2.0) (2025-04-23)


### ✨ Features ✨

* upgrade infra configuration ([c677999](https://github.com/AtomiCloud/sulfone.boron/commit/c677999b113ff316411f9cba0295c6773aa6f161))

## [1.1.2](https://github.com/AtomiCloud/sulfone.boron/compare/v1.1.1...v1.1.2) (2025-01-07)


### 🐛 Bug Fixes 🐛

* update image during ansible deployment ([7ca9534](https://github.com/AtomiCloud/sulfone.boron/commit/7ca953447dc1085da4b78cf3611c9eabe14f7c29))

## [1.1.1](https://github.com/AtomiCloud/sulfone.boron/compare/v1.1.0...v1.1.1) (2025-01-07)


### 🐛 Bug Fixes 🐛

* auto detect template reference instead of latest ([30d1c10](https://github.com/AtomiCloud/sulfone.boron/commit/30d1c10953642a32240741621ae75e2421f08295))

## [1.1.0](https://github.com/AtomiCloud/sulfone.boron/compare/v1.0.0...v1.1.0) (2025-01-01)


### ✨ Features ✨

* all necessary scripts for deployment ([ea48e06](https://github.com/AtomiCloud/sulfone.boron/commit/ea48e06a4e38c90794b65c229bfbcf2d1897891c))

## 1.0.0 (2025-01-01)


### ✨ Features ✨

* documentation and CI ([749eae0](https://github.com/AtomiCloud/sulfone.boron/commit/749eae0c354be11ca2829bc60fa1a0a86aac67f0))
* initial commit ([6119fe9](https://github.com/AtomiCloud/sulfone.boron/commit/6119fe94cd1c329891c82631fa8caa95e588da57))
* migrate to tag-based images ([eb7c9b6](https://github.com/AtomiCloud/sulfone.boron/commit/eb7c9b651a76bb9e8cbbf74cd1f6e8c5174c7698))
* non-transparent proxy ([cdd7a56](https://github.com/AtomiCloud/sulfone.boron/commit/cdd7a56abaf47c84a4af60977e006e66d6fc0ecf))
* upgrade to new infra v2 ([58360ae](https://github.com/AtomiCloud/sulfone.boron/commit/58360ae3ee91dd71c6f68e978d53bae069ec0165))
* use self-image for merger instead of collecting from client ([9bf822c](https://github.com/AtomiCloud/sulfone.boron/commit/9bf822cea5b1e5c5a1fa5f14a2142aeb81ab62b4))


### 🐛 Bug Fixes 🐛

* account for docker.io ([389a5f2](https://github.com/AtomiCloud/sulfone.boron/commit/389a5f2031ae2e7d9762f269018026ab871b980f))
* correct base image in Dockerfile ([33a83f1](https://github.com/AtomiCloud/sulfone.boron/commit/33a83f13b6c2f04a9779814c0cdee17e1c43353e))
* incorrect CI requiring S3 KEY ID ([a9d94f1](https://github.com/AtomiCloud/sulfone.boron/commit/a9d94f19044a05c547254e747f36b7214dad6ca6))
* incorrect Dockerfile path ([d6bb87c](https://github.com/AtomiCloud/sulfone.boron/commit/d6bb87cafe845462830782d3f7a016a12db46c1c))
* **ci:** missing github slug action ([ec0599e](https://github.com/AtomiCloud/sulfone.boron/commit/ec0599e77043d1c1224908ab7bbd6b5571648940))
* not propogating cyanglob correctly ([ca55738](https://github.com/AtomiCloud/sulfone.boron/commit/ca55738d75bf78fff3b3d5478dfafac8fdf4e224))
* releaser ([88deaa8](https://github.com/AtomiCloud/sulfone.boron/commit/88deaa82438c3e6af1aafe5fb2fec1a0b0e2918f))
* **try:** sg release pin to npm ([4798a9d](https://github.com/AtomiCloud/sulfone.boron/commit/4798a9ddee2243f90c25e101e30186ad0671e125))
* unzip not waited before removing container ([43a3ba6](https://github.com/AtomiCloud/sulfone.boron/commit/43a3ba6c21ad5babb502a2dfe87118c37634a352))
