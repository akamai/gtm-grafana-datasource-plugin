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

import * as React from 'react';
import { ChangeEvent, PureComponent } from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { Field, Input } from '@grafana/ui';
import { MyDataSourceOptions } from './types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}
interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onClientSecretChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      clientSecret: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onHostChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      host: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onAccessTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      accessToken: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onClientTokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      clientToken: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  render() {
    const { options } = this.props;
    const { jsonData } = options;

    return (
      <div>
        <Field label="Client Secret" description="Enter client secret">
          <Input
            value={jsonData.clientSecret || ''}
            onChange={this.onClientSecretChange}
            placeholder="Enter client secret"
          />
        </Field>

        <Field label="Host" description="Enter host">
          <Input
            value={jsonData.host || ''}
            onChange={this.onHostChange}
            placeholder="Enter host"
          />
        </Field>

        <Field label="Access Token" description="Enter access token">
          <Input
            value={jsonData.accessToken || ''}
            onChange={this.onAccessTokenChange}
            placeholder="Enter access token"
          />
        </Field>

        <Field label="Client Token" description="Enter client token">
          <Input
            value={jsonData.clientToken || ''}
            onChange={this.onClientTokenChange}
            placeholder="Enter client token"
          />
        </Field>
      </div>
    );
  }
}