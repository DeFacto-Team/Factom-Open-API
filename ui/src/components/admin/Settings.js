import React, { useState, useEffect, useLayoutEffect } from 'react';
import axios from 'axios';

import {
  Typography,
  Button,
  Icon,
  Divider,
  Input,
  message,
  Form,
  Switch,
  Card
} from 'antd';
import { NotifyNetworkError } from './../common/Notifications';

const { Title, Paragraph, Text } = Typography;

const Settings = () => {
  const [formHasErrors, setFormHasErrors] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [settings, setSettings] = useState({});
  const [address, setAddress] = useState({});
  const [factomPassword, setFactomPassword] = useState(0);

  const toggleFactomPassword = () => {
    const current = factomPassword === 0 ? 1 : 0;
    setFactomPassword(current);
  };

  const handleSubmit = event => {
    event.preventDefault();
    setIsSubmitting(true);

    const form = event.target;
    const data = new FormData(form);

    axios
      .post('/admin/settings', data)
      .then(function() {
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

    axios
      .get('/admin/restart')
      .then(function() {
        message.loading('Restarting API…', 0);
        setTimeout(function() {
          window.parent.location = window.parent.location.href;
        }, 8000);
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
        if (response.data.Factom.factomEsAddress) {
          getECBalance(response.data.Factom.factomEsAddress);
        }
      })
      .catch(function(error) {
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
      });
  };

  const getECBalance = (esaddress) => {

    axios
      .get('/admin/ec/'+esaddress)
      .then(function(response) {
        setAddress(response.data.result);
      })
      .catch(function(error) {
        setAddress({});
      });    
  }

  const getRandomAddress = settings => {
    
    axios
      .get('/admin/ec/random')
      .then(function(response) {
        setAddress(response.data.result);
      })
      .catch(function(error) {
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
        setAddress({});
      });
  }

  useEffect(() => getSettings(), []);

  return (
    <div className="settings-form">
      <Title level={3}>Settings</Title>

      <Paragraph type="secondary"><Icon type="info-circle" theme="twoTone" /> API server will be restarted after settings update.<br />Current admin session will be terminated.</Paragraph>
      
      {settings.Admin ? (
        <Form layout="vertical" onSubmit={handleSubmit}>

          <Title level={4}>Admin Credentials</Title>
          <Divider />

          <Form.Item label="User">
            <Input prefix={<Icon type="user" style={{ color: 'rgba(0,0,0,.25)' }} />} size="large" name="adminUser" defaultValue={settings.Admin.adminUser} />
          </Form.Item>

          <Form.Item label="Password">
            <Input prefix={<Icon type="lock" style={{ color: 'rgba(0,0,0,.25)' }} />} size="large" name="adminPassword" defaultValue={settings.Admin.adminPassword} />
          </Form.Item>

          <Title level={4}>Factomd Endpoint</Title>
          <Divider />

          <Form.Item label="Factomd URL">
            <Input prefix={<Icon type="global" style={{ color: 'rgba(0,0,0,.25)' }} />} size="large" name="factomURL" defaultValue={settings.Factom.factomURL} />
          </Form.Item>

          <Form.Item>
            <Switch size="small" checked={factomPassword ? true : false} onClick={toggleFactomPassword} />
            <Text>Password to access factomd</Text>
          </Form.Item>

          {factomPassword ? (
            <div>
              <Form.Item label="User">
                <Input prefix={<Icon type="user" style={{ color: 'rgba(0,0,0,.25)' }} />} size="large" name="factomUser" defaultValue={settings.Factom.factomUser} />
              </Form.Item>

              <Form.Item label="Password">
                <Input prefix={<Icon type="lock" style={{ color: 'rgba(0,0,0,.25)' }} />} size="large" name="factomPassword" defaultValue={settings.Factom.factomPassword} />
              </Form.Item>
            </div>
          ) :
            null
          }
          <Title level={4}>Factom EC address</Title>
          <Divider />

          <Form.Item label="Private Es address">
            <Input placeholder="Es…" prefix={<Icon type="wallet" style={{ color: 'rgba(0,0,0,.25)' }} />} size="large" name="factomEsAddress" defaultValue={settings.Factom.factomEsAddress} onChange={(event) => getECBalance(event.target.value)} />
            {address.ecAddress ? (
              <div>
                <Card size="small" title={<div><Icon type="check-circle" theme="twoTone" twoToneColor="#52c41a" /><Text>  EC address is valid!</Text></div>} className="ec-block">
                  <Paragraph copyable={{ text: address.ecAddress }}>
                    <strong>EC address:</strong><br />
                    {address.ecAddress}
                  </Paragraph>
                  <Paragraph>
                    <strong>Balance:</strong><br />
                    {address.balance} EC
                  </Paragraph>
                  <Button type="primary" icon="credit-card" href="https://ec.de-facto.pro" target="_blank" style={{marginBottom: "6px"}}>
                    Buy Entry Credits
                  </Button>
                  <br />
                  <Text type="secondary">You need EC address filled with Entry Credits to write data on the Factom.</Text>
                </Card>
              </div>
            ) : (
              <div>
                <Button type="primary" style={{marginTop: "6px"}} onClick={getRandomAddress}>Generate random address</Button>
                <Paragraph class="ec-block"><Text type="danger"><Icon type="warning" theme="twoTone" twoToneColor="#f5222d" />{' '}Invalid EC address</Text></Paragraph>
              </div>
            )
            }
          </Form.Item>

          <Divider />

          <Form.Item>
            <Button
              type="primary"
              icon="check-circle"
              htmlType="submit"
              size="large"
              loading={isSubmitting}
            >
              Save settings
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
