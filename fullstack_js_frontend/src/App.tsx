import { useEffect, useState } from 'react';
import { Routes, Route, Link, Navigate } from "react-router-dom";
import { GoogleLogin } from 'react-google-login';
import './App.css';
import axios from 'axios';

function App() {
  return (
    <div className="container">
      <h1>Mini Aplikasi Login + Google OAuth2 </h1>

      <p>
        Dibuat dalam rangka mengikuti program IP - Refactory.id
      </p>
      <Routes>
        <Route index element={<Home />} />
        <Route path="signup" element={<Signup />} />
        <Route path="profile_view" element={<ProfileView />} />
        <Route path="*" element={<NoMatch />} />
      </Routes>
    </div>
  );
}

function Home() {
  const [submitting, setSubmitting] = useState(false);
  const [formdata, setformData] = useState<{
    email: string;
    password: string;
  }>({ email: '', password: '' });
  const [errors_msgs, setErrorsMsgs] = useState<{ email: string, password: string }>({ email: '', password: '' });
  const [errors_conds, setErrorsConds] = useState<{ email_err: boolean, password_err: boolean }>({ email_err: true, password_err: true });
  const [signup, setSignup] = useState(false);
  const [logged, setLogged] = useState(false);
  const [message, setMessage] = useState('');

  const handleSubmit = (event: { preventDefault: () => void; }) => {
    event.preventDefault();
    setSubmitting(true);
    let valid = !errors_conds.email_err ? !errors_conds.password_err ? true : false : false

    if (valid) {
      axios.post('http://localhost:1323/login_api', formdata, { withCredentials: true }).then(function (v) {
        if (v.data.message === "InvalidPassword") {
          setErrorsMsgs({ ...errors_msgs, password: "Invalid Password" })
          setErrorsConds({ ...errors_conds, password_err: true })
        }
        if (v.data.message === "Logged") {
          setLogged(true)
        }

        if (v.data.message === "Unregistered") {
          setMessage("Email & Password are not registerd, try signup?");
        }

        if (v.data.message === "Authorized") {
          setLogged(true)
        }
      })
    }
  }

  window.onload = function () {
    axios.get('http://localhost:1323/profile', { withCredentials: true }).then(function (v) {
      if (v.status === 200) {
        setSubmitting(true);
        setLogged(true)
      }
    })
  }

  if (submitting && logged) {
    return (<Navigate to="/profile_view" replace={true} />);
  } else if (signup) {
    return (<Navigate to="/signup" replace={true} />);
  } else {
    return (
      <div>
        <h1>Login</h1>
        {message}
        <div>
          <form>
            <label>
              <p>Email</p>
              <input name="email" onChange={e => {
                setSubmitting(false)
                let value = e.target.value;
                if (value === '') {
                  setErrorsMsgs({ ...errors_msgs, email: 'Email can not be empty!' })
                  setErrorsConds({ ...errors_conds, email_err: true })
                } else if (!/^[a-zA-Z0-9]+@(?:[a-zA-Z0-9]+\.)+[A-Za-z]+$/.test(value)) {
                  setErrorsMsgs({ ...errors_msgs, email: 'Email invalid.' })
                  setErrorsConds({ ...errors_conds, email_err: true })
                } else {
                  setErrorsMsgs({ ...errors_msgs, email: '' })
                  setErrorsConds({ ...errors_conds, email_err: false })
                  setformData({ ...formdata, email: value });
                }
              }} />
            </label>
            {errors_msgs.email ? errors_msgs.email : ''}
            <label>
              <p>Password</p>
              <input name="password" onChange={e => {
                setSubmitting(false)
                let value = e.target.value;
                if (value === '') {
                  setErrorsMsgs({ ...errors_msgs, password: 'Password can not be empty!' })
                  setErrorsConds({ ...errors_conds, password_err: true })
                } else if (value.length < 6) {
                  setErrorsMsgs({ ...errors_msgs, password: 'Password length must >= 6 at least.' })
                  setErrorsConds({ ...errors_conds, password_err: true })
                } else {
                  setErrorsMsgs({ ...errors_msgs, password: '' })
                  setErrorsConds({ ...errors_conds, password_err: false })
                  setformData({ ...formdata, password: value });
                }
              }} />
            </label>
            {errors_msgs.password ? errors_msgs.password : ''}
          </form>
          <button onClick={handleSubmit}>Login</button>
          <button onClick={() => {
            setSignup(true)
          }}>Signup</button>
          <GoogleLogin
            clientId="578732033166-4rvtsrgn4s5s0ppfrbqg46cd2ihtt71c.apps.googleusercontent.com"
            buttonText="Login with"
            onSuccess={responseGoogle}
            onFailure={responseGoogle}
            cookiePolicy={'single_host_origin'}
          />
        </div>
      </div>
    )
  }
}

const responseGoogle = (response: any) => {
  console.log(response);
}

function ProfileView() {
  let [user, setUser] = useState({ name: '', email: '' })
  let [unauthorized, setunAuthorized] = useState(false)
  window.onload = async function () {
    let v = await axios.get('http://localhost:1323/profile', { withCredentials: true })
    if (v.status === 200) {
      setUser({ name: v.data.name, email: v.data.email })
    } else {
      setunAuthorized(true)
    }

  }
  if (unauthorized) {
    return (<div>
      Unauthorized 401
    </div>)
  } else {
    return (<div>
      {user.name ? "Hello" : "Page not shown? need to be refreshed maybe?"} {user.name} {user.email}
    </div>)
  }
}

function Signup() {
  const [submitting, setSubmitting] = useState(false);
  const [formdata, setformData] = useState<{
    name: string;
    email: string;
    password: string;
  }>({ name: '', email: '', password: '' });
  const [errors_msgs, setErrorsMsgs] = useState<{ name: string, email: string, password: string }>({ name: '', email: '', password: '' });
  const [errors_conds, setErrorsConds] = useState<{ name_err: boolean, email_err: boolean, password_err: boolean }>({ name_err: true, email_err: true, password_err: true });
  const [redirect, setRedirect] = useState(false);

  const [message, setMessage] = useState('');

  const handleSubmit = (event: { preventDefault: () => void; }) => {
    event.preventDefault();
    setSubmitting(true);
    let valid = !errors_conds.name_err ? !errors_conds.email_err ? !errors_conds.password_err ? true : false : false : false

    if (valid) {
      axios.post('http://localhost:1323/signup_api', formdata, { withCredentials: true }).then(function (v) {
        if (v.data.message === "EmailTaken") {
          setErrorsMsgs({ ...errors_msgs, email: "email is already taken!" })
          setErrorsConds({ ...errors_conds, email_err: true })
          setMessage(v.data.message);
        }
        if (v.data.message === "SignupAccepted") {
          setMessage("Signed up, will be redirected to login page...");
          setTimeout(() => {
            setRedirect(true)
          }, 3000);
        }

        if (v.data.message === "Authorized") {
          setRedirect(true)
        }
      })
    }
  }

  window.onload = function () {
    axios.get('http://localhost:1323/profile', { withCredentials: true }).then(function (v) {
      if (v.status === 200) {
        setRedirect(true)
        setSubmitting(true)
      }
    })
  }

  if (submitting && redirect) {
    return (<Navigate to="/profile_view" replace={true} />);
  } else {
    return (
      <div>
        <h1>Login</h1>
        {message}
        <div>
          <form>
            <label>
              <p>Userame</p>
              <input name="username" onChange={e => {
                setSubmitting(false)
                let value = e.target.value;
                if (value === '') {
                  setErrorsMsgs({ ...errors_msgs, name: 'Username can not be empty!' })
                  setErrorsConds({ ...errors_conds, email_err: true })
                } else if (value.length < 3) {
                  setErrorsMsgs({ ...errors_msgs, name: 'Username must contains 3 of chars at least.' })
                  setErrorsConds({ ...errors_conds, name_err: true })
                } else {
                  setErrorsMsgs({ ...errors_msgs, name: '' })
                  setErrorsConds({ ...errors_conds, name_err: false })
                  setformData({ ...formdata, name: value });
                }
              }} />
            </label>
            {errors_msgs.name ? errors_msgs.name : ''}
            <label>
              <p>Email</p>
              <input name="email" onChange={e => {
                setSubmitting(false)
                let value = e.target.value;
                if (value === '') {
                  setErrorsMsgs({ ...errors_msgs, email: 'Email can not be empty!' })
                  setErrorsConds({ ...errors_conds, email_err: true })
                } else if (!/^[a-zA-Z0-9]+@(?:[a-zA-Z0-9]+\.)+[A-Za-z]+$/.test(value)) {
                  setErrorsMsgs({ ...errors_msgs, email: 'Email invalid.' });
                  setErrorsConds({ ...errors_conds, email_err: true })
                } else {
                  setErrorsMsgs({ ...errors_msgs, email: '' })
                  setErrorsConds({ ...errors_conds, email_err: false })
                  setformData({ ...formdata, email: value });
                }
              }} />
            </label>
            {errors_msgs.email ? errors_msgs.email : ''}
            <label>
              <p>Password</p>
              <input name="password" onChange={e => {
                setSubmitting(false)
                let value = e.target.value;
                if (value === '') {
                  setErrorsMsgs({ ...errors_msgs, password: 'Password can not be empty!' })
                  setErrorsConds({ ...errors_conds, password_err: true })
                } else if (value.length < 6) {
                  setErrorsMsgs({ ...errors_msgs, password: 'Password length must >= 6 at least.' })
                  setErrorsConds({ ...errors_conds, password_err: true })
                } else {
                  setErrorsMsgs({ ...errors_msgs, password: '' })
                  setErrorsConds({ ...errors_conds, password_err: false })
                  setformData({ ...formdata, password: value });
                }
              }} />
            </label>
            {errors_msgs.password ? errors_msgs.password : ''}
          </form>
          <button onClick={handleSubmit}>Signup</button>
          <GoogleLogin
            clientId="578732033166-4rvtsrgn4s5s0ppfrbqg46cd2ihtt71c.apps.googleusercontent.com"
            buttonText="Login with"
            onSuccess={responseGoogle}
            onFailure={responseGoogle}
            cookiePolicy={'single_host_origin'}
          />
        </div>
      </div>
    )
  }
}

function NoMatch() {
  return (
    <div>
      <h2>Page Not Found</h2>
      <p>
        <Link to="/">Go to the home page</Link>
      </p>
    </div>
  );
}

export default App;
