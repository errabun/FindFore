import type { Player } from './types';
import type { Friend } from '../social/types';

export function buildFriendList(allPlayers: Player[], currentPlayer: Player): Friend[] {
  return allPlayers
    .filter((p) => currentPlayer.friends.includes(p.id))
    .map((f) => ({ name: f.name, id: f.id }));
}
