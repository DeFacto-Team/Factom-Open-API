import React, { useState, useEffect } from 'react';
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
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [settings, setSettings] = useState({});
  const [address, setAddress] = useState({});
  const [esAddress, setEsAddress] = useState("");
  const [invalidAddress, setInvalidAddress] = useState(false);
  const [factomPassword, setFactomPassword] = useState(0);
  const [cardIsLoading, setCardIsLoading] = useState(true);

  const toggleFactomPassword = () => {
    const current = factomPassword === 0 ? 1 : 0;
    setFactomPassword(current);
  };

  const handleSubmit = event => {
    event.preventDefault();
    setIsSubmitting(true);
    if (!event.target.checkValidity()) {
      setIsSubmitting(false);
      return;
    }

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
          window.location.href = "/";
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
        if (response.data.Factom.factomEsAddress) {
          setEsAddress(response.data.Factom.factomEsAddress);
          getECBalance(response.data.Factom.factomEsAddress);
        }
        if (response.data.Factom.factomUser || response.data.Factom.factomPassword) {
          setFactomPassword(1);
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

    setCardIsLoading(true);

    axios
      .get('/admin/ec/'+esaddress)
      .then(function(response) {
        setAddress(response.data.result);
      })
      .catch(function(error) {
        setInvalidAddress(true);
        setAddress({});
      })
      .finally(function () {
        setCardIsLoading(false);
      });
  }

  const changeEsAddress = (newAddress) => {

    setEsAddress(newAddress);

    if (newAddress === '') {
      setInvalidAddress(false);
      setAddress({});
    } else {
      getECBalance(newAddress);
    }

  }

  const getRandomAddress = () => {
    
    axios
      .get('/admin/ec/random')
      .then(function(response) {
        setAddress(response.data.result);
        setEsAddress(response.data.result.esAddress);
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

          <Title level={4}><Icon type="idcard" theme="twoTone" />  Admin Credentials</Title>
          <Divider />

          <Form.Item label="User">
            <Input required prefix={<Icon type="user" style={{ color: 'rgba(0,0,0,.25)' }} />} size="large" name="adminUser" defaultValue={settings.Admin.adminUser} />
          </Form.Item>

          <Form.Item label="Password">
            <Input.Password required prefix={<Icon type="lock" style={{ color: 'rgba(0,0,0,.25)' }} />} size="large" name="adminPassword" defaultValue={settings.Admin.adminPassword} />
          </Form.Item>

          <Title level={4}><Icon type="database" theme="twoTone" />  Factomd Node</Title>
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
                <Input.Password prefix={<Icon type="lock" style={{ color: 'rgba(0,0,0,.25)' }} />} size="large" name="factomPassword" defaultValue={settings.Factom.factomPassword} />
              </Form.Item>
            </div>
          ) :
            null
          }
          <Title level={4}><Icon type="wallet" theme="twoTone" />  Factom EC address</Title>
          <Divider />

          <Form.Item label="Private Es address">
            <Input allowClear placeholder="Es…" prefix={<Icon type="wallet" style={{ color: 'rgba(0,0,0,.25)' }} />} size="large" name="factomEsAddress" defaultValue={settings.Factom.factomEsAddress} value={esAddress} onChange={(event) => changeEsAddress(event.target.value)} />
            {!cardIsLoading ? (
              <div>
              {address.ecAddress ? (
                <div>
                  <Card size="small" title={<div><Icon type="check-circle" theme="twoTone" twoToneColor="#52c41a" /><Text>  Valid EC address</Text></div>} className="ec-block">
                    <Paragraph copyable={{ text: address.ecAddress }}>
                      <strong>EC address:</strong><br />
                      {address.ecAddress}
                    </Paragraph>
                    <Paragraph>
                      <strong>Balance:</strong><br />
                      {address.balance} EC
                    </Paragraph>
                    <Button type="primary" icon="credit-card" href={"https://ec.de-facto.pro/?ec="+address.ecAddress} target="_blank" style={{marginBottom: "6px"}}>
                      Buy Entry Credits
                    </Button>
                    <br />
                    <Text type="secondary">You need EC address filled with Entry Credits to write data on the Factom.</Text>
                  </Card>
                </div>
              ) : (
                <div>
                    <Card size="small" title={<div><Icon type="warning" theme="twoTone" twoToneColor="#f5222d" /><Text>  {invalidAddress ? "Invalid EC address" : "EC address not set"}</Text></div>} className="ec-block">
                      <Paragraph type="secondary">You need EC address to write data on the Factom.</Paragraph>
                      <Button icon="sync" type="primary" onClick={getRandomAddress}>Generate random address</Button>
                    </Card>
                </div>
              )
              }
              </div>
            ) : (
                <Card size="small" title={<div><Text style={{ color: '#1890ff' }}><Icon type="loading" />  Checking address…</Text></div>} className="ec-block" />
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
          <Icon type="loading" style={{ color: '#1890ff' }} />
        </div>
      }

    </div>
  );
};

export default Settings;
