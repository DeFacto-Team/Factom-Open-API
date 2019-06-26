import React, { useState, useEffect } from 'react';
import axios from 'axios';

import { Typography, Modal, Button, Icon, Table, Divider, Popconfirm, message } from 'antd';

const { Title } = Typography;

const Users = () => {

    const [modalShown, setModalShown] = useState(false);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [users, setUsers] = useState(null);
    
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

    const confirmDeleteUser = event => {
        console.log(event);
        message.success('Click on Yes');
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
            title: 'ID',
            dataIndex: 'id',
        },
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
                    onConfirm={confirmDeleteUser}
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
            <Table rowSelection={rowSelection} dataSource={users} columns={columns} />
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