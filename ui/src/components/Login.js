import React, { useState } from 'react';
import axios from 'axios';

import { Menu, Icon, Layout, Button, Form, Input, Typography } from 'antd';
import Logo from './common/Logo';
import Version from './common/Version';

const { Header, Content, Footer } = Layout;
const { Title, Text } = Typography;

const Login = props => {
    
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [loginError, setLoginError] = useState(null);

    const handleSubmit = event => {
        event.preventDefault();
        setIsSubmitting(true);
        if (!event.target.checkValidity()) {
            setLoginError("Please fill login/password");
            setIsSubmitting(false);
            return;
        }

        const form = event.target;
        const data = new FormData(form);
        
        axios.post('/login', data)
        .then(function (response) {
            setLoginError(null);
            props.setLoggedIn(true);
        })
        .catch(function (error) {
            setLoginError("Invalid credentials");
            setIsSubmitting(false);
        });

    };
        
    return (
        <Layout>
            <Header style={{ padding: 0, margin: 0, background: '#fff', height: 48 }}>
                <Menu selectedKeys={['login']} mode="horizontal">
                    <Logo />
                    <Menu.Item key="login">
                    Open API
                    </Menu.Item>
                </Menu>
            </Header>
            <Content style={{ padding: 24, margin: 0 }}>
                <Title level={4}>Administrator area</Title>
                <Form layout="inline" onSubmit={handleSubmit} noValidate>
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
                        <Button type="primary" htmlType="submit" loading={ isSubmitting }>
                            Login
                    </Button>
                    </Form.Item>
                </Form>
                { loginError ? <Text type="danger">{loginError}</Text> : null }
            </Content>
            <Footer style={{ padding: '18px 24px', margin: 0, background: '#fff' }}>
                <small><code>
                Version: <Version />
                </code></small>
            </Footer>
        </Layout>
    );

};

export default Login;