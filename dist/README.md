# Akamai Global Traffic Management (GTM) Datasource

Use the Akamai Global Traffic Management plugin to observe GTM  metrics.

## Install Grafana 7.0 or newer

[Install Grafana](https://grafana.com/docs/grafana/latest/installation/) details the process of installing Grafana on several operating systems.

(Be sure to get version 7.0 or newer.  Your package manager may install an older version.  It's best to go to
[Install Grafana](https://grafana.com/docs/grafana/latest/installation/) and follow the directions there.)

## Obtain Akamai API credentials

"Akamai GTM Datasource" gets data from the
[Load Balancing DNS Traffic All Properties API](https://developer.akamai.com/api/core_features/reporting/load-balancing-dns-traffic-all-properties.html).

You need to create an "API Client" with authorization to use the
[Load Balancing DNS Traffic All Properties API](https://developer.akamai.com/api/core_features/reporting/load-balancing-dns-traffic-all-properties.html).

See the "Get Started" section of [Reporting API v1](https://developer.akamai.com/api/core_features/reporting/v1.html)
which says, "To enable this API, choose the API service named reporting-api, and set the access level to READ-WRITE".

Follow directions at [Authenticate With EdgeGrid](https://developer.akamai.com/getting-started/edgegrid) to generate
the required client credentials.

A customized version of those directions follows:

* Go to [Control Center](https://control.akamai.com/)
* Navigate to the "Identity & Access" page.
* Press "New API Client for Me"

![New API Client](https://github.com/akamai/gtm-grafana-datasource-plugin/blob/master/static/new-api-client.png)

* Select the "Advanced" option.
* Choose "Select APIs".
* Select "Reporting API" and "READ-WRITE" access.

![Reporting API](https://github.com/akamai/gtm-grafana-datasource-plugin/blob/master/static/reporting-api.png)

* Press "Create API client"
* Copy the credentials (client_secret, host, access_token, and client token).

![Credentials](https://github.com/akamai/gtm-grafana-datasource-plugin/blob/master/static/credential.png)

The credentials will later be entered into "Akamai GTM Datasource" configuration.

Note that Step 2 in [Authenticate With EdgeGrid](https://developer.akamai.com/getting-started/edgegrid)
"Decide which tool you’ll use to make requests" is not necessary. "Akamai GTM Datasource" makes
the requests.

## Installing this plugin on a local Grafana

* On the [gtm-grafana-datasource-plugin](https://github.com/akamai/gtm-grafana-datasource-plugin) GitHub repository, 
under "Releases", select "Grafana datasource for Akamai Global Traffic Management (GTM)  metrics v1.0.1".

* Copy "akamai-gtm-datasource-1.0.1.zip" to your computer.  Unzip the archive.

### Linux OSs (Debian, Ubuntu, CentOS, Fedora, OpenSuse, Red Hat)

Configuration file: /etc/grafana/grafana.ini  
Plugin directory: /var/lib/grafana/plugins  
Log directory: /var/log/grafana/

* You may have to use 'sudo' to edit the configuration file or to view the log file.
* You may have to change permissons on the 'plugin' directory, for example: sudo chmod 777 /var/lib/grafana/plugins
* Under the plugin directory (/var/lib/grafana/plugins), create a directory called 'gtm-grafana-datasource'.

From the unzipped archive, copy:
* LICENSE
* README.md
* img (directory and its contents)
* module.js
* module.js.LICENSE.txt
* module.js.map
* plugin.json  
to /var/lib/grafana/plugins/gtm-grafana-datasource

From the unzipped archive, copy one of (as appropriate for your hardware):
* gpx_akamai-gtm-datasource-plugin_linux_amd64
* gpx_akamai-gtm-datasource-plugin_linux_arm
* gpx_akamai-gtm-datasource-plugin_linux_arm64  
to /var/lib/grafana/plugins/gtm-grafana-datasource

### Macintosh

Configuration file: /usr/local/etc/grafana/grafana.ini  
Plugin directory: /usr/local/var/lib/grafana/plugins  
Log directory: /usr/local/var/log/grafana/

Under the plugin directory (/usr/local/var/lib/grafana/plugins), create a directory called 'gtm-grafana-datasource'.

From the unzipped archive, copy:
* LICENSE
* README.md
* img (directory and its contents)
* module.js
* module.js.LICENSE.txt
* module.js.map
* plugin.json
to /usr/local/var/lib/grafana/plugins/gtm-grafana-datasource

From the unzipped archive, copy:
* gpx_akamai-gtm-datasource-plugin_darwin_amd64  
to /var/lib/grafana/plugins/gtm-grafana-datasource

### Windows

Grafana can be installed into any directory (install_dir).  

Configuration file: install_dir\conf  
Plugin directory: install_dir\data\plugins  
Log directory: install_dir\data\log

Under the plugin directory (install_dir\data\plugins), create a directory called 'gtm-grafana-datasource'.

From the unzipped archive, copy:
* LICENSE
* README.md
* img (directory and its contents)
* module.js
* module.js.LICENSE.txt
* module.js.map
* plugin.json
to install_dir\data\plugins\gtm-grafana-datasource

From the unzipped archive, copy:
* gpx_akamai-gtm-datasource-plugin_windows_amd64.exe  
to install_dir\data\plugins\gtm-grafana-datasource

### Grafana configuration

[Configuration](https://grafana.com/docs/grafana/latest/administration/configuration/) describes configuration for each 
operating system.  Carefully read the directions.

* Using a text editor, open the configuration file (as described in [Configuration](https://grafana.com/docs/grafana/latest/administration/configuration/)).

* Under the [paths] section header, uncomment "plugins" by removing the semicolon.  For example:
```
[paths]
# Directory where grafana will automatically scan and look for plugins
plugins = /var/lib/grafana/plugins
```
* To the right of "plugins =", insert the complete path to the plugin directory.  
  NOTE: The plugin directory differs by operating system!

* Under the [plugins] section header, uncomment "allow_loading_unsigned_plugins".
* To the right of "allow_loading_unsigned_plugins =", add "akamai-gtm-datasource" (without quotes).  For example:
```
[plugins]
# Enter a comma-separated list of plugin identifiers to identify plugins that are allowed to be loaded even if they lack a valid signature.
allow_loading_unsigned_plugins = akamai-gtm-datasource
```

### Restart Grafana
[Restart Grafana](https://grafana.com/docs/grafana/latest/installation/restart-grafana/)
describes how to restart Grafana for each operating system.

Under the log directory for your operating system, in "grafana.log", you should see something similar to:
```
t=2021-03-24T10:31:09-0400 lvl=info msg="Registering plugin" logger=plugins id=akamai-gtm-datasource
```

[Troubleshooting](https://grafana.com/docs/grafana/latest/troubleshooting/) contains troubleshooting tips.

### Log in to Grafana
[Getting started with Grafana](https://grafana.com/docs/grafana/latest/getting-started/getting-started/) 
describes how to log in to Grafana.  The default username/password are: admin/admin.

## "Akamai GTM Datasource" Configuration

Select Configuration (gear icon) -> Datasources -> Akamai GTM Datasource

In the datasource configuration panel, enter your Akamai credentials.

![Data Source](https://github.com/akamai/gtm-grafana-datasource-plugin/blob/master/static/data-source-config.png)

Create a new dashboard and add a panel.

In each query, enter one domain name. Create additional queries, as needed.

![Domain](https://github.com/akamai/gtm-grafana-datasource-plugin/blob/master/static/domains-config.png)

Metric name is optional. If empty then a metric name is automatically generated.

![Metric Name](https://github.com/akamai/gtm-grafana-datasource-plugin/blob/master/static/metric-name-config.png)

