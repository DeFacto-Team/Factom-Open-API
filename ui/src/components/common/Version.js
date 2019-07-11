import React, { useState, useEffect } from 'react';
import axios from 'axios';

const Version = () => {
  const [version, setVersion] = useState('Loading…');

  const getVersion = () => {
    axios
      .get('/v1')
      .then(function(response) {
        setVersion(response.data.result.version);
      })
      .catch(function(error) {
        setVersion(error.message);
      });
  };

  useEffect(() => getVersion(), []);

  return <span>{version}</span>;
};

export default Version;
