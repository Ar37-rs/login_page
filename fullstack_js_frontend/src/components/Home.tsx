import { useState } from 'react';
import { Navigate } from "react-router-dom";
import axios from 'axios';

function Home() {
  const [formdata, setformData] = useState<{
    email: string;
    password: string;
  }>({ email: '', password: '' });
  const [errors_msgs, setErrorsMsgs] = useState<{ email: string, password: string }>({ email: '', password: '' });
  const [errors_conds, setErrorsConds] = useState<{ email_err: boolean, password_err: boolean }>({ email_err: true, password_err: true });
  const [neterror, setNeterror] = useState('');
  const [signup, setSignup] = useState(false);
  const [submitted, setSubmitted] = useState(false);
  const [logged, setLogged] = useState(false);
  const [message, setMessage] = useState('');
  const [showpassword, setShowPassword] = useState("password");

  const handleSubmit = (event: { preventDefault: () => void; }) => {
    event.preventDefault();
    setSubmitted(true);

    let valid = !errors_conds.email_err ? !errors_conds.password_err ? true : false : false;

    if (valid) {
      axios.post("http://localhost:1323/login_api", formdata, { withCredentials: true }).then(function (res) {
        if (res.data.message === "InvalidPassword") {
          setErrorsMsgs({ ...errors_msgs, password: "Invalid Password" });
          setErrorsConds({ ...errors_conds, password_err: true });
        }
        if (res.data.message === "Logged") {
          setLogged(true);
        }

        if (res.data.message === "Unregistered") {
          setMessage("Email & Password are not registerd, try signup?");
        }

        if (res.data.message === "Authorized") {
          setLogged(true);
        }
      }, function (error) {
        setNeterror(error.message);
      })
    } else {
      setMessage("Please fill the required fields!");
    }
  }

  if (submitted && logged) {
    return (<Navigate to="/profile_view" replace={true} />);
  } else if (signup) {
    return (<Navigate to="/signup" replace={true} />);
  } else {
    return (
      <div>
        <h1>Login {neterror}</h1>
        <div>
          <div>{message}</div>
          <form>
            <div className="mb-3">
              <label className="form-label">
                <p>Email</p>
                <input className="form-control" name="email" onChange={e => {
                  setSubmitted(false);
                  setMessage('');
                  let value = e.target.value;
                  if (value === '') {
                    setErrorsMsgs({ ...errors_msgs, email: "Email can not be empty!" });
                    setErrorsConds({ ...errors_conds, email_err: true });
                  } else if (!/^[a-zA-Z0-9]+@(?:[a-zA-Z0-9]+\.)+[A-Za-z]+$/.test(value)) {
                    setErrorsMsgs({ ...errors_msgs, email: "Make sure to use vaild email." });
                    setErrorsConds({ ...errors_conds, email_err: true });
                  } else {
                    setErrorsMsgs({ ...errors_msgs, email: '' });
                    setErrorsConds({ ...errors_conds, email_err: false });
                    setformData({ ...formdata, email: value });
                  }
                }} />
              </label>
              <div>{errors_msgs.email ? errors_msgs.email : ''}</div>
              <label className="form-label">
                <p>Password</p>
                <input className="form-control" name="password" type={showpassword} onChange={e => {
                  setSubmitted(false);
                  setMessage('');
                  let value = e.target.value;
                  if (value === '') {
                    setErrorsMsgs({ ...errors_msgs, password: "Password can not be empty!" })
                    setErrorsConds({ ...errors_conds, password_err: true });
                  } else if (value.length < 6) {
                    setErrorsMsgs({ ...errors_msgs, password: "Password length must >= 6 at least." })
                    setErrorsConds({ ...errors_conds, password_err: true });
                  } else {
                    setErrorsMsgs({ ...errors_msgs, password: '' });
                    setErrorsConds({ ...errors_conds, password_err: false });
                    setformData({ ...formdata, password: value });
                  }
                }} />
                <div className="mb-3 form-check">
                  <input type="checkbox" className="form-check-input" onClick={() => {
                    if (showpassword == "password") {
                      setShowPassword("text");
                    } else {
                      setShowPassword("password");
                    }
                  }} />
                  {showpassword === "password" ? "Show" : "Hide"} password?
                </div>
              </label>
              <div>{errors_msgs.password ? errors_msgs.password : ''}</div>
            </div>
          </form>
          <br />
          <div className="d-grid gap-2 d-md-flex justify-content-md-center">
            <button className="btn btn-primary" onClick={handleSubmit}>Login</button>
            <button className="btn btn-primary" onClick={() => {
              setSignup(true);
            }}>Signup</button>
            <img src="/GoogleBtn.svg" id="clickable-img" onClick={() => {
              window.location.href = 'http://localhost:1323/login_with_google';
            }}></img>
          </div>
        </div>
      </div>
    )
  }
}

export default Home;