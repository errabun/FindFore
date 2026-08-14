import Header from '../Header/Header';
import Dashboard from '../Dashboard/Dashboard';
import PlayerList from '../PlayerList/PlayerList';
import Login from '../Login/Login';
import CreateProfile from '../CreateProfile/CreateProfile';
import EventForm from '../EventForm/EventForm';
import EditTeeTime from '../EditTeeTime/EditTeeTime';
import Profile from '../Profile/Profile';
import GroupsPage from '../Groups/GroupsPage';
import GroupDetailPage from '../Groups/GroupDetailPage';
import CreateGroupForm from '../Groups/CreateGroupForm';
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
  useParams,
} from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { useTeeTimes } from '../../hooks/useTeeTimes';
import { useFriends } from '../../hooks/useFriends';
import { useScreenWidth } from '../../hooks/useScreenWidth';
import type { Event, Friend } from '../../types';

function App() {
  const { hostPlayer, loginError, allPlayers, validateLogin, logout, clearLoginError, updateProfile, changePassword } = useAuth();
  const { events, friendsEvents, updateInvite, cancelCommitment, joinTeeTime, refreshEvents } = useTeeTimes(hostPlayer);
  const {
    friends,
    incomingRequests,
    outgoingPendingIds,
    requestFriend,
    acceptRequest,
    declineRequest,
    removeFriend,
  } = useFriends(hostPlayer, allPlayers);
  const screenWidth = useScreenWidth();

  const currentUserName = allPlayers.find((p) => p.id === hostPlayer)?.name || '';
  const handleFriends = {
    request: requestFriend,
    remove: removeFriend,
    accept: acceptRequest,
    decline: declineRequest,
  };

  return (
    <Router>
      <Header screenWidth={screenWidth} isLoggedIn={!!hostPlayer} onLogout={logout} />
      <main className="ff-main-content">
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
                clearLoginError={clearLoginError}
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
                currentUserName={currentUserName}
                screenWidth={screenWidth}
                handleInviteAction={{
                  update: updateInvite,
                  cancel: cancelCommitment,
                  join: joinTeeTime,
                }}
                players={allPlayers}
                friends={friends}
                incomingRequests={incomingRequests}
                outgoingPendingIds={outgoingPendingIds}
                handleFriends={handleFriends}
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
                incomingRequests={incomingRequests}
                outgoingPendingIds={outgoingPendingIds}
                handleFriends={handleFriends}
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
                courses={[]}
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
        <Route
          path='/groups'
          element={
            !hostPlayer ? (
              <Navigate to='/login' replace />
            ) : (
              <GroupsPage hostPlayer={hostPlayer} />
            )
          }
        />
        <Route
          path='/groups/new'
          element={
            !hostPlayer ? (
              <Navigate to='/login' replace />
            ) : (
              <CreateGroupForm />
            )
          }
        />
        <Route
          path='/groups/:groupId'
          element={
            !hostPlayer ? (
              <Navigate to='/login' replace />
            ) : (
              <GroupDetailPage hostPlayer={hostPlayer} players={allPlayers} />
            )
          }
        />
        <Route
          path='/profile'
          element={
            !hostPlayer ? (
              <Navigate to='/login' replace />
            ) : (
              (() => {
                const currentPlayer = allPlayers.find((p) => p.id === hostPlayer);
                return currentPlayer ? (
                  <Profile
                    player={currentPlayer}
                    onUpdateProfile={updateProfile}
                    onChangePassword={changePassword}
                    onPasswordChanged={logout}
                  />
                ) : null;
              })()
            )
          }
        />
        <Route path='*' element={<Navigate to='/login' />} />
      </Routes>
      </main>
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
