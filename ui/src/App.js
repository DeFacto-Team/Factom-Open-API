import React, { useState, useEffect } from 'react';
import axios from 'axios';
// import { HashRouter as Router } from 'react-router-dom';
import Login from './components/Login';
import Admin from './components/Admin';
import './App.css';

axios.defaults.baseURL = process.env.REACT_APP_API_PATH;

const App = () => {
  
  const [loggedIn, setLoggedIn] = useState(false);

  const renderApp = () => {
    if (loggedIn) {
      return <Admin setLoggedIn={setLoggedIn} />;
    } else {
      return <Login setLoggedIn={setLoggedIn} />;
    }
  };

  const checkLogin = () => {

    axios.get("/admin")
      .then(function (response) {
        setLoggedIn(true);
      });

  };

  useEffect(() => {
    checkLogin();
  }, []);

  return (
    <div>{renderApp()}</div>
  );
  
};

export default App;