import { FC, useEffect, useState } from 'react';
import { Navigate } from "react-router-dom";
import axios from 'axios';

interface UnauthorizedProps {
  content: {
    name: string,
    email: string
  };
  setLogin: React.Dispatch<React.SetStateAction<boolean>>;
}

const Unauthorized: FC<UnauthorizedProps> = (prop) => {
  return (<div>
    <div>{prop.content.name} {prop.content.email}</div>
    <div>
      <br />
      <button className='btn btn-primary' onClick={() => {
        prop.setLogin(true);
      }}>Login?</button>
    </div>
  </div>)
}

interface ContentProps {
  content: {
    name: string,
    email: string
  };
  handleLogout: (event: { preventDefault: () => void; }) => void;
}

const Authorized: FC<ContentProps> = (prop) => {
  return (<div>
    <div>
      <div>{prop.content.name} {prop.content.email}</div>
      <div>congratulations, you are logged in.</div>
      <br />
      <div>
        <button className='btn btn-primary' onClick={prop.handleLogout}>Logout?</button>
      </div>
    </div>
  </div>)
}

function ProfileView() {
  const [content, setContetnt] = useState({ name: '', email: '' });
  const [fetching, setFetching] = useState(true);
  const [logged, setLogged] = useState(false);
  const [logout, setLogout] = useState(false);
  const [login, setLogin] = useState(false);

  useEffect(() => {
    if (fetching) {
      axios.get('http://localhost:1323/profile', { withCredentials: true }).then(function (res) {
        if (res.status === 200) {
          setContetnt({ name: res.data.name, email: res.data.email });
          setLogged(true);
          setFetching(false);
        }
      }, function (error) {
        if (error.message) {
          setContetnt({ name: error.message, email: "" });
          setLogged(false);
          setFetching(false);
        } else {
          setContetnt({ name: error.response.message, email: "" });
          setLogged(false);
          setFetching(false);
        }
      });
    }
  }, [fetching]);

  const handleLogout = (event: { preventDefault: () => void; }) => {
    event.preventDefault();
    axios.get('http://localhost:1323/logout_api', { withCredentials: true }).then(function (res) {
      if (res.data.message === "Logged_out") {
        setLogout(true);
      }
    });
  }

  if (logout || login) {
    return (<Navigate to="/" replace={true} />);
  }

  if (fetching) {
    return (
      <div>
        <div className="spinner-border" role={content.name}>
        </div>
      </div>
    )
  } else {
    return (
      <div>
        {logged ? <Authorized content={content} handleLogout={handleLogout}></Authorized> : <Unauthorized content={content} setLogin={setLogin}></Unauthorized>}
      </div>
    )
  }
}

export default ProfileView;
