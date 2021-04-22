# Build Instructions
Build, package and release the "Global Traffic Management (GTM) Datasource" plugin.

## Clone or fork the repository
See [Git Handbook](https://guides.github.com/introduction/git-handbook/) for instructions.  
If you clone, all changes should be made on the 'develop' branch.

## Update the version number
Edit package.json

Advance the version number.  For example:
```
  "version": "2.0.0",
```

## Build
See these references:  
* [Build a plugin](https://grafana.com/docs/grafana/latest/developers/plugins/)
* [Build a data source plugin](https://grafana.com/tutorials/build-a-data-source-plugin/)

### First time build
Run this command:
```
yarn install --pure-lockfile
```

### Build the back end
Run this command:
```
mage -v
```

My output (after having previously built), looks like this:
```
$ mage -v
Running dependency: github.com/grafana/grafana-plugin-sdk-go/build.Build.LinuxARM-fm
Running dependency: github.com/grafana/grafana-plugin-sdk-go/build.Build.Linux-fm
Running dependency: github.com/grafana/grafana-plugin-sdk-go/build.Build.Darwin-fm
Running dependency: github.com/grafana/grafana-plugin-sdk-go/build.Build.LinuxARM64-fm
Running dependency: github.com/grafana/grafana-plugin-sdk-go/build.Build.Windows-fm
exec: go build -o dist/gpx_akamai-gtm-datasource-plugin_linux_arm -ldflags -w -s -extldflags "-static" ./pkg
exec: go build -o dist/gpx_akamai-gtm-datasource-plugin_windows_amd64.exe -ldflags -w -s -extldflags "-static" ./pkg
exec: go build -o dist/gpx_akamai-gtm-datasource-plugin_darwin_amd64 -ldflags -w -s -extldflags "-static" ./pkg
exec: go build -o dist/gpx_akamai-gtm-datasource-plugin_linux_amd64 -ldflags -w -s -extldflags "-static" ./pkg
exec: go build -o dist/gpx_akamai-gtm-datasource-plugin_linux_arm64 -ldflags -w -s -extldflags "-static" ./pkg
```

### Build the front end
Run this command:
```
yarn build
```

My output (after having previously built), looks like this:
```
$ yarn build
yarn run v1.22.10
$ grafana-toolkit plugin:build
✔ Preparing
✔ Linting
No tests found, exiting with code 0
✔ Running tests
⠙ Compiling...  Starting type checking service...
  Using 1 worker with 2048MB memory limit
⠙ Compiling...  
   Hash: ff7600d671e6bd368aad
  Version: webpack 4.41.5
  Time: 5879ms
  Built at: 03/25/2021 3:54:15 PM
                  Asset       Size  Chunks                   Chunk Names
           CHANGELOG.md    0 bytes          [emitted]        
                LICENSE   11.1 KiB          [emitted]        
              README.md   5.02 KiB          [emitted]        
    img/akamai-logo.png   1.72 KiB          [emitted]        
           img/logo.svg   1.55 KiB          [emitted]        
              module.js     12 KiB       0  [emitted]        module
  module.js.LICENSE.txt  808 bytes          [emitted]        
          module.js.map   75.4 KiB       0  [emitted] [dev]  module
            plugin.json   1.02 KiB          [emitted]        
  Entrypoint module = module.js module.js.map
   [0] external "react" 42 bytes {0} [built]
   [1] ../node_modules/lodash/isObject.js 733 bytes {0} [built]
   [2] external "@grafana/ui" 42 bytes {0} [built]
   [6] ../node_modules/lodash/identity.js 370 bytes {0} [built]
  [10] ../node_modules/lodash/eq.js 799 bytes {0} [built]
  [11] ../node_modules/lodash/isArrayLike.js 830 bytes {0} [built]
  [13] ../node_modules/lodash/_isIndex.js 759 bytes {0} [built]
  [15] external "@grafana/data" 42 bytes {0} [built]
  [16] external "@grafana/runtime" 42 bytes {0} [built]
  [17] ../node_modules/lodash/defaults.js 1.71 KiB {0} [built]
  [18] ../node_modules/lodash/_baseRest.js 559 bytes {0} [built]
  [19] ../node_modules/lodash/_overRest.js 1.07 KiB {0} [built]
  [35] ../node_modules/lodash/_isIterateeCall.js 877 bytes {0} [built]
  [36] ../node_modules/lodash/keysIn.js 778 bytes {0} [built]
  [51] ./module.ts + 5 modules 19.3 KiB {0} [built]
       | ./module.ts 905 bytes [built]
       | ./DataSource.ts 961 bytes [built]
       | ./ConfigEditor.tsx 4.07 KiB [built]
       | ./QueryEditor.tsx 2.68 KiB [built]
       | ../node_modules/tslib/tslib.es6.js 10 KiB [built]
       | ./types.ts 638 bytes [built]
      + 37 hidden modules 
  
✔ Compiling...
✨  Done in 15.14s.
```

## Commit your changes 
See [Git Handbook](https://guides.github.com/introduction/git-handbook/) for instructions.  
Open a Pull Request.

## Package
Copy the 'dist' directory to 'akamai-gtm-datasource' and then compress.
```
cp -r dist akamai-gtm-datasource
zip akamai-gtm-datasource-2.0.0.zip akamai-gtm-datasource/ -r
```
'2.0.0' is an example. Use your current plugin version number.

## Release
Navigate to https://github.com/akamai/gtm-grafana-datasource-plugin.
Log in. (You'll need admin rights.)

Follow the directions in [Managing releases in a repository](https://docs.github.com/en/github/administering-a-repository/managing-releases-in-a-repository).  
Tags should start with 'v', followed by the build number.  For example, 'v2.0.0'.  

