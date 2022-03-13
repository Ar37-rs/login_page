import { useState } from 'react';
import { Navigate } from "react-router-dom";
import axios from 'axios';

function Signup() {
  const [submitting, setSubmitting] = useState(false);
  const [formdata, setformData] = useState<{
    name: string;
    email: string;
    password: string;
  }>({ name: '', email: '', password: '' });
  const [errors_msgs, setErrorsMsgs] = useState<{ name: string, email: string, password: string }>({ name: '', email: '', password: '' });
  const [errors_conds, setErrorsConds] = useState<{ name_err: boolean, email_err: boolean, password_err: boolean }>({ name_err: true, email_err: true, password_err: true });
  const [neterror, setNeterror] = useState('');
  const [redirect, setRedirect] = useState(false);
  const [login, setLogin] = useState(false);
  const [_message, setMessage] = useState('');
  const [showpassword, setShowPassword] = useState("password");

  const handleSubmit = (event: { preventDefault: () => void; }) => {
    event.preventDefault();
    setSubmitting(true);
    let valid = !errors_conds.name_err ? !errors_conds.email_err ? !errors_conds.password_err ? true : false : false : false;

    if (valid) {
      axios.post("http://localhost:1323/signup_api", formdata, { withCredentials: true }).then(function (res) {
        if (res.data.message === "EmailTaken") {
          setErrorsMsgs({ ...errors_msgs, email: "Email is already taken!" });
          setErrorsConds({ ...errors_conds, email_err: true });
          setMessage(res.data.message);
        }
        if (res.data.message === "SignupAccepted") {
          setMessage("Signed up, will be redirected to /profile_view");
          setTimeout(() => {
            setRedirect(true);
          }, 1000);
        }

        if (res.data.message === "Authorized") {
          setRedirect(true);
        }
      }, function (error) {
        setNeterror(error.message)
      })
    }
  }

  if (login) {
    return (<Navigate to="/" replace={true} />);
  }

  if (submitting && redirect) {
    return (<Navigate to="/profile_view" replace={true} />);
  } else {
    return (
      <div>
        <h1>Signup {neterror}</h1>
        <div>
          <form>
            <div className="mb-3">
              <label className="form-label">
                <p>Userame</p>
                <input className="form-control" name="username" onChange={e => {
                  setSubmitting(false)
                  let value = e.target.value;
                  if (value === '') {
                    setErrorsMsgs({ ...errors_msgs, name: "Username can not be empty!" });
                    setErrorsConds({ ...errors_conds, email_err: true });
                  } else if (value.length < 3) {
                    setErrorsMsgs({ ...errors_msgs, name: "Username must contains 3 of chars at least." });
                    setErrorsConds({ ...errors_conds, name_err: true });
                  } else {
                    setErrorsMsgs({ ...errors_msgs, name: '' });
                    setErrorsConds({ ...errors_conds, name_err: false });
                    setformData({ ...formdata, name: value });
                  }
                }} />
              </label>
            </div>
            <div>{errors_msgs.name ? errors_msgs.name : ''}</div>
            <label className="form-label">
              <p>Email</p>
              <input className="form-control" name="email" onChange={e => {
                setSubmitting(false);
                let value = e.target.value;
                if (value === '') {
                  setErrorsMsgs({ ...errors_msgs, email: "Email can not be empty!" })
                  setErrorsConds({ ...errors_conds, email_err: true });
                } else if (!/^[a-zA-Z0-9]+@(?:[a-zA-Z0-9]+\.)+[A-Za-z]+$/.test(value)) {
                  setErrorsMsgs({ ...errors_msgs, email: "Email invalid." });
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
                setSubmitting(false);
                let value = e.target.value;
                if (value === '') {
                  setErrorsMsgs({ ...errors_msgs, password: 'Password can not be empty!' });
                  setErrorsConds({ ...errors_conds, password_err: true });
                } else if (value.length < 6) {
                  setErrorsMsgs({ ...errors_msgs, password: 'Password length must >= 6 at least.' });
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
                    setShowPassword("text")
                  } else {
                    setShowPassword("password")
                  }
                }} />
                {showpassword === "password" ? "Show" : "Hide"} password?
              </div>
            </label>
            <div>{errors_msgs.password ? errors_msgs.password : ''}</div>
          </form>
          <br />
          <div className="d-grid gap-2 d-md-flex justify-content-md-center">
            <button className="btn btn-primary" onClick={handleSubmit}>Signup</button>
            <button className="btn btn-primary" onClick={() => {
              setLogin(true);
            }}>Login</button>
             <img src="/GoogleBtn.svg" id="clickable-img" onClick={() => {
              window.location.href = 'http://localhost:1323/login_with_google';
            }}></img> 
          </div>
        </div>
      </div>
    )
  }
}

export default Signup;