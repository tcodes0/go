# cmd: v0.1.5 *(2024-08-05)*

### [Diff with cmd/v0.1.4](https://github.com/tcodes0/go/compare/cmd/v0.1.4..cmd/v0.1.5)

## Features
- **ci**: handle workflow_dispatch events, improve ci flags ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- **filer**: remove action flag, pass action as config head line ([c82a6681](https://github.com/tcodes0/go/commit/c82a6681099dcf852d02e8ca774e14a4b639dc5f))
- **scripts**: include in setup checks for wiki clone locally and .env copy ([c82a6681](https://github.com/tcodes0/go/commit/c82a6681099dcf852d02e8ca774e14a4b639dc5f))

## Bug Fixes
- **ci**: run spellcheck when .md files change ([c82a6681](https://github.com/tcodes0/go/commit/c82a6681099dcf852d02e8ca774e14a4b639dc5f))
- **cmd/changelog**: correct diff link ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))
- **misc**: do not override env vars with dot env file ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))
- **spellcheck**: ignore go.sum files ([c82a6681](https://github.com/tcodes0/go/commit/c82a6681099dcf852d02e8ca774e14a4b639dc5f))
- **workflows**: remove step skips based on go.mod issues, move files.yml ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))
- **workflows/release-pr**: fetch tags and all history ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- update all go.mods ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))

## Improvements
- **ci**: improve shared lib functions ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- **cmd/changelog**: improve readability ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))
- **cmd/changelog**: parse merge commit bodies ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))
- **cmd/changelog**: remove unnecessary logic since always run on main ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))

## Documentation
- **ci**: improve usage message ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- **runner**: improve comments ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))
- **runner**: remove incorrect all module from usage ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))

## Tests
- **cmd/changelog**: update tests ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))

#### Other
- **cmd/runner**: rename lint commands ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))
- **workflows/release**: improve script ([5974cb8f](https://github.com/tcodes0/go/commit/5974cb8f96fb6da96a5b917c5f43203daee1b431))
- testing ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- debug ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- debug ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- correct case ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- code review ([f9063a76](https://github.com/tcodes0/go/commit/f9063a7680f3b3342b70d2828e64894d3d34e507))
- accept cmd as a valid module ([7ba238f8](https://github.com/tcodes0/go/commit/7ba238f8d6cd68e5a42b39335d4386d4126f31cc))# all: v0.1.4 _(2024-08-02)_

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