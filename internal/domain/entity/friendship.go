package entity

type FriendshipStatus int32

const (
	FriendshipStatusPending  FriendshipStatus = 0
	FriendshipStatusAccepted FriendshipStatus = 1
)

func (s FriendshipStatus) String() string {
	switch s {
	case FriendshipStatusPending:
		return "pending"
	case FriendshipStatusAccepted:
		return "accepted"
	default:
		return "unknown"
	}
}

type Friendship struct {
	ID          int64
	RequesterID int32
	AddresseeID int32
	Status      FriendshipStatus
}

// OtherPlayerID returns the player on this friendship who is not viewerID.
func (f Friendship) OtherPlayerID(viewerID int32) int32 {
	if f.RequesterID == viewerID {
		return f.AddresseeID
	}
	return f.RequesterID
}
