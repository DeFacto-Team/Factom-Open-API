import React, { Component } from 'react';

class Version extends Component {

    constructor(props) {
      super(props);
      this.state = {
        error: null,
        isLoaded: false,
        version: ""
      };
    }
  
    componentDidMount() {
      fetch(process.env.REACT_APP_API_PATH + "/v1")
        .then(res => res.json())
        .then(
          (result) => {
            this.setState({
              isLoaded: true,
              version: result.result.version
            });
          },
          (error) => {
            this.setState({
              isLoaded: true,
              error
            });
          }
        )
    }
  
    render() {
      const { error, isLoaded, version } = this.state;
      if (error) {
        return <span>{error.message}</span>;
      } else if (!isLoaded) {
        return <span>Loading...</span>;
      } else {
        return (
          <span>{version}</span>
        );
      }
    }

  }

  export default Version;