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
  let [content, setContetnt] = useState({ name: '', email: '' });
  let [logged, setLogged] = useState(false);
  let [logout, setLogout] = useState(false);
  let [login, setLogin] = useState(false);
  let [firstload, setFirstLoad] = useState(true);
  const handleLogout = (event: { preventDefault: () => void; }) => {
    event.preventDefault();
    axios.get('http://localhost:1323/logout_api', { withCredentials: true }).then(function (res) {
      if (res.data.message === "Logged_out") {
        setLogout(true);
      }
    });
  }

  useEffect(() => {
    const timer = window.setInterval(async () => {
      axios.get('http://localhost:1323/profile', { withCredentials: true }).then(function (res) {
        if (res.status === 200) {
          setContetnt({ name: res.data.name, email: res.data.email });
          setLogged(true);
          setFirstLoad(false)
        }
      }, function (error) {
        if (error.message) {
          setContetnt({ name: error.message, email: "" });
          setLogged(false);
          setFirstLoad(false)
        } else {
          setContetnt({ name: error.response.message, email: "" });
          setLogged(false);
          setFirstLoad(false)
        }
      });
    }, 800);
    return () => {
      window.clearInterval(timer);
    }
  }, [setLogged, setFirstLoad])

  if (logout || login) {
    return (<Navigate to="/" replace={true} />);
  }

  if (firstload) {
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