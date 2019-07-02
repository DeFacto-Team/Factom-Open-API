import React, { useState, useEffect } from 'react';
import axios from 'axios';

import { Typography, Modal, Button, Icon, Table, Divider, Popconfirm, message, Tag } from 'antd';
import { NotifyNetworkError } from './../common/Notifications';

const { Title } = Typography;

const Users = () => {

    const [modalShown, setModalShown] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [users, setUsers] = useState([]);
    
    const handleOk = event => {
        setIsSubmitting(true);
        setTimeout(() => {
            setModalShown(false);
            setIsSubmitting(false);
        }, 2000);
    };

    const handleCancel = event => {
        setModalShown(false);
    };

    const addUser = (id, name, status) => {
        setUsers([
            ...users,
            {
                id: id,
                name: name,
                status: status
            }
        ]);
    };

    const deleteUser = (id) => {
        var array = [...users];
        var index = array.findIndex(v => v.id === id);
        array.splice(index, 1);
        setUsers(array);
    };

    const confirmDeleteUser = (id, name) => {
        axios.delete("/admin/users", { data: { id: id }})
        .then(function (response) {
            deleteUser(id);
            message.success(`User '${name}' deleted`);
        })
        .catch(function (error) {
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
        },
        {
            title: 'Token',
            dataIndex: 'accessToken',
        },
        {
            title: 'Usage',
            dataIndex: 'usage',
        },
        {
            title: 'Usage Limit',
            dataIndex: 'usageLimit',
        },
        {
            title: 'Status',
            dataIndex: 'status',
            render: (text, record) => (
                <span>
                    <Tag color={record.status ? 'green' : 'orange'}>
                        {record.status ? 'ON' : 'OFF'}
                    </Tag>
                </span>
            ),
        },
        {
            title: 'Action',
            key: 'action',
            render: (text, user) => (
                <span>
                  <a href="javascript:;"><Icon type="edit" theme="twoTone" /> Edit</a>
                  <Divider type="vertical" />
                  <Popconfirm
                    title={`Delete user '${user.name}'?`}
                    onConfirm={() => confirmDeleteUser(user.id, user.name)}
                    okText="Delete"
                    cancelText="No"
                  >
                    <a href="javascript:;"><Icon type="delete" theme="twoTone" /> Delete</a>
                  </Popconfirm>
                </span>
            ),
        },
    ];

    // rowSelection object indicates the need for row selection
    const rowSelection = {
        onChange: (selectedRowKeys, selectedRows) => {
            console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', selectedRows);
        },
        getCheckboxProps: record => ({
            disabled: record.name === 'Disabled User', // Column configuration not to be checked
            name: record.name,
        }),
    };
      
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
            <Table rowSelection={rowSelection} dataSource={users} columns={columns} rowKey="id" />
            <Modal
                title="Title"
                visible={modalShown}
                onOk={handleOk}
                confirmLoading={isSubmitting}
                onCancel={handleCancel}
            >
                <p>Test</p>
            </Modal>
        </div>
    );

};

export default Users;