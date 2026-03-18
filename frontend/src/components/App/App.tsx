import Header from '../Header/Header';
import Dashboard from '../Dashboard/Dashboard';
import PlayerList from '../PlayerList/PlayerList';
import Login from '../Login/Login';
import CreateProfile from '../CreateProfile/CreateProfile';
import { useState, useEffect, useCallback } from 'react';
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
  useParams,
} from 'react-router-dom';
import EventForm from '../EventForm/EventForm';
import EditTeeTime from '../EditTeeTime/EditTeeTime';
import {
  getAllCourses,
  getAllPlayers,
  getAllEvents,
  getFriendsEvents,
  joinEvent,
  postInviteAction,
  deleteEvent,
  postFriendship,
  deleteFriendship,
  validateStandardLogin,
} from '../../APICalls/APICalls';
import type { Event, Friend, Course, Player } from '../../types';

const SESSION_TIMEOUT_MS = 30 * 60 * 1000; // 30 minutes
const ACTIVITY_KEY = 'last_activity';

function getPlayerIdFromToken(): number {
  try {
    const token = localStorage.getItem('jwt_token');
    if (!token) return 0;
    const payload = JSON.parse(atob(token.split('.')[1]));
    if (payload.exp * 1000 < Date.now()) return 0;
    return payload.player_id || 0;
  } catch {
    return 0;
  }
}

function isSessionActive(): boolean {
  const last = localStorage.getItem(ACTIVITY_KEY);
  if (!last) return false;
  return Date.now() - Number(last) < SESSION_TIMEOUT_MS;
}

function touchActivity() {
  localStorage.setItem(ACTIVITY_KEY, String(Date.now()));
}

function App() {
  const restoredId = getPlayerIdFromToken();
  const initialPlayer = restoredId && isSessionActive() ? restoredId : 0;

  const [events, setEvents] = useState<Event[]>([]);
  const [screenWidth, setScreenWidth] = useState<number>(window.innerWidth);
  const [allPlayers, setAllPlayers] = useState<Player[]>([]);
  const [hostPlayer, setHostPlayer] = useState<number>(initialPlayer);
  const [friends, setFriends] = useState<Friend[]>([]);
  const [courses, setCourses] = useState<Course[]>([]);
  const [friendsEvents, setFriendsEvents] = useState<Event[]>([]);
  const [loginError, setLoginError] = useState<string>('');

  const addFriend = (friend: Friend) => {
    if (!hostPlayer) {
      return;
    }

    postFriendship(hostPlayer, friend.id).then((data) => {
      setFriends([
        ...friends,
        {
          id: data.followee.id,
          name: data.followee.name,
        },
      ]);
    });
  };

  const removeFriend = (unFriend: Friend) => {
    if (!hostPlayer) {
      return;
    }

    deleteFriendship(hostPlayer, unFriend.id).then(
      () => {
        setFriends([
          ...friends.filter((f) => f.id !== unFriend.id),
        ]);
      }
    );
  };

  const updateInvite = (eventId: number, status: string) => {
    postInviteAction(hostPlayer, eventId, status).then((events) => {
      setEvents(events);
      getFriendsEvents(hostPlayer).then(setFriendsEvents);
    });
  };

  const validateLogin = (email: string, password: string) => {
    setLoginError('');
    validateStandardLogin(email, password)
      .then(data => {
        if (!data) {
          setLoginError('Invalid email or password. Please try again.');
          return;
        }
        setHostPlayer(data.id);
        if (data.token) {
          localStorage.setItem('jwt_token', data.token);
        }
        touchActivity();
        // Friends will be populated from allPlayers once hostPlayer is set
      })
      .catch(() => {
        setLoginError('Unable to sign in right now. Please try again.');
      });
  };

  const logout = useCallback(() => {
    localStorage.removeItem('jwt_token');
    localStorage.removeItem(ACTIVITY_KEY);
    setHostPlayer(0);
    setEvents([]);
    setFriends([]);
    setFriendsEvents([]);
  }, []);

  // Track user activity and enforce 30-min inactivity timeout
  useEffect(() => {
    if (!hostPlayer) return;

    const onActivity = () => touchActivity();
    window.addEventListener('click', onActivity);
    window.addEventListener('keydown', onActivity);

    const interval = setInterval(() => {
      if (!isSessionActive()) {
        logout();
      }
    }, 60_000); // check every minute

    return () => {
      window.removeEventListener('click', onActivity);
      window.removeEventListener('keydown', onActivity);
      clearInterval(interval);
    };
  }, [hostPlayer, logout]);

  const cancelCommitment = (event: Event) => {
    if (event.host_id === hostPlayer) {
      deleteEvent(event.id, hostPlayer).then((events) => {
        setEvents(events);
        getFriendsEvents(hostPlayer).then(setFriendsEvents);
      });
    } else {
      postInviteAction(hostPlayer, event.id, 'declined').then((events) => {
        setEvents(events);
        getFriendsEvents(hostPlayer).then(setFriendsEvents);
      });
    }
  };

  const joinTeeTime = (eventId: number) => {
    joinEvent(hostPlayer, eventId).then((events) => {
      setEvents(events);
      setFriendsEvents((prev) => prev.filter((e) => e.id !== eventId));
    });
  };

  const refreshEvents = () => {
    getAllEvents(hostPlayer).then((events) => setEvents(events));
    getFriendsEvents(hostPlayer).then(setFriendsEvents);
  };

  const handleResize = () => setScreenWidth(window.innerWidth);

  useEffect(() => {
    getAllPlayers().then((players) => {
      setAllPlayers(players);
    });
    getAllCourses().then((courses) => setCourses(courses));
  }, []);

  useEffect(() => {
    if (hostPlayer) {
      // Build friend list from allPlayers + the logged-in player's friends array
      const currentPlayer = allPlayers.find((p) => p.id === hostPlayer);
      if (currentPlayer && currentPlayer.friends) {
        const friendList = allPlayers
          .filter((p) => currentPlayer.friends.includes(p.id))
          .map((f) => ({ name: f.name, id: f.id }));
        setFriends(friendList);
      }

      getAllEvents(hostPlayer).then((events) => {
        setEvents(events);
      });
      getFriendsEvents(hostPlayer).then(setFriendsEvents);
    }
  }, [allPlayers, hostPlayer]);

  useEffect(() => {
    window.addEventListener('resize', handleResize);
  }, []);

  return (
    <Router>
      <Header screenWidth={screenWidth} isLoggedIn={!!hostPlayer} onLogout={logout} />
      <Routes>
        <Route
          path='/login'
          element={
            hostPlayer ? (
              <Navigate to='/dashboard' replace />
            ) : (
              <Login
                validateLogin={validateLogin}
                loginError={loginError}
                clearLoginError={() => setLoginError('')}
              />
            )
          }
        />
        <Route
          path='/create-profile'
          element={<CreateProfile />}
        />
        <Route
          path='/dashboard'
          element={
            !hostPlayer ? (
              <Navigate to='/login' replace />
            ) : (
              <Dashboard
                events={events}
                friendsEvents={friendsEvents}
                currentUserId={hostPlayer}
                currentUserName={allPlayers.find((p) => p.id === hostPlayer)?.name || ''}
                screenWidth={screenWidth}
                handleInviteAction={{
                  update: updateInvite,
                  cancel: cancelCommitment,
                  join: joinTeeTime,
                }}
                players={allPlayers}
                friends={friends}
                handleFriends={{ add: addFriend, remove: removeFriend }}
              />
            )
          }
        />
        <Route
          path='/community'
          element={
            !hostPlayer ? (
              <Navigate to='/login' replace />
            ) : screenWidth > 480 ? (
              <Navigate to='/dashboard' />
            ) : (
              <PlayerList
                screenWidth={screenWidth}
                userId={hostPlayer}
                players={allPlayers}
                friends={friends}
                handleFriends={{ add: addFriend, remove: removeFriend }}
              />
            )
          }
        />
        <Route
          path='/event-form'
          element={
            !hostPlayer ? (
              <Navigate to='/login' replace />
            ) : (
              <EventForm
                courses={courses}
                friends={friends}
                hostId={hostPlayer}
                refreshEvents={refreshEvents}
              />
            )
          }
        />
        <Route
          path='/edit-tee-time/:eventId'
          element={
            !hostPlayer ? (
              <Navigate to='/login' replace />
            ) : (
              <EditTeeTimeRoute
                events={events}
                friends={friends}
                refreshEvents={refreshEvents}
              />
            )
          }
        />
        <Route path='*' element={<Navigate to='/login' />} />
      </Routes>
    </Router>
  );
}

function EditTeeTimeRoute({ events, friends, refreshEvents }: { events: Event[]; friends: Friend[]; refreshEvents: () => void }) {
  const { eventId } = useParams();
  const event = events.find((e) => e.id === Number(eventId));

  if (!event) {
    return <Navigate to='/dashboard' replace />;
  }

  return <EditTeeTime event={event} friends={friends} refreshEvents={refreshEvents} />;
}

export default App;
