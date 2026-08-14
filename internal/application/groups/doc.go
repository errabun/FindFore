// Package groups owns persistent golfer groups, membership, and invitations.
// Owners can update, transfer ownership, or delete a group; they cannot leave
// until ownership moves. The admin role is enforced (managers can invite and
// remove members, but cannot remove another admin or owner, transfer, or delete).
// Promote-to-admin remains deferred.
package groups
