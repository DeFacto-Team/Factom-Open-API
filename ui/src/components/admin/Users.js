import React, { useState, useEffect } from 'react';
import axios from 'axios';

import { Typography, Modal, Button, Icon, Table, Divider, Input, Popconfirm, Form, message, Tag } from 'antd';
import { NotifyNetworkError } from './../common/Notifications';
import EditableText from './../common/EditableText';

const { Title, Text } = Typography;

const Users = () => {

    const [modalShown, setModalShown] = useState(false);
    const [modalHasErrors, setModalHasErrors] = useState(false);
    const [modalIsSubmitting, setModalIsSubmitting] = useState(false);
    const [users, setUsers] = useState([]);
    const [tableIsLoading, setTableIsLoading] = useState(false);
    
    const handleFormOk = event => {
        setModalIsSubmitting(true);
        setTimeout(() => {
            setModalShown(false);
            setModalIsSubmitting(false);
        }, 2000);
    };

    const handleFormCancel = event => {
        setModalShown(false);
    };

    const toggleUserStatus = (user) => {

        setTableIsLoading(true);

        const array = [...users];
        const index = array.findIndex(v => v.id === user.id);
        array[index].status = array[index].status === 0 ? 1 : 0;

        axios.put("/admin/users/"+user.id, { status: array[index].status })
        .then(function () {
            setUsers(array);
            setTableIsLoading(false);
            const statusText = array[index].status === 1 ? "enabled" : "disabled";
            message.success(`User '${user.name}' ${statusText}`);
        })
        .catch(function (error) {
            setTableIsLoading(false);
            if (error.response) {
                message.error(error.response.data.error);
            }
            else {
                NotifyNetworkError();
            }
        });

    }

    const createUser = (name) => {

        axios.post("/admin/users", { name: name })
        .then(function (response) {
            const user = response.data.result;
            setUsers([
                ...users,
                {
                    id: user.id,
                    name: user.name,
                    accessToken: user.accessToken,
                    usage: 0,
                    usageLimit: user.usageLimit,
                    status: user.status,
                }
            ]);
    
        })
        
    };

    const rotateToken = (user) => {

        setTableIsLoading(true);

        axios.post("/admin/users/rotate", { id: user.id })
        .then(function (response) {
            const array = [...users];
            const index = array.findIndex(v => v.id === user.id);
            array[index].accessToken = response.data.result.accessToken;
            setUsers(array);
            setTableIsLoading(false);
            message.success(`Access token for '${user.name}' rotated`);
        })
        .catch(function (error) {
            setTableIsLoading(false);
            if (error.response) {
                message.error(error.response.data.error);
            }
            else {
                NotifyNetworkError();
            }
        });

    }

    const deleteUser = (user) => {

        setTableIsLoading(true);

        axios.delete("/admin/users", { data: { id: user.id }})
        .then(function () {
            const remainingUsers = [...users].filter(v => v.id !== user.id);
            setUsers(remainingUsers);
            setTableIsLoading(false);
            message.success(`User '${user.name}' deleted`);
        })
        .catch(function (error) {
            setTableIsLoading(false);
            if (error.response) {
                message.error(error.response.data.error);
            }
            else {
                NotifyNetworkError();
            }
        });

    }
      
    const getUsers = () => {
        axios.get("/admin/users")
        .then(function (response) {
          setUsers(response.data.result);
        })
        .catch(function (error) {
          console.log(error.message);
        });
    }
    
    const columns = [
        {
            title: 'Name',
            dataIndex: 'name',
            sorter: (a, b) => a.name.localeCompare(b.name),
            defaultSortOrder: 'ascend',
            sortDirections: ['descend', 'ascend'],
            render: (text, user) => (
                <EditableText text={user.name} placeholder="name" />
            ),
        },
        {
            title: 'Token',
            dataIndex: 'accessToken',
            render: (text, user) => (
                <Text copyable={{ text: user.accessToken }}>{user.accessToken.substring(0,6)}…</Text>
            ),
        },
        {
            title: 'Usage',
            dataIndex: 'usage',
            render: (text, user) => (
                <EditableText text={user.usage} placeholder="0" type="number" />
            )
        },
        {
            title: 'Usage Limit',
            dataIndex: 'usageLimit',
            render: (text, user) => (
                <EditableText text={user.usageLimit} placeholder="0" type="number" />
            )
        },
        {
            title: 'Status',
            dataIndex: 'status',
            render: (text, user) => (
                <span>
                    <Tag color={user.status ? 'green' : 'orange'} onClick={() => toggleUserStatus(user)} className="pointer">
                        {user.status ? 'ON' : 'OFF'}
                    </Tag>
                </span>
            ),
        },
        {
            title: 'Action',
            key: 'action',
            render: (text, user) => (
                <span>
                  <Popconfirm
                    title={`Rotate access token for '${user.name}'?`}
                    onConfirm={() => rotateToken(user)}
                    okText="Rotate"
                    cancelText="No"
                  >
                    <a href="javascript:;"><Icon type="switcher" theme="twoTone" /> Rotate token</a>
                  </Popconfirm>
                  <Divider type="vertical" />
                  <Popconfirm
                    title={`Delete user '${user.name}'?`}
                    onConfirm={() => deleteUser(user)}
                    okText="Delete"
                    cancelText="No"
                  >
                    <a href="javascript:;" style={{ color: '#f5222d' }}><Icon type="delete" theme="twoTone" twoToneColor="#f5222d" /> Delete</a>
                  </Popconfirm>
                </span>
            ),
        },
    ];
      
    useEffect(() => getUsers(), []);
    
    return (
        <div>
            <Title level={4}>
                Users
            </Title>
            <p>
            <Button type="primary" onClick={() => setModalShown(true)} className="titleAdd">
                <Icon type="user-add" />
                New user
            </Button>
            </p>
            <Table dataSource={users} columns={columns} rowKey="id" loading={tableIsLoading} />
            <Modal
                title="New user"
                visible={modalShown}
                onOk={handleFormOk}
                confirmLoading={modalIsSubmitting}
                onCancel={handleFormCancel}
            >
                <Input addonBefore="Name" size="large" style={{ width: '50%' }} />
                <Input addonBefore="Usage Limit" size="large" style={{ width: '50%' }} />
            </Modal>
        </div>
    );

};

export default Users;