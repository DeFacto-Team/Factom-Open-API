import React, { useState, useEffect, useRef } from 'react';
import Moment from 'react-moment';
import axios from 'axios';

import {
  Typography,
  Icon,
  Table,
  Popconfirm,
  Tooltip,
  message,
  Tag
} from 'antd';
import { NotifyNetworkError } from './../common/Notifications';

const { Title, Text, Paragraph } = Typography;

const Queue = () => {
  const [queue, setQueue] = useState([]);
  const [tableIsLoading, setTableIsLoading] = useState(true);

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
      })
      .finally(function() {
        setTableIsLoading(false);
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
      title: 'Status',
      dataIndex: 'processedAt',
      render: (text, queue) => {
        if (queue.processedAt) {
          return (
            <Tooltip placement="top" title={
              <span><b>Processed:</b> <Moment date={queue.processedAt} format="YYYY-MM-DD HH:mm:ss" local /></span>
            }>
              <Tag color="green">PROCESSED</Tag>
            </Tooltip>
          )
        } else {
          if (queue.tryCount > 0) {
            return (            
              <Tooltip placement="top" title={
                <span><b>Number of attempts:</b> {queue.tryCount}<br /><b>Next try:</b> <Moment date={queue.nextTryAt} format="YYYY-MM-DD HH:mm:ss" local /></span>
              }>
                <Tag color="red">FAILED</Tag>
              </Tooltip>
            )
          } else {
            return (            
              <Icon type="loading" style={{ color: "#1890ff" }} />
            )
          }
        }
      }
    },
    {
      title: 'EntryHash',
      dataIndex: 'result',
      render: (text, queue) => {
        if (queue.result) {
          return (
            <Text copyable={{ text: queue.result }}>
              {queue.result.substring(0, 6)}…
            </Text>
          )
        }
      }
    },
    {
      title: 'Type',
      dataIndex: 'action',
      render: (text, queue) => {
        if (queue.action === "chain") {
          return (
            <Text type="secondary">
              <Icon type="link" /> Chain
            </Text>
          )
        }
        else if (queue.action === "entry") {
          return (
            <Text type="secondary">
              <Icon type="number" /> Entry
            </Text>
          )
        }
      }
    },
    {
      title: 'Actions',
      key: 'actions',
      render: (queue) => {
        if (queue.processedAt) {
          return (
            <Text disabled>
              <Icon type="delete" theme="twoTone" twoToneColor="#BFBFBF" />
               Delete
            </Text>
          )
        } else {
          return (
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
      }
    }
  ];

  useEffect(() => getQueue(), []);
  
  useInterval(() => {
    getQueue();
  }, 1000);

  return (
    <div>
      <Title level={3}>Queue</Title>
      <Paragraph type="secondary"><Icon type="info-circle" theme="twoTone" /> Processed tasks are automatically cleared every hour</Paragraph>
      <Table
        dataSource={queue}
        columns={columns}
        rowKey="id"
        loading={tableIsLoading}
      />
    </div>
  );
};

function useInterval(callback, delay) {
  const savedCallback = useRef();

  // Remember the latest callback.
  useEffect(() => {
    savedCallback.current = callback;
  }, [callback]);

  // Set up the interval.
  useEffect(() => {
    function tick() {
      savedCallback.current();
    }
    if (delay !== null) {
      let id = setInterval(tick, delay);
      return () => clearInterval(id);
    }
  }, [delay]);
}

export default Queue;
