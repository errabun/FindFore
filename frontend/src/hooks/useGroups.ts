import { useCallback, useEffect, useState } from 'react';
import { groupAdapter } from '../adapters/api/groupAdapter';
import { ApiError } from '../adapters/api/httpClient';
import type { GroupInvitation, GroupMember, GroupSummary } from '../domain/group/types';

const DISCOVER_PAGE = 20;

function messageFrom(err: unknown, fallback: string) {
  return err instanceof ApiError ? err.message : fallback;
}

export function useGroups(hostPlayer: number, search = '') {
  const [mine, setMine] = useState<GroupSummary[]>([]);
  const [discover, setDiscover] = useState<GroupSummary[]>([]);
  const [discoverHasMore, setDiscoverHasMore] = useState(false);
  const [invitations, setInvitations] = useState<GroupInvitation[]>([]);
  const [loading, setLoading] = useState(Boolean(hostPlayer));
  const [error, setError] = useState('');

  const refreshMine = useCallback(() => {
    if (!hostPlayer) {
      setMine([]);
      setInvitations([]);
      return Promise.resolve();
    }
    return Promise.all([groupAdapter.listMine(), groupAdapter.listInvitations()])
      .then(([m, inv]) => {
        setMine(m);
        setInvitations(inv);
      })
      .catch((err) => setError(messageFrom(err, 'Could not load groups')));
  }, [hostPlayer]);

  const refreshDiscover = useCallback(() => {
    if (!hostPlayer) {
      setDiscover([]);
      setDiscoverHasMore(false);
      return Promise.resolve();
    }
    return groupAdapter
      .listDiscover(search, 0)
      .then((d) => {
        setDiscover(d);
        setDiscoverHasMore(d.length === DISCOVER_PAGE);
      })
      .catch((err) => setError(messageFrom(err, 'Could not load groups')));
  }, [hostPlayer, search]);

  const refresh = useCallback(() => {
    if (!hostPlayer) {
      setMine([]);
      setDiscover([]);
      setInvitations([]);
      setDiscoverHasMore(false);
      return Promise.resolve();
    }
    setLoading(true);
    setError('');
    return Promise.all([refreshMine(), refreshDiscover()]).finally(() => setLoading(false));
  }, [hostPlayer, refreshMine, refreshDiscover]);

  const loadMoreDiscover = useCallback(() => {
    if (!hostPlayer) return Promise.resolve();
    return groupAdapter
      .listDiscover(search, discover.length)
      .then((d) => {
        setDiscover((prev) => [...prev, ...d]);
        setDiscoverHasMore(d.length === DISCOVER_PAGE);
      })
      .catch((err) => setError(messageFrom(err, 'Could not load more groups')));
  }, [hostPlayer, search, discover.length]);

  useEffect(() => {
    if (!hostPlayer) return;
    setLoading(true);
    setError('');
    refreshMine().finally(() => setLoading(false));
  }, [hostPlayer, refreshMine]);

  useEffect(() => {
    refreshDiscover();
  }, [refreshDiscover]);

  return {
    mine,
    discover,
    discoverHasMore,
    invitations,
    loading,
    error,
    refresh,
    loadMoreDiscover,
  };
}

export function useGroup(id: number, hostPlayer: number) {
  const [group, setGroup] = useState<GroupSummary | null>(null);
  const [members, setMembers] = useState<GroupMember[]>([]);
  const [joinRequests, setJoinRequests] = useState<GroupMember[]>([]);
  const [invitations, setInvitations] = useState<GroupInvitation[]>([]);
  const [loading, setLoading] = useState(Boolean(hostPlayer && id));
  const [error, setError] = useState('');
  const [notFound, setNotFound] = useState(false);

  const refresh = useCallback(() => {
    if (!hostPlayer || !id) {
      setGroup(null);
      return Promise.resolve();
    }
    setLoading(true);
    setError('');
    setNotFound(false);
    return groupAdapter
      .get(id)
      .then(async (g) => {
        setGroup(g);
        const canManage =
          g.viewer_membership?.status === 'active' &&
          (g.viewer_membership.role === 'owner' || g.viewer_membership.role === 'admin');
        if (g.viewer_membership?.status === 'active') {
          const [mems, reqs, invs] = await Promise.all([
            groupAdapter.listMembers(id),
            canManage ? groupAdapter.listJoinRequests(id) : Promise.resolve([] as GroupMember[]),
            canManage ? groupAdapter.listGroupInvitations(id) : Promise.resolve([] as GroupInvitation[]),
          ]);
          setMembers(mems);
          setJoinRequests(reqs);
          setInvitations(invs);
        } else {
          setMembers([]);
          setJoinRequests([]);
          setInvitations([]);
        }
      })
      .catch((err) => {
        setGroup(null);
        setMembers([]);
        setJoinRequests([]);
        setInvitations([]);
        if (err instanceof ApiError && err.status === 404) {
          setNotFound(true);
          return;
        }
        setError(messageFrom(err, 'Could not load this group'));
      })
      .finally(() => setLoading(false));
  }, [hostPlayer, id]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { group, members, joinRequests, invitations, loading, error, notFound, refresh };
}
