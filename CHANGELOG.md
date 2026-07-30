# Changelog

## 2.1.0
- Migrated Go backend to Go 1.26.5
- Upgraded to Grafana 13.1.0
- Upgraded the plugin to use Akamai Edgegrid v13.3.0
- Upgraded the Node version to v24
- Upgraded React from 18.2.0 to 18.3.1 (remains on React 18, as required by Grafana 13)
- Upgraded TypeScript to 6.0.3
- Upgraded mage from 1.15.0 to 1.17.2
- Upgraded Grafana Plugin SDK (Go backend) from v0.290.0 to v0.293.0
- Upgraded Yarn from 1.22.22 to 4.17.1 (Yarn Modern)
- Upgraded frontend build and test tooling to latest (Jest 30, webpack 5.108, Prettier 3.9, Playwright, SWC)
- Updated project license file

## 2.0.1
- Upgraded to Grafana 12.4.0
- Added notice regarding plugin visibility in the Grafana Catalog for unsigned plugins
- Upgraded the plugin to use Akamai Edgegrid v13

## 2.0.0 

- Upgraded to Grafana 12.3.1
- Migrated build system from deprecated @grafana/toolkit to @grafana/create-plugin
- Upgraded the plugin to use Akamai Edgegrid v12.3.0
- Added native support for Apple Silicon