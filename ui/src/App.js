import React, { Component } from 'react';
import { Menu, Icon, Layout, Button, Form, Input, Typography } from 'antd';
import { HashRouter as Router } from 'react-router-dom';
import Version from './components/common/Version';
import Logo from './components/common/Logo';
import './App.css';

const { Header, Content, Footer } = Layout;
const { Title, Text } = Typography;

function hasErrors(fieldsError) {
  return Object.keys(fieldsError).some(field => fieldsError[field]);
}

class App extends Component {

  setState() {
    let idToken = localStorage.getItem("id_token");
    if (idToken) {
      this.loggedIn = true;
    } else {
      this.loggedIn = false;
    }
  }

  componentWillMount() {
    this.setState();
  }

  render() {
    if (this.loggedIn) {
      return (<AdminApp />);
    } else {
      return (<LoginForm />);
    }
  }
}

class LoginForm extends Component {

  constructor() {
    super();
    this.state = {};
    this.handleSubmit = this.handleSubmit.bind(this);
  }

  handleSubmit(event) {
    event.preventDefault();
    if (!event.target.checkValidity()) {
    	this.setState({
        loginError: "Please fill login/password"
      });
      return;
    }
    const form = event.target;
    const data = new FormData(form);
    
    fetch(process.env.REACT_APP_API_PATH + '/login', {
       method: 'POST',
       body: data,
    })
    .then(res => res.json())
    .then(
      (result) => {
        if (result.token) {
          this.setState({
            loginError: null,
          });
          sessionStorage.setItem("token", result.token);
        }
        else {
          this.setState({
            loginError: "Invalid credentials",
          });
        }
      },
      (error) => {
        this.setState({
          loginError: "Connection error",
        });
      }
    )
  }

  render() {
    const { res, invalid } = this.state;
    return (
      <Layout>
        <Header style={{ padding: 0, margin: 0, background: '#fff', height: 48 }}>
          <Menu selectedKeys={['login']} mode="horizontal">
            <Logo />
            <Menu.Item key="login">
              <Icon type="login" />
              Log in
            </Menu.Item>
          </Menu>
        </Header>
        <Content style={{ padding: 24, margin: 0 }}>
          <Title level={4}>Administrator area</Title>
          <Form layout="inline" onSubmit={this.handleSubmit} noValidate>
            <Form.Item>
                <Input
                  prefix={<Icon type="user" style={{ color: 'rgba(0,0,0,.25)' }} />}
                  placeholder="User"
                  id="user"
                  name="user"
                  required
                />
            </Form.Item>
            <Form.Item>
                <Input
                  prefix={<Icon type="lock" style={{ color: 'rgba(0,0,0,.25)' }} />}
                  type="password"
                  placeholder="Password"
                  id="password"
                  name="password"
                  required
                />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit">
                Log in
              </Button>
            </Form.Item>
          </Form>
          { this.state.loginError ? <Text type="danger">{this.state.loginError}</Text> : null }
        </Content>
        <Footer style={{ padding: '18px 24px', margin: 0, background: '#fff' }}>
          <small><code>
            Version: <Version />
          </code></small>
        </Footer>
      </Layout>
    );
  }

}

class AdminApp extends Component {

  state = {
    current: 'dashboard',
  };

  handleClick = e => {
    console.log('click ', e);
    this.setState({
      current: e.key,
    });
  };

  render() {
    return (
      <Layout>
        <Header style={{ padding: 0, margin: 0, background: '#fff', height: 48 }}>
          <Menu onClick={this.handleClick} selectedKeys={[this.state.current]} mode="horizontal">
              <Logo />
            <Menu.Item key="dashboard">
              <Icon type="appstore" />
              Dashboard
            </Menu.Item>
            <Menu.Item key="users">
              <Icon type="team" />
              Users
            </Menu.Item>
            <Menu.Item key="queue">
              <Icon type="inbox" />
              Queue
            </Menu.Item>
            <Menu.Item key="settings">
              <Icon type="setting" />
              Settings
            </Menu.Item>
            <Menu.Item key="logout">
              <Icon type="logout" />
              Logout
            </Menu.Item>
          </Menu>
        </Header>
        <Content style={{ padding: 24, margin: 0 }}>
          Content
        </Content>
        <Footer style={{ padding: '18px 24px', margin: 0, background: '#fff' }}>
          <small><code>
            Version: <Version />
          </code></small>
        </Footer>
      </Layout>
    );
  }
}

export default App;