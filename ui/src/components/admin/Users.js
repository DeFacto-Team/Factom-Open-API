import React, { useState, useEffect } from 'react';
import axios from 'axios';

import {
  Typography,
  Button,
  Icon,
  Table,
  Divider,
  Input,
  Popconfirm,
  Tooltip,
  message,
  Tag,
  Form
} from 'antd';
import { NotifyNetworkError } from './../common/Notifications';
import EditableText from './../common/EditableText';

const { Title, Text } = Typography;

const Users = () => {
  const [formHasErrors, setFormHasErrors] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [users, setUsers] = useState([]);
  const [tableIsLoading, setTableIsLoading] = useState(false);

  const validateForm = event => setFormHasErrors(event.target.value === '');

  const toggleUserStatus = user => {
    const status = user.status === 0 ? 1 : 0;
    updateUser(user, 'status', status);
  };

  const handleSubmit = event => {
    event.preventDefault();
    setIsSubmitting(true);

    const form = event.target;
    const data = new FormData(form);

    axios
      .post('/admin/users', data)
      .then(function(response) {
        const user = response.data.result;
        setUsers([
          ...users,
          {
            id: user.id,
            name: user.name,
            accessToken: user.accessToken,
            usage: user.usage,
            usageLimit: user.usageLimit,
            status: user.status
          }
        ]);
        setIsSubmitting(false);
        message.success(`User '${user.name}' added`);
      })
      .catch(function(error) {
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
        setIsSubmitting(false);
      });
  };

  const updateUser = (user, field, value) => {
    const array = [...users];
    const index = array.findIndex(v => v.id === user.id);
    array[index][field] = value;

    var payload = {};
    payload[field] = value;

    axios
      .put('/admin/users/' + user.id, payload)
      .then(function() {
        setUsers(array);
        message.success(`User '${user.name}' updated`);
      })
      .catch(function(error) {
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
      });
  };

  const rotateToken = user => {
    axios
      .post('/admin/users/rotate', { id: user.id })
      .then(function(response) {
        const array = [...users];
        const index = array.findIndex(v => v.id === user.id);
        array[index].accessToken = response.data.result.accessToken;
        setUsers(array);
        message.success(`Access token for '${user.name}' rotated`);
      })
      .catch(function(error) {
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
      });
  };

  const deleteUser = user => {
    setTableIsLoading(true);

    axios
      .delete('/admin/users', { data: { id: user.id } })
      .then(function() {
        const remainingUsers = [...users].filter(v => v.id !== user.id);
        setUsers(remainingUsers);
        setTableIsLoading(false);
        message.success(`User '${user.name}' deleted`);
      })
      .catch(function(error) {
        setTableIsLoading(false);
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
      });
  };

  const getUsers = () => {
    axios
      .get('/admin/users')
      .then(function(response) {
        setUsers(response.data.result);
      })
      .catch(function(error) {
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
      });
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      sorter: (a, b) => a.name.localeCompare(b.name),
      defaultSortOrder: 'ascend',
      sortDirections: ['descend', 'ascend'],
      render: (text, user) => (
        <EditableText
          text={user.name}
          placeholder="Name"
          onSave={value => updateUser(user, 'name', value)}
        />
      )
    },
    {
      title: 'Token',
      dataIndex: 'accessToken',
      render: (text, user) => (
        <Text copyable={{ text: user.accessToken }}>
          {user.accessToken.substring(0, 6)}…
        </Text>
      )
    },
    {
      title: 'Usage',
      dataIndex: 'usage',
      render: (text, user) => (
        <EditableText
          text={user.usage}
          placeholder="0"
          type="number"
          onSave={value => updateUser(user, 'usage', value)}
        />
      )
    },
    {
      title: 'Usage Limit',
      dataIndex: 'usageLimit',
      render: (text, user) => (
        <EditableText
          text={user.usageLimit}
          placeholder="0"
          type="number"
          onSave={value => updateUser(user, 'usageLimit', value)}
        />
      )
    },
    {
      title: 'Status',
      dataIndex: 'status',
      render: (text, user) => (
        <span>
          <Tag
            color={user.status ? 'green' : 'orange'}
            onClick={() => toggleUserStatus(user)}
            className="pointer"
          >
            {user.status ? 'ON' : 'OFF'}
          </Tag>
        </span>
      )
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (text, user) => (
        <span>
          <Popconfirm
            title={`Rotate access token for '${user.name}'?`}
            onConfirm={() => rotateToken(user)}
            okText="Rotate"
            cancelText="No"
          >
            <a href="javascript:;">
              <Icon type="switcher" theme="twoTone" />
               Rotate token
            </a>
          </Popconfirm>
          <Divider type="vertical" />
          <Popconfirm
            title={`Delete user '${user.name}'?`}
            onConfirm={() => deleteUser(user)}
            okText="Delete"
            cancelText="No"
          >
            <a href="javascript:;" style={{ color: '#f5222d' }}>
              <Icon type="delete" theme="twoTone" twoToneColor="#f5222d" />
               Delete
            </a>
          </Popconfirm>
        </span>
      )
    }
  ];

  useEffect(() => getUsers(), []);

  return (
    <div>
      <Title level={4}>Users</Title>
      <Form
        layout="inline"
        noValidate
        className="new-user-form"
        onSubmit={handleSubmit}
      >
        <Form.Item>
          <Input
            placeholder="New user name"
            size="large"
            name="name"
            onChange={validateForm}
            suffix={
              <Tooltip title="Any name to identify client/project">
                <Icon type="info-circle" style={{ color: 'rgba(0,0,0,.45)' }} />
              </Tooltip>
            }
          />
        </Form.Item>
        <Form.Item>
          <Button
            type="primary"
            icon="plus"
            size="large"
            htmlType="submit"
            disabled={formHasErrors}
            loading={isSubmitting}
          >
            Add user
          </Button>
        </Form.Item>
      </Form>
      <Table
        dataSource={users}
        columns={columns}
        rowKey="id"
        loading={tableIsLoading}
      />
    </div>
  );
};

export default Users;
