# Akamai Global Traffic Management (GTM) Datasource
v1.0.0

Use the Akamai Global Traffic Management plugin to observe GTM  metrics.

## Prerequisites

* Grafana 7.0 or newer.
* An Akamai API client with authorization to use the [Load Balancing DNS Traffic All Properties API](https://developer.akamai.com/api/core_features/reporting/load-balancing-dns-traffic-all-properties.html). 
* [Authenticate With EdgeGrid](https://developer.akamai.com/getting-started/edgegrid) provides information to generate the required credentials. Choose the API service named "Reporting API", and set the access level to READ-WRITE.

## Installing Grafana

![Install Grafana](https://grafana.com/docs/grafana/latest/installation/) details the process of installing Grafana on several operating systems.

## Installing this plugin on a local Grafana

* On the ![gtm-grafana-datasource-plugin](https://github.com/akamai/gtm-grafana-datasource-plugin) GitHub repository, 
under "Releases", select "Grafana datasource for Akamai Global Traffic Management (GTM)  metrics v1.0.0".

* Copy "akamai-gtm-datasource-1.0.0.zip" to your computer.  Unzip the archive.

### Linux OSs (Debian, Ubuntu, CentOS, Fedora, OpenSuse, Red Hat)

Configuration file: /etc/grafana/grafana.ini
Plugin directory: /var/lib/grafana/plugins

From the unzipped archive, copy one of (as appropriate for your hardware):
* gpx_akamai-gtm-datasource-plugin_linux_amd64
* gpx_akamai-gtm-datasource-plugin_linux_arm
* gpx_akamai-gtm-datasource-plugin_linux_arm64
to /var/lib/grafana/plugins

### Macintosh

Configuration file: /usr/local/etc/grafana/grafana.ini
Plugin directory: /usr/local/var/lib/grafana/plugins

From the unzipped archive, copy:
* gpx_akamai-gtm-datasource-plugin_darwin_amd64
to /var/lib/grafana/plugins

### Windows

Grafana can be installed into any directory (install_dir).
Configuration file: install_dir\conf
Plugin directory: install_dir\data\plugins

From the unzipped archive, copy:
* gpx_akamai-gtm-datasource-plugin_windows_amd64.exe
to install_dir\data\plugins

### Grafana configuration

![Configuration](https://grafana.com/docs/grafana/latest/administration/configuration/) 
describes configuration for each operating system.

* Using a text editor, open the configuration file (as described in ![Configuration](https://grafana.com/docs/grafana/latest/administration/configuration/).
* Under the [paths] section, uncomment "plugins".
* To the right of "plugins =", insert the complete path to the plugin directory.
* Under the [plugins] section, uncomment "allow_loading_unsigned_plugins".
* To the right of "allow_loading_unsigned_plugins =", add "akamai-gtm-datasource" (without quotes).

### Restart Grafana
![Restart Grafana](https://grafana.com/docs/grafana/latest/installation/restart-grafana/)
describes how to restart Grafana for each operating system.

## "Akamai Edge DNS Datasource" Configuration

In the datasource configuration panel, enter your Akamai credentials.

![Data Source](https://github.com/akamai/gtm-grafana-datasource-plugin/blob/develop/static/data-source-config.png)

Create a new dashboard and add a panel.

Enter one domain name.

![Domain](https://github.com/akamai/gtm-grafana-datasource-plugin/blob/develop/static/domains-config.png)

Metric name is optional. If empty then a metric name is automatically generated.

![Metric Name](https://github.com/akamai/gtm-grafana-datasource-plugin/blob/develop/static/metric-name-config.png)

