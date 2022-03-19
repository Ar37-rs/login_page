import { Routes, Route, Link } from "react-router-dom";
import './App.css';
import ProfileView from './components/ProfileView';
import Home from './components/Home';
import Signup from './components/Signup';

function App() {
  return (
    <div className="container">
      <h1>Login page + Google OAuth2</h1>
      <p>
        Fullstack built with Echo/golang + React/TS
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
