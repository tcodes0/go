# cmd/v0.2.0 *(2024-09-06)*
### [Diff with cmd/v0.1.5](https://github.com/tcodes0/go/compare/cmd/v0.2.0..cmd/v0.1.5)
# logging/v0.2.0 *(2024-09-06)*
### [Diff with logging/v0.1.5](https://github.com/tcodes0/go/compare/logging/v0.2.0..logging/v0.1.5)

### PRs in this release: [#49](https://github.com/tcodes0/go/pull/49), [#50](https://github.com/tcodes0/go/pull/50), [#51](https://github.com/tcodes0/go/pull/51), [#52](https://github.com/tcodes0/go/pull/52), [#53](https://github.com/tcodes0/go/pull/53), [#54](https://github.com/tcodes0/go/pull/54), [#55](https://github.com/tcodes0/go/pull/55), [#56](https://github.com/tcodes0/go/pull/56), [#57](https://github.com/tcodes0/go/pull/57), [#58](https://github.com/tcodes0/go/pull/58), [#60](https://github.com/tcodes0/go/pull/60)
## *Breaking Changes*
- **logging**: add new level param to logger.Stacktrace ([25a207c9](https://github.com/tcodes0/go/commit/25a207c991fab87ed0de0d4b64af2ddc03958901))
- **cmd/changelog**: introduce new flags -url -title and -tagprefix to be useful in more repos ([c18f0ed0](https://github.com/tcodes0/go/commit/c18f0ed0386ff191d95449a7361add3a990a3b4f))
- **cmd/changelog**: release multiple tags, rename -tagprefix to -tagprefixes ([a499c79e](https://github.com/tcodes0/go/commit/a499c79eb50347c5840fcdaf9a0421c6bc382f66))

## Features
- **cmd**: update template with better recover function and -version ([f1faeb39](https://github.com/tcodes0/go/commit/f1faeb399e27c857b56d0384429aa01a2da9b19e))
- **cmd/changelog**: add -version flag, improve recover function ([c36ecd64](https://github.com/tcodes0/go/commit/c36ecd645edab844d3904cbcbbace03ec77a7da4))
- **cmd/changelog**: add h3 with all PRs in the release ([4f94692c](https://github.com/tcodes0/go/commit/4f94692ca904b914aa249d85ec56d0d9eb394061))
- **cmd/changelog**: fetch commit hashes from github ([f1cb948d](https://github.com/tcodes0/go/commit/f1cb948d6a10c9bf76368b5fb286c94952321cc5))
- **cmd/changelog**: remove repetitive commits ([66973b16](https://github.com/tcodes0/go/commit/66973b16c340a517e9c8009e0d18a7cf436f8512))
- **cmd/changelog**: write new tags to a file if flag tagsfile passed ([04b585e1](https://github.com/tcodes0/go/commit/04b585e17e177f4fa1fc966ca0095633191ae06f))
- **cmd/copyright**: add -version flag, improve recover function ([85ac507e](https://github.com/tcodes0/go/commit/85ac507e6aa0c8828227d7df54973e768e20ddc6))
- **cmd/filer**: add -version flag, improve recover function, fix double flag usage message ([acdc0399](https://github.com/tcodes0/go/commit/acdc039948427aedf511b93c92658e68abd924c8))
- **cmd/runner**: add -version flag, improve recover function ([88cb7668](https://github.com/tcodes0/go/commit/88cb7668cca1ac58b0806e92317ccc0cf2fd129f))
- **go**: update to 1.23 ([410acacf](https://github.com/tcodes0/go/commit/410acacf962eb9ba16f5bc143112bcb5c96ba002))
- **workflows**: update release automation ([11d7a1d7](https://github.com/tcodes0/go/commit/11d7a1d7a4723442d0efaf2466b40fc0ed0027da))
- **workflows/release**: bump cmd configs version ([1bcf3b65](https://github.com/tcodes0/go/commit/1bcf3b650948d2d26678ec10f5b13cc780595308))

## Bug Fixes
- **cmd/changelog**: collect breaking changes on body header ([ae63ffd7](https://github.com/tcodes0/go/commit/ae63ffd76c499258ac39ee3597629f4bc81392f6))
- **cmd/changelog**: remove newline ([1e8d4474](https://github.com/tcodes0/go/commit/1e8d447438e8f20649a38ef457aa3b26d1e72611))
- **cmd/changelog**: resolve race conditions ([8a3a8017](https://github.com/tcodes0/go/commit/8a3a8017ab16b80ad5a9723fcdea19a0ab0d83fc))
- **cmd/runner**: correct index passed to task.Execute ([cdb1a67f](https://github.com/tcodes0/go/commit/cdb1a67f3125bf65aefd1bfe34f9a9b960015f90))
- **configs**: correct missing extends in commitlint ([c047de97](https://github.com/tcodes0/go/commit/c047de97a32e3a3bc3b20b75f0fc1909a3b27500))
- **scripts**: correct undefined var in new module ([cf20a3a9](https://github.com/tcodes0/go/commit/cf20a3a98fda82f7368f14a8306adc48dd10c2c0))
- **scripts**: handle some edge case errors on mock generation and git log parsing ([05766457](https://github.com/tcodes0/go/commit/05766457b2d212d00025a59e493969a436b9091a))
- **vscode**: silence custom context workflow variables warnings ([b4c6e9cb](https://github.com/tcodes0/go/commit/b4c6e9cb1829889660710cfc2bd2cfe942c2b0ea))
- **workflows**: change prettier action ([134312c4](https://github.com/tcodes0/go/commit/134312c4db134af7257c99c1d715b4bd2ba38575))
- **workflows**: checkout submodule with ssh-key ([9db969be](https://github.com/tcodes0/go/commit/9db969be19d8d311f4b84961a80116c8fb486380))
- **workflows**: correct BASH_ENV path ([abc3d2f4](https://github.com/tcodes0/go/commit/abc3d2f4a2ddd9bb308d079728bf2667334c705a))
- **workflows**: release-pr pass locally ([d9e7b28b](https://github.com/tcodes0/go/commit/d9e7b28b131e989f563ec8d7f3dc63ccdf62668d))
- **workflows**: update release automation ([337ecf4f](https://github.com/tcodes0/go/commit/337ecf4f491bb949215333616cd84e376b680cde))
- **workflows/release**: pass envs from workflows ([c473e93f](https://github.com/tcodes0/go/commit/c473e93f7d6b9b3001098a9ec0ad55f698eeaa46))
- correct run wrapper not passing args to ci ([e994f3da](https://github.com/tcodes0/go/commit/e994f3da9c7e2eaceb9d314a49699de265ca6801))

## Performance
- **cmd/changelog**: query in parallel ([6ea70c61](https://github.com/tcodes0/go/commit/6ea70c6166359a1e309aa79743e87bda7b62dff7))

## Improvements
- **cmd/changelog**: remove title validation ([1db60c34](https://github.com/tcodes0/go/commit/1db60c343cc6f5c67c093775667b307735f61c89))
- **cmd/changelog**: update title generation ([ae522260](https://github.com/tcodes0/go/commit/ae52226089c61a1ac62437f56cf9c25d6d0ac090))
- **cmd/filer**: improve output when no changes occur ([e7baa159](https://github.com/tcodes0/go/commit/e7baa159fa922c0590b4539dca6113c67402b8b7))
- **cmd/filer**: simplify filer func using action funcs that process slices ([ec5f6047](https://github.com/tcodes0/go/commit/ec5f60479d709af9ae83fcd9d377d2d13ffe2a47))
- **cmd/filer**: some warns to debug and errors, collect all errors in one run ([8b25d524](https://github.com/tcodes0/go/commit/8b25d524f9f1535979320cc400adead8911c01fa))
- **cmd/runner**: simplify main, fix space and quotes issues in config ([c609e245](https://github.com/tcodes0/go/commit/c609e2456fdd8dfbc049a6bac295f09493172780))
- **lib.sh**: use BASH_ENV to source lib.sh remove hardcoded source ([62fec87b](https://github.com/tcodes0/go/commit/62fec87b28ef4867dbf47bea327ce504a8e9e243))
- **workflows**: organize script module_pr/test_pretty ([dafe3726](https://github.com/tcodes0/go/commit/dafe3726e50294ff0e3e12b4045631ea3c9caa0b))
- **repo**: replace run and ci links with a wrapper script ([f2bdb62e](https://github.com/tcodes0/go/commit/f2bdb62ec412dceb41536b57d3d8ff386441040c))
- **scripts**: refactor _sed to SED variable ([c14f83c8](https://github.com/tcodes0/go/commit/c14f83c86857f974b41a731ce20384e008a063d4))
- **scripts**: reference sh/lib for several scripts ([6f7a3477](https://github.com/tcodes0/go/commit/6f7a3477fb7b7ccdd7a58489b1f542a905bd12b8))
- **scripts/tag**: organize script into functions ([45f27d01](https://github.com/tcodes0/go/commit/45f27d0171690ad904b93704d754e640da19d665))
- **scripts/tag**: parse changelog and push tags for latest release ([6a2c0aa9](https://github.com/tcodes0/go/commit/6a2c0aa991a58a6516023c2b155792209ccae169))
- **setup.sh**: improve root validation, remove wiki ([b76217fe](https://github.com/tcodes0/go/commit/b76217fe60683d7fc10f7e0d01dbe00b24f309d0))
- **sh/new_module**: improve script organization ([490c1824](https://github.com/tcodes0/go/commit/490c1824c47b6616e6bbc3dbff5dadbedd4c1955))

## Documentation
- **cmd**: use printf to print help messages ([28556c28](https://github.com/tcodes0/go/commit/28556c285e4c1ec2f51514ea9eca5f9dca02aa7f))
- **cmd/changelog**: remove some references to module ([fecf61a3](https://github.com/tcodes0/go/commit/fecf61a3a8e969963eed0d580089778254d2772f))
- **regex**: document complex regexes with links ([90b30e2a](https://github.com/tcodes0/go/commit/90b30e2af874bc5b7b6b0c7a5e1a1d31653b9886))
- **scripts**: document single iteration loop ([07fdc9e4](https://github.com/tcodes0/go/commit/07fdc9e46968316dc0959481116ebd7abed000a9))
- remove wiki from readme ([abe53eb4](https://github.com/tcodes0/go/commit/abe53eb4579aa6235b0be4f017aa74eb1a332b9d))

## Styling
- **cmd/filer**: improve output ([39c37c69](https://github.com/tcodes0/go/commit/39c37c69beba0d1359d030fad46a1ef0a0753a8c))

#### Other
- **cmd/changelog**: list prs in ascending order ([0bdccb46](https://github.com/tcodes0/go/commit/0bdccb464c39744c1fc4c43f5d0420e89ef2eadb))
- **cmd/changelog**: warn instead of error ([08a259fa](https://github.com/tcodes0/go/commit/08a259fae2e223920da0575aa8fd28bf768563b5))
- **repo**: add sh submodule ([36927122](https://github.com/tcodes0/go/commit/3692712208ad0c0196f9b8446a7d59c01d0ba155))
- **scripts**: use a wrapper script with bash_env ([c36f7f85](https://github.com/tcodes0/go/commit/c36f7f856aa1ed0b773d2a0de85016fa888efc2c))
- **workflows**: try to not use a ssh key ([93332d0a](https://github.com/tcodes0/go/commit/93332d0ada818c05083c2de71ada6a43d2958e94))
- fix ci ([1a608ca7](https://github.com/tcodes0/go/commit/1a608ca7ccf42a8b02a0a889d880e8181d9b1360))
- format configs ([1700b324](https://github.com/tcodes0/go/commit/1700b3243be384197c1eb29d45fdf9fdd5095940))
- bump sh/lib to v0.2.4 ([c91aa103](https://github.com/tcodes0/go/commit/c91aa1031d8d2e7a5db54d2bbaee8321948547f0))
- bump sh/lib to v0.2.5 ([f0984d69](https://github.com/tcodes0/go/commit/f0984d6918766e52897470dfeac88a2fcd4f9424))
- renames ([87e449d0](https://github.com/tcodes0/go/commit/87e449d07eee51b1929e0fc1b5fe3f72755de3d3))
- remove embed import ([03d1b2ba](https://github.com/tcodes0/go/commit/03d1b2ba483ae2bb9ea09bceab98a72c798bfa55))
- update setup.sh to include submodule init ([181d18a7](https://github.com/tcodes0/go/commit/181d18a715850539f1b2e2f9de2e1c476715c83f))
- rename some dirs to snake case ([e21edfac](https://github.com/tcodes0/go/commit/e21edfac70c62d5c0d84cdabe1b5b0a9dfabf641))
- symlink lib.sh ([8be6a7c7](https://github.com/tcodes0/go/commit/8be6a7c7131fdcf814279bdc3e648d88fa23700d))
- update env_default ([d5684893](https://github.com/tcodes0/go/commit/d5684893cf7472643baac851063716365d7d6bef))
- runner related changes and misc ([4b604b9b](https://github.com/tcodes0/go/commit/4b604b9b44026775582e5b0936db71fd55ea5461))
- bump sh/lib ([65489f99](https://github.com/tcodes0/go/commit/65489f99a1a7d9c11a0930b9fff639e3d156008c))
- fix workflows ([f8ceda74](https://github.com/tcodes0/go/commit/f8ceda747f99a3340bf02a6cf6a389cbc813c679))
- update env default ([e4cc378f](https://github.com/tcodes0/go/commit/e4cc378ff83558703023e39a797efc448b260d98))
- update workflow input description ([4d803452](https://github.com/tcodes0/go/commit/4d803452c47cb915fc9ddc41e389eb14e97b734c))
- remove BASH_ENV from env-default ([e7024068](https://github.com/tcodes0/go/commit/e7024068239b6cf3ad897f0752e26fbaf2f812dd))
- lintfix ([8983e0f2](https://github.com/tcodes0/go/commit/8983e0f20074cd31173c313f86ff7390ff392196))
- code review ([fc4b3754](https://github.com/tcodes0/go/commit/fc4b3754b7198345ca756acaea6b77fa8195d4bd))

# httpmisc: v0.1.5 *(2024-08-05)*
### [Diff with httpmisc/v0.1.4](https://github.com/tcodes0/go/compare/httpmisc/v0.1.4..httpmisc/v0.1.5)
## Bug Fixes
- update several go.mods ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))

# jsonutil: v0.1.5 *(2024-08-05)*
### [Diff with jsonutil/v0.1.4](https://github.com/tcodes0/go/compare/jsonutil/v0.1.4..jsonutil/v0.1.5)
## Bug Fixes
- update several go.mods ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))

# logging: v0.1.5 *(2024-08-05)*
### [Diff with logging/v0.1.4](https://github.com/tcodes0/go/compare/logging/v0.1.4..logging/v0.1.5)
## Bug Fixes
- update several go.mods ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))

# misc: v0.1.5 *(2024-08-05)*
### [Diff with misc/v0.1.4](https://github.com/tcodes0/go/compare/misc/v0.1.4..misc/v0.1.5)
## Bug Fixes
- update several go.mods ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))
- **misc**: do not override env vars with dot env file ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))

# cmd: v0.1.5 *(2024-08-05)*

### [Diff with cmd/v0.1.4](https://github.com/tcodes0/go/compare/cmd/v0.1.4..cmd/v0.1.5)

## Features
- **scripts/ci**: handle workflow_dispatch events, improve ci flags ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- **cmd/filer**: remove action flag, pass action as config head line ([c82a6681](https://github.com/tcodes0/go/commit/c82a6681099dcf852d02e8ca774e14a4b639dc5f))
- **scripts**: include in setup checks for wiki clone locally and .env copy ([c82a6681](https://github.com/tcodes0/go/commit/c82a6681099dcf852d02e8ca774e14a4b639dc5f))

## Bug Fixes
- **scripts/ci**: run spellcheck when .md files change ([c82a6681](https://github.com/tcodes0/go/commit/c82a6681099dcf852d02e8ca774e14a4b639dc5f))
- **cmd/changelog**: correct diff link ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))
- **spellcheck**: ignore go.sum files ([c82a6681](https://github.com/tcodes0/go/commit/c82a6681099dcf852d02e8ca774e14a4b639dc5f))
- **workflows**: remove step skips based on go.mod issues, move files.yml ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))
- **workflows/release-pr**: fetch tags and all history ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- update several go.mods ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))

## Improvements
- **scripts/ci**: improve shared lib functions ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- **cmd/changelog**: improve readability ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))
- **cmd/changelog**: parse merge commit bodies ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))
- **cmd/changelog**: remove unnecessary logic since always run on main ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))

## Documentation
- **scripts/ci**: improve usage message ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- **cmd/runner**: improve comments ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))
- **cmd/runner**: remove incorrect all module from usage ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))

## Tests
- **cmd/changelog**: update tests ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))

#### Other
- **cmd/runner**: rename lint commands ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))
- **workflows/release**: improve script ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))
- accept cmd as a valid module ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))# all: v0.1.4 _(2024-08-02)_

# all: v0.1.4 _(2024-08-02)_

### No diff, this is the first release!

## Modules released

- clock _v0.1.4_
- httpmisc _v0.1.4_
- hue _v0.1.4_
- identifier _v0.1.4_
- jsonutil _v0.1.4_
- logging _v0.1.4_
- misc _v0.1.4_
- cmd _v0.1.4_
  - changelog
  - copyright
  - filer
  - gengowork
  - runner

Since the initial commit modules have been developed together with internal scripts and tools.

_v0.1.4_ assumes semver moving forward. Consider all commits before breaking changes.