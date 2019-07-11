import React, { useState, useEffect } from 'react';
import { Spin } from 'antd';
import axios from 'axios';
import Login from './components/Login';
import Admin from './components/Admin';
import './App.css';

axios.defaults.baseURL = process.env.REACT_APP_API_PATH;

const App = () => {
  const [loggedIn, setLoggedIn] = useState(false);
  const [loaded, setLoaded] = useState(false);

  const renderApp = () => {
    if (loaded) {
      if (loggedIn) {
        return <Admin setLoggedIn={setLoggedIn} />;
      } else {
        return <Login setLoggedIn={setLoggedIn} />;
      }
    } else {
      return <Spin size="large" className="loader" />;
    }
  };

  const checkLogin = () => {
    axios
      .get('/admin')
      .then(function(response) {
        setLoggedIn(true);
      })
      .finally(function() {
        setLoaded(true);
      });
  };

  useEffect(() => {
    checkLogin();
  }, []);

  return <div>{renderApp()}</div>;
};

export default App;
