import React from 'react';
import { notification } from 'antd';

export function NotifyNetworkError() {

    notification["error"]({
        message: 'Network error',
        description:
        'Open API server is unavailable',
    });

    return (
        <NotifyNetworkError />
    );

}

export function NotifyLoginSuccess() {

    notification["success"]({
        message: 'Success login',
        description:
        'You successfully logged in as administrator',
    });

    return (
        <NotifyLoginSuccess />
    );

}