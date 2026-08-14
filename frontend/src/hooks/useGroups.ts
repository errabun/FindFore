import { useCallback, useEffect, useState } from 'react';
import { groupAdapter } from '../adapters/api/groupAdapter';
import type { GroupInvitation, GroupMember, GroupSummary } from '../domain/group/types';

export function useGroups(hostPlayer: number) {
  const [mine, setMine] = useState<GroupSummary[]>([]);
  const [discover, setDiscover] = useState<GroupSummary[]>([]);
  const [invitations, setInvitations] = useState<GroupInvitation[]>([]);
  const [loading, setLoading] = useState(false);

  const refresh = useCallback(() => {
    if (!hostPlayer) {
      setMine([]);
      setDiscover([]);
      setInvitations([]);
      return Promise.resolve();
    }
    setLoading(true);
    return Promise.all([
      groupAdapter.listMine(),
      groupAdapter.listDiscover(),
      groupAdapter.listInvitations(),
    ])
      .then(([m, d, inv]) => {
        setMine(m);
        setDiscover(d);
        setInvitations(inv);
      })
      .catch(() => undefined)
      .finally(() => setLoading(false));
  }, [hostPlayer]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { mine, discover, invitations, loading, refresh, setDiscover };
}

export function useGroup(id: number, hostPlayer: number) {
  const [group, setGroup] = useState<GroupSummary | null>(null);
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [joinRequests, setJoinRequests] = useState<GroupMember[]>([]);
  const [loading, setLoading] = useState(false);

  const refresh = useCallback(() => {
    if (!hostPlayer || !id) {
      setGroup(null);
      return Promise.resolve();
    }
    setLoading(true);
    return groupAdapter
      .get(id)
      .then(async (g) => {
        setGroup(g);
        if (g.viewer_membership?.status === 'active') {
          const [mems, reqs] = await Promise.all([
            groupAdapter.listMembers(id),
            g.viewer_membership.role === 'owner' || g.viewer_membership.role === 'admin'
              ? groupAdapter.listJoinRequests(id)
              : Promise.resolve([] as GroupMember[]),
          ]);
          setMembers(mems);
          setJoinRequests(reqs);
        } else {
          setMembers([]);
          setJoinRequests([]);
        }
      })
      .catch(() => setGroup(null))
      .finally(() => setLoading(false));
  }, [hostPlayer, id]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { group, members, joinRequests, loading, refresh, setGroup };
}
