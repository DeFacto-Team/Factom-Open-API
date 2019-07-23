import React, { useState } from 'react';
import { HashRouter as Router, Route, Link } from 'react-router-dom';
import axios from 'axios';

import { Menu, Icon, Layout } from 'antd';
import Logo from './common/Logo';
import Version from './common/Version';
import { NotifyNetworkError } from './common/Notifications';

import Dashboard from './admin/Dashboard';
import Queue from './admin/Queue';
import Users from './admin/Users';
import Settings from './admin/Settings';

const { Header, Content, Footer } = Layout;

const Admin = props => {
  const currentLocation = window.location.pathname;
  const [currentMenu, setCurrentMenu] = useState([currentLocation]);

  const handleMenuClick = e => {
    setCurrentMenu([e.key]);
  };

  const logout = () => {
    axios
      .get('/admin/logout')
      .then(function(response) {
        props.setLoggedIn(false);
      })
      .catch(function(error) {
        NotifyNetworkError();
      });
  };

  return (
    <Router>
      <Layout>
        <Header
          style={{ padding: 0, margin: 0, background: '#fff', height: 48 }}
        >
          <Menu
            onClick={handleMenuClick}
            selectedKeys={currentMenu}
            mode="horizontal"
          >
            <Logo />
            <Menu.Item key="/">
              <Link to="/">
                <Icon type="appstore" />
                Dashboard
              </Link>
            </Menu.Item>
            <Menu.Item key="/users">
              <Link to="/users">
                <Icon type="team" />
                Users
              </Link>
            </Menu.Item>
            <Menu.Item key="/queue">
              <Link to="/queue">
                <Icon type="inbox" />
                Queue
              </Link>
            </Menu.Item>
            <Menu.Item key="/settings">
              <Link to="/settings">
                <Icon type="setting" />
                Settings
              </Link>
            </Menu.Item>
            <Menu.Item key="/logout" onClick={logout} className="menu-logout">
              <Link to="/">
                <Icon type="logout" />
                Logout
              </Link>
            </Menu.Item>
          </Menu>
        </Header>
        <Content style={{ padding: 24, margin: 0 }}>
          <Route exact path="/" component={Dashboard} />
          <Route exact path="/users" component={Users} />
          <Route exact path="/queue" component={Queue} />
          <Route exact path="/settings" component={Settings} />
        </Content>
        <Footer style={{ padding: '18px 24px', margin: 0, background: '#fff' }}>
          <small>
            <code>
              Version: <Version />
            </code>
          </small>
        </Footer>
      </Layout>
    </Router>
  );
};

export default Admin;
