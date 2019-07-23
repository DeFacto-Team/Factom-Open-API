import React, { useState, useEffect, useLayoutEffect } from 'react';
import axios from 'axios';

import {
  Typography,
  Button,
  Icon,
  Divider,
  Input,
  message,
  Form
} from 'antd';
import { NotifyNetworkError } from './../common/Notifications';

const { Title, Paragraph } = Typography;

const Settings = () => {
  const [formHasErrors, setFormHasErrors] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [settings, setSettings] = useState({});

  const handleSubmit = event => {
    event.preventDefault();
    setIsSubmitting(true);

    const form = event.target;
    const data = new FormData(form);

    axios
      .post('/admin/settings', data)
      .then(function() {
        setIsSubmitting(false);
        message.success(`Settings updated`);
      })
      .catch(function(error) {
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
        setIsSubmitting(false);
      })
      .finally(function() {
        restartAPI();
      });
  };

  const restartAPI = () => {

    setSettings([]);

    axios
      .get('/admin/restart')
      .then(function() {
        message.loading('Restarting API…', 0);
        setTimeout(function() {
          window.parent.location = window.parent.location.href;
        }, 5000);
      })
      .catch(function(error) {
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
      });
  };

  const getSettings = () => {
    axios
      .get('/admin/settings')
      .then(function(response) {
        setSettings(response.data);
      })
      .catch(function(error) {
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
        setIsSubmitting(false);
      });
  };

  useEffect(() => getSettings(), []);

  return (
    <div>
      <Title level={3}>Settings</Title>
      <Paragraph type="secondary"><Icon type="info-circle" theme="twoTone" /> API server will be restarted after settings update. Current admin session will be terminated.</Paragraph>
      
      {settings.Admin ? (
        <Form layout="vertical" onSubmit={handleSubmit}>

          <Divider orientation="left">Admin</Divider>

          <Form.Item label="User">
            <Input size="large" name="adminUser" defaultValue={settings.Admin.adminUser} />
          </Form.Item>

          <Form.Item label="Password">
            <Input size="large" name="adminPassword" defaultValue={settings.Admin.adminPassword} />
          </Form.Item>

          <Divider orientation="left">Factom</Divider>

          <Form.Item label="Es Address">
            <Input size="large" name="factomEsAddress" defaultValue={settings.Factom.factomEsAddress} />
          </Form.Item>

          <Form.Item label="Factomd URL">
            <Input size="large" name="factomURL" defaultValue={settings.Factom.factomURL} />
          </Form.Item>

          <Form.Item label="User">
            <Input size="large" name="factomUser" defaultValue={settings.Factom.factomUser} />
          </Form.Item>

          <Form.Item label="Password">
            <Input size="large" name="factomPassword" defaultValue={settings.Factom.factomPassword} />
          </Form.Item>

          <Divider orientation="left">Database</Divider>

          <Form.Item label="Host">
            <Input size="large" name="storeHost" defaultValue={settings.Store.storeHost} />
          </Form.Item>

          <Form.Item label="Port">
            <Input size="large" name="storePort" defaultValue={settings.Store.storePort} />
          </Form.Item>

          <Form.Item label="User">
            <Input size="large" name="storeUser" defaultValue={settings.Store.storeUser} />
          </Form.Item>

          <Form.Item label="Password">
            <Input size="large" name="storePassword" defaultValue={settings.Store.storePassword} />
          </Form.Item>

          <Form.Item label="Database">
            <Input size="large" name="storeDBName" defaultValue={settings.Store.storeDBName} />
          </Form.Item>

          <Divider />

          <Form.Item>
            <Button
              type="primary"
              icon="check"
              htmlType="submit"
              size="large"
              loading={isSubmitting}
            >
              Save
            </Button>
          </Form.Item>
        </Form>
      ) :
        <div style={{ color: '#1890ff' }}>
          <Icon type="loading" />
        </div>
      }

    </div>
  );
};

export default Settings;
