import type { Player, Course, Event, LoginResponse, Post, Reaction, Reply } from '../types';

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? '';

const endpoints = {
  players: `${API_BASE}/api/v1/players`,
  courses: `${API_BASE}/api/v1/courses`,
  playerEvent: `${API_BASE}/api/v1/player-event`,
  joinEvent: `${API_BASE}/api/v1/player-event/join`,
  singleEvent: `${API_BASE}/api/v1/event`,
  friendship: `${API_BASE}/api/v1/friendship`,
  posts: `${API_BASE}/api/v1/posts`,
  sessions: `${API_BASE}/api/v1/sessions`,
};

function authHeaders(): Record<string, string> {
  const token = localStorage.getItem('jwt_token');
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

export const getAllPlayers = (): Promise<Player[]> => {
  return fetch(endpoints.players)
    .then(resp => {
      if (!resp.ok) {
        throw new Error("Can't fetch any players, please try again!");
      } else {
        return resp.json();
      }
    });
};

export const getAllCourses = (): Promise<Course[]> => {
  return fetch(endpoints.courses)
    .then(resp => {
      if (!resp.ok) {
        throw new Error("Can't fetch any courses, please try again!");
      } else {
        return resp.json();
      }
    });
};

export const getAllEvents = (playerId: number): Promise<Event[]> => {
  return fetch(`${endpoints.players}/${playerId}/events`)
    .then(resp => {
      if (!resp.ok) {
        throw new Error('Can\'t fetch any events, please try again!');
      } else {
        return resp.json();
      }
    });
};

export const postEvent = (
  courseId: string,
  date: string,
  teeTime: string,
  openSpots: string,
  numHoles: string,
  isPrivate: boolean,
  hostId: number,
  selectedFriends: number[]
): Promise<Event> | undefined => {
  if (!courseId || !teeTime) {
    return;
  }
  return fetch(`${endpoints.singleEvent}`, {
    method: 'POST',
    body: JSON.stringify({
      course_id: courseId,
      date: date,
      tee_time: teeTime,
      open_spots: openSpots,
      number_of_holes: numHoles,
      private: isPrivate,
      host_id: hostId,
      invitees: selectedFriends,
    }),
    headers: authHeaders()
  })
  .then(resp => {
    if (resp.ok) {
      return resp.json();
    } else {
      throw new Error();
    }
  });
};

export const postInviteAction = (playerId: number, eventId: number, inviteStatus: string): Promise<Event[]> => {
  return fetch(endpoints.playerEvent, {
    method: 'PATCH',
    body: JSON.stringify({
      player_id: playerId,
      event_id: eventId,
      invite_status: inviteStatus
    }),
    headers: authHeaders()
  })
  .then(() => getAllEvents(playerId));
};

export const getFriendsEvents = (playerId: number): Promise<Event[]> => {
  return fetch(`${endpoints.players}/${playerId}/friends-events`, {
    headers: authHeaders()
  })
    .then(resp => {
      if (!resp.ok) {
        throw new Error("Can't fetch friends events, please try again!");
      }
      return resp.json();
    });
};

export const joinEvent = (playerId: number, eventId: number): Promise<Event[]> => {
  return fetch(endpoints.joinEvent, {
    method: 'POST',
    body: JSON.stringify({
      player_id: playerId,
      event_id: eventId,
    }),
    headers: authHeaders()
  })
  .then(() => getAllEvents(playerId));
};

export const deleteEvent = (eventId: number, playerId: number): Promise<Event[]> => {
  return fetch(`${endpoints.singleEvent}/${eventId}`, {
    method: 'DELETE',
    headers: authHeaders()
  })
  .then(() => getAllEvents(playerId));
};

export interface FriendshipResponse {
  id: number;
  follower_id: number;
  followee_id: number;
  follower: Player;
  followee: Player;
}

export const postFriendship = (followerId: number, followeeId: number): Promise<FriendshipResponse> => {
  return fetch(`${endpoints.friendship}`, {
    method: 'POST',
    body: JSON.stringify({
      follower_id: followerId,
      followee_id: followeeId
    }),
    headers: authHeaders()
  })
  .then(resp => {
    if (resp.ok) {
      return resp.json();
    } else {
      throw new Error('Unable to update friendship, please try again!');
    }
  });
};

export const deleteFriendship = (followerId: number, followeeId: number): Promise<Response> => {
  return fetch(`${endpoints.friendship}`, {
    method: 'DELETE',
    body: JSON.stringify({
      follower_id: followerId,
      followee_id: followeeId
    }),
    headers: authHeaders()
  })
  .then(resp => {
    if (resp.ok) {
      return resp;
    } else {
      throw new Error('Unable to update friendship, please try again!');
    }
  });
};

export const createNewProfile = (
  name: string,
  phone: string,
  email: string,
  userName: string,
  password: string,
  passwordConfir: string
): Promise<Player> => {
  return fetch(`${endpoints.players}`, {
    method: 'POST',
    body: JSON.stringify({
      name: name,
      phone: phone,
      email: email,
      username: userName,
      password: password,
      password_confirmation: passwordConfir
    }),
    headers: { 'Content-Type': 'application/json' }
  })
  .then(resp => {
    if (resp.ok) {
      return resp.json();
    } else {
      throw new Error('Unable to create new profile, please try again!');
    }
  });
};

export const validateStandardLogin = (email: string, password: string): Promise<LoginResponse | undefined> => {
  return fetch(`${endpoints.sessions}`, {
    method: 'POST',
    body: JSON.stringify({
      email: email,
      password: password
    }),
    headers: { 'Content-Type': 'application/json' }
  })
  .then(resp => {
    if (resp.ok) {
      return resp.json();
    }
  });
};

// --- Newsfeed ---

export const getPosts = (limit = 50, offset = 0): Promise<Post[]> => {
  return fetch(`${endpoints.posts}?limit=${limit}&offset=${offset}`, {
    headers: authHeaders()
  })
    .then(resp => {
      if (!resp.ok) throw new Error("Can't fetch posts, please try again!");
      return resp.json();
    });
};

export const createPost = (playerId: number, body: string): Promise<Post> => {
  return fetch(endpoints.posts, {
    method: 'POST',
    body: JSON.stringify({ player_id: playerId, body }),
    headers: authHeaders()
  })
    .then(resp => {
      if (!resp.ok) throw new Error('Failed to create post');
      return resp.json();
    });
};

export const deletePost = (postId: number, playerId: number): Promise<void> => {
  return fetch(`${endpoints.posts}/${postId}`, {
    method: 'DELETE',
    body: JSON.stringify({ player_id: playerId }),
    headers: authHeaders()
  })
    .then(resp => {
      if (!resp.ok) throw new Error('Failed to delete post');
    });
};

export const toggleReaction = (postId: number, playerId: number, emoji: string): Promise<Reaction[]> => {
  return fetch(`${endpoints.posts}/${postId}/reactions`, {
    method: 'POST',
    body: JSON.stringify({ player_id: playerId, emoji }),
    headers: authHeaders()
  })
    .then(resp => {
      if (!resp.ok) throw new Error('Failed to toggle reaction');
      return resp.json();
    });
};

export const createReply = (postId: number, playerId: number, body: string): Promise<Reply> => {
  return fetch(`${endpoints.posts}/${postId}/replies`, {
    method: 'POST',
    body: JSON.stringify({ player_id: playerId, body }),
    headers: authHeaders()
  })
    .then(resp => {
      if (!resp.ok) throw new Error('Failed to create reply');
      return resp.json();
    });
};

export const deleteReply = (postId: number, replyId: number, playerId: number): Promise<void> => {
  return fetch(`${endpoints.posts}/${postId}/replies/${replyId}`, {
    method: 'DELETE',
    body: JSON.stringify({ player_id: playerId }),
    headers: authHeaders()
  })
    .then(resp => {
      if (!resp.ok) throw new Error('Failed to delete reply');
    });
};
