import React, { useState } from 'react';
import axios from 'axios';

import { Menu, Icon, Layout } from 'antd';
import Logo from './common/Logo';
import Version from './common/Version';

const { Header, Content, Footer } = Layout;

const Admin = props => {

    const [currentMenu, setCurrentMenu] = useState(["dashboard"]);

    const handleMenuClick = e => {
        setCurrentMenu([e.key]);
    };

    const logout = () => {
        
        axios.get("/admin/logout")
        .then(function (response) {
          props.setLoggedIn(false)
        });
  
    }

    return (
        <Layout>
            <Header style={{ padding: 0, margin: 0, background: '#fff', height: 48 }}>
                <Menu onClick={handleMenuClick} selectedKeys={currentMenu} mode="horizontal">
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
                    <Menu.Item key="logout" onClick={logout}>
                        <Icon type="logout" />
                        Logout
                    </Menu.Item>
                </Menu>
            </Header>
            <Content style={{ padding: 24, margin: 0 }}>
                Content
            </Content>
            <Footer style={{ padding: '18px 24px', margin: 0, background: '#fff' }}>
                <small>
                    <code>
                        Version: <Version />
                    </code>
                </small>
            </Footer>
        </Layout>
    );
    
};

export default Admin;