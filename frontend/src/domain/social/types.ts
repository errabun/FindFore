export interface Friend {
  id: number;
  name: string;
}

export interface Reaction {
  id: number;
  player_id: number;
  player_name: string;
  emoji: string;
}

export interface Reply {
  id: number;
  player_id: number;
  player_name: string;
  body: string;
  created_at: string;
}

export interface Post {
  id: number;
  player_id: number;
  player_name: string;
  body: string;
  created_at: string;
  reactions: Reaction[];
  replies: Reply[];
}
