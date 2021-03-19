/*
 * Copyright 2021 Akamai Technologies, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, { ChangeEvent, PureComponent } from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { LegacyForms } from '@grafana/ui';
import { MyDataSourceOptions } from './types';

const { FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}
interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onClientSecretChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      clientSecret: event.target.value,
    };
    console.log('clientSecret: ' + event.target.value);
    onOptionsChange({ ...options, jsonData });
  };

  onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      host: event.target.value,
    };
    console.log('host: ' + event.target.value);
    onOptionsChange({ ...options, jsonData });
  };

  onAccessTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      accessToken: event.target.value,
    };
    console.log('accessToken: ' + event.target.value);
    onOptionsChange({ ...options, jsonData });
  };

  onClientTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      clientToken: event.target.value,
    };
    console.log('clientToken: ' + event.target.value);
    onOptionsChange({ ...options, jsonData });
  };

  render() {
    const { options } = this.props;
    const { jsonData } = options;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <FormField
            label="Client Secret"
            labelWidth={8}
            inputWidth={24}
            onChange={this.onClientSecretChange}
            value={jsonData.clientSecret || ''}
            placeholder="Enter client secret"
          />
        </div>
        <div className="gf-form">
          <FormField
            label="Host"
            labelWidth={8}
            inputWidth={24}
            onChange={this.onHostChange}
            value={jsonData.host || ''}
            placeholder="Enter host"
          />
        </div>
        <div className="gf-form">
          <FormField
            label="Access Token"
            labelWidth={8}
            inputWidth={24}
            onChange={this.onAccessTokenChange}
            value={jsonData.accessToken || ''}
            placeholder="Enter access token"
          />
        </div>
        <div className="gf-form">
          <FormField
            label="Client Token"
            labelWidth={8}
            inputWidth={24}
            onChange={this.onClientTokenChange}
            value={jsonData.clientToken || ''}
            placeholder="Enter client token"
          />
        </div>
      </div>
    );
  }
}
