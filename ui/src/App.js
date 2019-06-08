import React, { Component } from 'react';
import { Menu, Icon, Layout, Button } from 'antd';
import './App.css';

const { Header, Content, Footer } = Layout;

class App extends Component {

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
              <span id="queue-size">0</span>
            </Menu.Item>
            <Menu.Item key="settings">
              <Icon type="setting" />
              Settings
            </Menu.Item>
          </Menu>
        </Header>
        <Content style={{ padding: 24, margin: 0 }}>
          Content
        </Content>
        <Footer style={{ padding: '18px 24px', margin: 0, background: '#fff', textAlign: 'right' }}>
          API version:
        </Footer>
      </Layout>
    );
  }
}

export default App;