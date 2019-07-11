import React from 'react';
import { message } from 'antd';

export function NotifyNetworkError() {
  message.error('Open API server is unavailable');

  return <NotifyNetworkError />;
}
