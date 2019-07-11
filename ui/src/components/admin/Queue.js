import React, { useState, useEffect } from 'react';
import Moment from 'react-moment';
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

const Queue = () => {
  const [formHasErrors, setFormHasErrors] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [queue, setQueue] = useState([]);
  const [tableIsLoading, setTableIsLoading] = useState(false);

  const getQueue = () => {
    axios
      .get('/admin/queue')
      .then(function(response) {
        setQueue(response.data.result);
      })
      .catch(function(error) {
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
      });
  };

  const deleteQueue = item => {
    setTableIsLoading(true);

    axios
      .delete('/admin/queue', { data: { id: item.id } })
      .then(function() {
        const remainingQueue = [...queue].filter(v => v.id !== item.id);
        setQueue(remainingQueue);
        setTableIsLoading(false);
        message.success(`Queue item #${item.id} deleted`);
      })
      .catch(function(error) {
        console.log(error);
        setTableIsLoading(false);
        if (error.response) {
          message.error(error.response.data.error);
        } else {
          NotifyNetworkError();
        }
      });
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      sorter: (a, b) => a.id - b.id,
      defaultSortOrder: 'descend',
      sortDirections: ['descend', 'ascend'],
    },
    {
      title: 'Created (UTC+'+ -(new Date().getTimezoneOffset() / 60) + ')',
      dataIndex: 'createdAt',
      sorter: (a, b) => new Date(a.createdAt) - new Date(b.createdAt),
      sortDirections: ['descend', 'ascend'],
      render: (text, queue) => (
        <Moment date={queue.createdAt} format="YYYY-MM-DD HH:mm:ss" local />
      )
    },
    {
      title: 'Processed (UTC+'+ -(new Date().getTimezoneOffset() / 60) + ')',
      dataIndex: 'processedAt',
      sorter: (a, b) => new Date(a.processedAt) - new Date(b.processedAt),
      sortDirections: ['descend', 'ascend'],
      render: (text, queue) => (
        <Moment date={queue.processedAt} format="YYYY-MM-DD HH:mm:ss" local />
      )
    },
    {
      title: 'Action',
      dataIndex: 'action',
      render: (text, queue) => (
        <span>{queue.action}</span>
      )
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (text, queue) => (
        <span>
          <Popconfirm
            title={`Delete queue item #${queue.id}?`}
            onConfirm={() => deleteQueue(queue)}
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

  useEffect(() => getQueue(), []);

  return (
    <div>
      <Title level={4}>Queue</Title>
      <Table
        dataSource={queue}
        columns={columns}
        rowKey="id"
        loading={tableIsLoading}
      />
    </div>
  );
};

export default Queue;
