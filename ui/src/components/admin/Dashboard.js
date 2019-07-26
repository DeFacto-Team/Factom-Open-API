import React from 'react';

import {
  Typography,
  Row,
  Col,
  Statistic,
  Icon
} from 'antd';

const { Title, Paragraph } = Typography;

const Dashboard = () => {
  return (
    <div>
      <Title level={3}>Dashboard</Title>
      <Paragraph type="secondary"><Icon type="info-circle" theme="twoTone" /> Dashboard is in development</Paragraph>
    </div>
  );
};

export default Dashboard;
